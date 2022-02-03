package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/google/uuid"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func main() {
	google_conf := &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  "http://localhost:3000/auth/google/callback",
		Scopes:       []string{"email", "profile"},
		Endpoint:     google.Endpoint,
	}

	GET("/", func(w Response, r Request) Output {
		user := current_user(r)
		if user != nil {
			return Redirect(fmt.Sprintf("/users/%s", user.Slug))
		} else {
			return Render("layout", "index", map[string]interface{}{
				"meta": map[string]string{},
			})
		}
	})

	GET("/privacy", func(w Response, r Request) Output {
		return Render("layout", "privacy", map[string]interface{}{
			"meta":         map[string]string{},
			"current_user": current_user(r),
		})
	})

	POST("/auth/google", func(w Response, r Request) Output {
		return Redirect(google_conf.AuthCodeURL("state"))
	})

	GET("/auth/google/callback", func(w Response, r Request) Output {
		tok, err := google_conf.Exchange(oauth2.NoContext, r.URL.Query().Get("code"))
		if err != nil {
			return BadRequest
		}

		client := google_conf.Client(oauth2.NoContext, tok)
		resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
		if err != nil {
			return Unauthorized
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return InternalServerError
		}

		user := struct {
			Name    string `json:"name"`
			Email   string `json:"email"`
			Picture string `json:"picture"`
		}{}
		err = json.Unmarshal(body, &user)
		if err != nil {
			return InternalServerError
		}

		u, err := queries.Signup(r.Context(), SignupParams{
			Name:  sql.NullString{String: user.Name, Valid: true},
			Image: sql.NullString{String: user.Picture, Valid: true},
			Slug:  uuid.New().String(),
			Email: sql.NullString{String: user.Email, Valid: true},
		})
		if err != nil {
			return InternalServerError
		}

		s := SESSION(r)
		s.Values["current_user"] = u
		if err = s.Save(r, w); err != nil {
			return InternalServerError
		}

		return Redirect("/")
	})

	GET("/logout", func(w Response, r Request) Output {
		s := SESSION(r)
		s.Values = map[interface{}]interface{}{}
		s.Save(r, w)
		return Redirect("/")
	})

	GET("/users/{user}", func(w Response, r Request) Output {
		vars := mux.Vars(r)

		user, err := queries.UserBySlug(r.Context(), vars["user"])
		if err != nil {
			return NotFound
		}

		data := map[string]interface{}{
			"current_user": current_user(r),
			"user":         user,
		}

		unshelved_books, err := queries.UserUnshelvedBooks(r.Context(), user.ID)
		if len(unshelved_books) > 0 {
			data["unshelved_books"] = unshelved_books
		}

		data["shelves"], err = queries.Shelves(r.Context(), user.ID)

		return Render("layout", "users/show", data)
	})

	GET("/users/{user}/edit", func(w Response, r Request) Output {
		actor := current_user(r)
		vars := mux.Vars(r)

		user, err := queries.UserBySlug(r.Context(), vars["user"])
		if err != nil {
			return NotFound
		}

		if !can(actor, "edit", user) {
			return Unauthorized
		}

		return Render("layout", "users/edit", map[string]interface{}{
			"current_user": actor,
			"user":         user,
		})
	}, loggedinMiddleware)

	POST("/users/{user}", func(w Response, r Request) Output {
		actor := current_user(r)
		vars := mux.Vars(r)

		user, err := queries.UserBySlug(r.Context(), vars["user"])
		if err != nil {
			return NotFound
		}

		if !can(actor, "edit", user) {
			return Unauthorized
		}

		return Redirect(fmt.Sprintf("/users/%s", user.Slug))
	}, loggedinMiddleware)

	GET("/users/{user}/books/new", func(w Response, r Request) Output {
		actor := current_user(r)
		vars := mux.Vars(r)

		user, err := queries.UserBySlug(r.Context(), vars["user"])
		if err != nil {
			return NotFound
		}

		if !can(actor, "create_book", user) {
			return Unauthorized
		}

		return Render("layout", "books/new", map[string]interface{}{
			"current_user": actor,
			"user":         user,
			"errors":       ValidationErrors{},
			"csrf":         csrf.TemplateField(r),
		})
	}, loggedinMiddleware)

	POST("/users/{user}/books", func(w Response, r Request) Output {
		actor := current_user(r)
		vars := mux.Vars(r)

		user, err := queries.UserBySlug(r.Context(), vars["user"])
		if err != nil {
			return NotFound
		}

		r.ParseForm()
		form := r.Form
		params := NewBookParams{
			Title:       form.Get("title"),
			Isbn:        form.Get("isbn"),
			Author:      form.Get("author"),
			Subtitle:    form.Get("subtitle"),
			Description: form.Get("description"),
			Publisher:   form.Get("publisher"),
			PageCount:   atoi32(form.Get("page_count")),
			GoogleBooksID: sql.NullString{
				String: form.Get("google_books_id"),
				Valid:  len(form.Get("google_books_id")) > 0,
			},
			UserID: user.ID,
		}
		errors := params.Validate()
		if len(errors) != 0 {
			return Render("layout", "books/new", map[string]interface{}{
				"book":         params,
				"current_user": actor,
				"user":         user,
				"errors":       errors,
				"csrf":         csrf.TemplateField(r),
			})
		}

		book, err := queries.NewBook(r.Context(), params)
		if err != nil {
			return InternalServerError
		}

		return Redirect(fmt.Sprintf("/users/%s/books/%s", user.Slug, book.Isbn))
	})

	GET("/users/{user}/books/{isbn}", func(w Response, r Request) Output {
		vars := mux.Vars(r)

		user, err := queries.UserBySlug(r.Context(), vars["user"])
		if err != nil {
			return InternalServerError
		}

		book, err := queries.BookByIsbnAndUser(r.Context(), BookByIsbnAndUserParams{
			UserID: user.ID,
			Isbn:   vars["isbn"],
		})
		if err != nil {
			return NotFound
		}

		highlights, err := queries.Highlights(r.Context(), book.ID)
		if err != nil {
			return InternalServerError
		}

		return Render("layout", "books/show", map[string]interface{}{
			"current_user": current_user(r),
			"user":         user,
			"book":         book,
			"highlights":   highlights,
		})
	})

	GET("/users/{user}/books/{isbn}/edit", func(w Response, r Request) Output {
		actor := current_user(r)
		vars := mux.Vars(r)

		user, err := queries.UserBySlug(r.Context(), vars["user"])
		if err != nil {
			return NotFound
		}

		book, err := queries.BookByIsbnAndUser(r.Context(), BookByIsbnAndUserParams{
			UserID: user.ID,
			Isbn:   vars["isbn"],
		})
		if err != nil {
			return NotFound
		}

		return Render("layout", "books/new", map[string]interface{}{
			"current_user": actor,
			"user":         user,
			"book":         book,
			"csrf":         csrf.TemplateField(r),
			"errors":       ValidationErrors{},
		})
	}, loggedinMiddleware)

	POST("/users/{user}/books/{isbn}", func(w Response, r Request) Output {
		actor := current_user(r)
		vars := mux.Vars(r)

		user, err := queries.UserBySlug(r.Context(), vars["user"])
		if err != nil {
			return NotFound
		}

		book, err := queries.BookByIsbnAndUser(r.Context(), BookByIsbnAndUserParams{
			UserID: user.ID,
			Isbn:   vars["isbn"],
		})
		if err != nil {
			return NotFound
		}

		if !can(actor, "edit", book) {
			return Unauthorized
		}

		return Redirect(fmt.Sprintf("/users/%s/books/%s", user.Slug, vars["isbn"]))
	}, loggedinMiddleware)

	DELETE("/users/{user}/books/{isbn}", func(w Response, r Request) Output {
		actor := current_user(r)
		vars := mux.Vars(r)

		user, err := queries.UserBySlug(r.Context(), vars["user"])
		if err != nil {
			return NotFound
		}

		book, err := queries.BookByIsbnAndUser(r.Context(), BookByIsbnAndUserParams{
			UserID: user.ID,
			Isbn:   vars["isbn"],
		})
		if err != nil {
			return NotFound
		}

		if !can(actor, "delete", book) {
			return Unauthorized
		}

		return Redirect(fmt.Sprintf("/users/%s", user.Slug))
	}, loggedinMiddleware)

	POST("/users/{user}/books/{isbn}/shelf", func(w Response, r Request) Output {
		actor := current_user(r)
		vars := mux.Vars(r)

		user, err := queries.UserBySlug(r.Context(), vars["user"])
		if err != nil {
			return NotFound
		}

		book, err := queries.BookByIsbnAndUser(r.Context(), BookByIsbnAndUserParams{
			UserID: user.ID,
			Isbn:   vars["isbn"],
		})
		if err != nil {
			return NotFound
		}

		if !can(actor, "edit", book) {
			return Unauthorized
		}

		return Redirect(fmt.Sprintf("/users/%s/books/%s", user.Slug, vars["isbn"]))
	}, loggedinMiddleware)

	GET("/users/{user}/books/{isbn}/image", func(w Response, r Request) Output {
		actor := current_user(r)
		vars := mux.Vars(r)

		user, err := queries.UserBySlug(r.Context(), vars["user"])
		if err != nil {
			return NotFound
		}

		book, err := queries.BookByIsbnAndUser(r.Context(), BookByIsbnAndUserParams{
			UserID: user.ID,
			Isbn:   vars["isbn"],
		})
		if err != nil {
			return NotFound
		}

		if !can(actor, "edit", book) {
			return Unauthorized
		}

		return Render("layout", "books/image", map[string]interface{}{
			"current_user": actor,
			"user":         user,
		})
	}, loggedinMiddleware)

	POST("/users/{user}/books/{isbn}/image", func(w Response, r Request) Output {
		actor := current_user(r)
		vars := mux.Vars(r)

		user, err := queries.UserBySlug(r.Context(), vars["user"])
		if err != nil {
			return NotFound
		}

		book, err := queries.BookByIsbnAndUser(r.Context(), BookByIsbnAndUserParams{
			UserID: user.ID,
			Isbn:   vars["isbn"],
		})
		if err != nil {
			return NotFound
		}

		if !can(actor, "edit", book) {
			return Unauthorized
		}

		return Redirect(fmt.Sprintf("/users/%s/books/%s", user.Slug, vars["isbn"]))
	}, loggedinMiddleware)

	GET("/users/{user}/shelves/new", func(w Response, r Request) Output {
		actor := current_user(r)
		vars := mux.Vars(r)

		user, err := queries.UserBySlug(r.Context(), vars["user"])
		if err != nil {
			return NotFound
		}

		return Render("layout", "shelves/new", map[string]interface{}{
			"current_user": actor,
			"user":         user,
			"csrf":         csrf.TemplateField(r),
		})
	}, loggedinMiddleware)

	GET("/users/{user}/shelves", func(w Response, r Request) Output {
		actor := current_user(r)
		vars := mux.Vars(r)

		user, err := queries.UserBySlug(r.Context(), vars["user"])
		if err != nil {
			return NotFound
		}

		return Render("layout", "shelves/index", map[string]interface{}{
			"current_user": actor,
			"user":         user,
		})
	}, loggedinMiddleware)

	POST("/users/{user}/shelves", func(w Response, r Request) Output {
		actor := current_user(r)
		vars := mux.Vars(r)

		user, err := queries.UserBySlug(r.Context(), vars["user"])
		if err != nil {
			return NotFound
		}

		if !can(actor, "create_shelf", user) {
			return Unauthorized
		}

		return Redirect(fmt.Sprintf("/users/%s/shelves", user.Slug))
	}, loggedinMiddleware)

	GET("/users/{user}/shelves/{shelf}/edit", func(w Response, r Request) Output {
		actor := current_user(r)
		vars := mux.Vars(r)

		user, err := queries.UserBySlug(r.Context(), vars["user"])
		if err != nil {
			return NotFound
		}

		return Render("layout", "shelves/edit", map[string]interface{}{
			"current_user": actor,
			"user":         user,
			"csrf":         csrf.TemplateField(r),
		})
	}, loggedinMiddleware)

	POST("/users/{user}/shelves/{shelf}", func(w Response, r Request) Output {
		actor := current_user(r)
		vars := mux.Vars(r)

		user, err := queries.UserBySlug(r.Context(), vars["user"])
		if err != nil {
			return NotFound
		}

		shelf, err := queries.ShelfByIdAndUser(r.Context(), ShelfByIdAndUserParams{
			UserID: user.ID,
			ID:     atoi64(vars["shelf"]),
		})
		if err != nil {
			return NotFound
		}

		if !can(actor, "edit", shelf) {
			return Unauthorized
		}

		return Redirect(fmt.Sprintf("/users/%s/shelves", user.Slug))
	}, loggedinMiddleware)

	POST("/users/{user}/shelves/{shelf}/up", func(w Response, r Request) Output {
		actor := current_user(r)
		vars := mux.Vars(r)

		user, err := queries.UserBySlug(r.Context(), vars["user"])
		if err != nil {
			return NotFound
		}

		shelf, err := queries.ShelfByIdAndUser(r.Context(), ShelfByIdAndUserParams{
			UserID: user.ID,
			ID:     atoi64(vars["shelf"]),
		})
		if err != nil {
			return NotFound
		}

		if !can(actor, "edit", shelf) {
			return Unauthorized
		}

		return Redirect(fmt.Sprintf("/users/%s/shelves", user.Slug))
	}, loggedinMiddleware)

	POST("/users/{user}/shelves/{shelf}/down", func(w Response, r Request) Output {
		actor := current_user(r)
		vars := mux.Vars(r)

		user, err := queries.UserBySlug(r.Context(), vars["user"])
		if err != nil {
			return NotFound
		}

		shelf, err := queries.ShelfByIdAndUser(r.Context(), ShelfByIdAndUserParams{
			UserID: user.ID,
			ID:     atoi64(vars["shelf"]),
		})
		if err != nil {
			return NotFound
		}

		if !can(actor, "edit", shelf) {
			return Unauthorized
		}

		return Redirect(fmt.Sprintf("/users/%s/shelves", user.Slug))
	}, loggedinMiddleware)

	DELETE("/users/{user}/shelves/{shelf}", func(w Response, r Request) Output {
		actor := current_user(r)
		vars := mux.Vars(r)

		user, err := queries.UserBySlug(r.Context(), vars["user"])
		if err != nil {
			return NotFound
		}

		shelf, err := queries.ShelfByIdAndUser(r.Context(), ShelfByIdAndUserParams{
			UserID: user.ID,
			ID:     atoi64(vars["shelf"]),
		})
		if err != nil {
			return NotFound
		}

		if !can(actor, "delete", shelf) {
			return Unauthorized
		}

		return Redirect(fmt.Sprintf("/users/%s/shelves", user.Slug))
	}, loggedinMiddleware)

	GET("/users/{user}/books/{isbn}/highlights/new", func(w Response, r Request) Output {
		actor := current_user(r)
		vars := mux.Vars(r)

		user, err := queries.UserBySlug(r.Context(), vars["user"])
		if err != nil {
			return NotFound
		}

		book, err := queries.BookByIsbnAndUser(r.Context(), BookByIsbnAndUserParams{
			UserID: user.ID,
			Isbn:   vars["isbn"],
		})
		if err != nil {
			return NotFound
		}

		if !can(actor, "create_highlight", book) {
			return Unauthorized
		}

		return Render("layout", "highlights/new", map[string]interface{}{
			"current_user": actor,
			"user":         user,
			"csrf":         csrf.TemplateField(r),
		})
	}, loggedinMiddleware)

	POST("/users/{user}/books/{isbn}/highlights", func(w Response, r Request) Output {
		actor := current_user(r)
		vars := mux.Vars(r)

		user, err := queries.UserBySlug(r.Context(), vars["user"])
		if err != nil {
			return NotFound
		}

		book, err := queries.BookByIsbnAndUser(r.Context(), BookByIsbnAndUserParams{
			UserID: user.ID,
			Isbn:   vars["isbn"],
		})
		if err != nil {
			return NotFound
		}

		if !can(actor, "create_highlight", book) {
			return Unauthorized
		}

		return Redirect(fmt.Sprintf("/users/%s/books/%s", user.Slug, vars["isbn"]))
	}, loggedinMiddleware)

	GET("/users/{user}/books/{isbn}/highlights/{id}/edit", func(w Response, r Request) Output {
		actor := current_user(r)
		vars := mux.Vars(r)

		user, err := queries.UserBySlug(r.Context(), vars["user"])
		if err != nil {
			return NotFound
		}

		book, err := queries.BookByIsbnAndUser(r.Context(), BookByIsbnAndUserParams{
			UserID: user.ID,
			Isbn:   vars["isbn"],
		})
		if err != nil {
			return NotFound
		}

		highlight, err := queries.HighlightByIDAndBook(r.Context(), HighlightByIDAndBookParams{
			ID:     atoi64(vars["id"]),
			BookID: book.ID,
		})
		if err != nil {
			return NotFound
		}

		if !can(actor, "edit_highlight", book) {
			return Unauthorized
		}

		return Render("layout", "highlights/edit", map[string]interface{}{
			"current_user": actor,
			"user":         user,
			"highlight":    highlight,
			"csrf":         csrf.TemplateField(r),
		})
	}, loggedinMiddleware)

	POST("/users/{user}/books/{isbn}/highlights/{id}", func(w Response, r Request) Output {
		actor := current_user(r)
		vars := mux.Vars(r)

		user, err := queries.UserBySlug(r.Context(), vars["user"])
		if err != nil {
			return NotFound
		}

		book, err := queries.BookByIsbnAndUser(r.Context(), BookByIsbnAndUserParams{
			UserID: user.ID,
			Isbn:   vars["isbn"],
		})
		if err != nil {
			return NotFound
		}

		_, err = queries.HighlightByIDAndBook(r.Context(), HighlightByIDAndBookParams{
			ID:     atoi64(vars["id"]),
			BookID: book.ID,
		})
		if err != nil {
			return NotFound
		}

		if !can(actor, "edit_highlight", book) {
			return Unauthorized
		}

		return Redirect(fmt.Sprintf("/users/%s/books/%s", user.Slug, vars["isbn"]))
	}, loggedinMiddleware)

	DELETE("/users/{user}/books/{isbn}/highlights/{id}", func(w Response, r Request) Output {
		actor := current_user(r)
		vars := mux.Vars(r)

		user, err := queries.UserBySlug(r.Context(), vars["user"])
		if err != nil {
			return NotFound
		}

		book, err := queries.BookByIsbnAndUser(r.Context(), BookByIsbnAndUserParams{
			UserID: user.ID,
			Isbn:   vars["isbn"],
		})
		if err != nil {
			return NotFound
		}

		_, err = queries.HighlightByIDAndBook(r.Context(), HighlightByIDAndBookParams{
			ID:     atoi64(vars["id"]),
			BookID: book.ID,
		})
		if err != nil {
			return NotFound
		}

		if !can(actor, "delete_highlight", book) {
			return Unauthorized
		}

		return Redirect(fmt.Sprintf("/users/%s/books/%s", user.Slug, vars["isbn"]))
	}, loggedinMiddleware)

	GET("/users/{user}/books/{isbn}/highlights/{id}/image", func(w Response, r Request) Output {
		actor := current_user(r)
		vars := mux.Vars(r)

		user, err := queries.UserBySlug(r.Context(), vars["user"])
		if err != nil {
			return NotFound
		}

		book, err := queries.BookByIsbnAndUser(r.Context(), BookByIsbnAndUserParams{
			UserID: user.ID,
			Isbn:   vars["isbn"],
		})
		if err != nil {
			return NotFound
		}

		highlight, err := queries.HighlightByIDAndBook(r.Context(), HighlightByIDAndBookParams{
			ID:     atoi64(vars["id"]),
			BookID: book.ID,
		})
		if err != nil {
			return NotFound
		}

		if !can(actor, "edit_highlight", book) {
			return Unauthorized
		}

		return Render("layout", "highlights/image", map[string]interface{}{
			"current_user": actor,
			"user":         user,
			"highlight":    highlight,
		})
	}, loggedinMiddleware)

	POST("/users/{user}/books/{isbn}/highlights/{id}/image", func(w Response, r Request) Output {
		actor := current_user(r)
		vars := mux.Vars(r)

		user, err := queries.UserBySlug(r.Context(), vars["user"])
		if err != nil {
			return NotFound
		}

		book, err := queries.BookByIsbnAndUser(r.Context(), BookByIsbnAndUserParams{
			UserID: user.ID,
			Isbn:   vars["isbn"],
		})
		if err != nil {
			return NotFound
		}

		_, err = queries.HighlightByIDAndBook(r.Context(), HighlightByIDAndBookParams{
			ID:     atoi64(vars["id"]),
			BookID: book.ID,
		})
		if err != nil {
			return NotFound
		}

		if !can(actor, "edit_highlight", book) {
			return Unauthorized
		}

		return Redirect(fmt.Sprintf("/users/%s/books/%s", user.Slug, vars["isbn"]))
	}, loggedinMiddleware)

	Helpers()
	Start()
}
