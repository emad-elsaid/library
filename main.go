package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/google/uuid"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	BOOK_COVER_PATH      = "public/books/image"
	HIGHLIGHT_IMAGE_PATH = "public/highlights/image"
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
			return InternalServerError(err)
		}

		user := struct {
			Name    string `json:"name"`
			Email   string `json:"email"`
			Picture string `json:"picture"`
		}{}
		err = json.Unmarshal(body, &user)
		if err != nil {
			return InternalServerError(err)
		}

		u, err := queries.Signup(r.Context(), SignupParams{
			Name:  NullString(user.Name),
			Image: NullString(user.Picture),
			Slug:  uuid.New().String(),
			Email: NullString(user.Email),
		})
		if err != nil {
			return InternalServerError(err)
		}

		s := SESSION(r)
		s.Values["current_user"] = u
		if err = s.Save(r, w); err != nil {
			return InternalServerError(err)
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
			"errors":       ValidationErrors{},
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

		// TODO find a way to remove this duplication
		params := UpdateUserParams{
			Description:        NullString(r.FormValue("description")),
			AmazonAssociatesID: NullString(r.FormValue("amazon_associates_id")),
			Facebook:           NullString(r.FormValue("facebook")),
			Twitter:            NullString(r.FormValue("twitter")),
			Linkedin:           NullString(r.FormValue("linkedin")),
			Instagram:          NullString(r.FormValue("instagram")),
			Phone:              NullString(r.FormValue("phone")),
			Whatsapp:           NullString(r.FormValue("whatsapp")),
			Telegram:           NullString(r.FormValue("telegram")),
			ID:                 user.ID,
		}
		errors := params.Validate()
		if len(errors) != 0 {
			user.Description = params.Description
			user.AmazonAssociatesID = params.AmazonAssociatesID
			user.Facebook = params.Facebook
			user.Twitter = params.Twitter
			user.Linkedin = params.Linkedin
			user.Instagram = params.Instagram
			user.Phone = params.Phone
			user.Whatsapp = params.Whatsapp
			user.Telegram = params.Telegram
			return Render("layout", "users/edit", map[string]interface{}{
				"current_user": actor,
				"user":         user,
				"errors":       errors,
			})
		}

		err = queries.UpdateUser(r.Context(), params)
		if err != nil {
			return InternalServerError(err)
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

		if !can(actor, "create_book", user) {
			return Unauthorized
		}

		r.ParseMultipartForm(MB * 10)
		params := NewBookParams{
			Title:         r.FormValue("title"),
			Isbn:          r.FormValue("isbn"),
			Author:        r.FormValue("author"),
			Subtitle:      r.FormValue("subtitle"),
			Description:   r.FormValue("description"),
			Publisher:     r.FormValue("publisher"),
			PageCount:     atoi32(r.FormValue("page_count")),
			GoogleBooksID: NullString(r.FormValue("google_books_id")),
			UserID:        user.ID,
		}
		errors := params.Validate()

		file, _, _ := r.FormFile("image")
		if file != nil {
			ValidateImage(file, "image", "Image", errors, 600, 600)
			file.Seek(0, os.SEEK_SET)
		}

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
			return InternalServerError(err)
		}

		if file != nil {
			name, err := UploadImage(file, BOOK_COVER_PATH, 432, 576)
			if err != nil {
				return InternalServerError(err)
			}

			err = queries.UpdateBookImage(r.Context(), UpdateBookImageParams{
				Image: NullString(name),
				ID:    book.ID,
			})
			if err != nil {
				return InternalServerError(err)
			}
		}

		return Redirect(fmt.Sprintf("/users/%s/books/%s", user.Slug, book.Isbn))
	})

	GET("/users/{user}/books/{isbn}", func(w Response, r Request) Output {
		vars := mux.Vars(r)

		user, err := queries.UserBySlug(r.Context(), vars["user"])
		if err != nil {
			return InternalServerError(err)
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
			return InternalServerError(err)
		}

		shelves, err := queries.Shelves(r.Context(), user.ID)
		if err != nil {
			return InternalServerError(err)
		}

		return Render("layout", "books/show", map[string]interface{}{
			"current_user": current_user(r),
			"user":         user,
			"book":         book,
			"shelves":      shelves,
			"highlights":   highlights,
			"csrf":         csrf.TemplateField(r),
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

		if !can(actor, "edit", book) {
			return Unauthorized
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

		r.ParseMultipartForm(MB * 10)

		params := UpdateBookParams{
			Title:       r.FormValue("title"),
			Author:      r.FormValue("author"),
			Subtitle:    r.FormValue("subtitle"),
			Description: r.FormValue("description"),
			Publisher:   r.FormValue("publisher"),
			PageCount:   atoi32(r.FormValue("page_count")),
			ID:          book.ID,
		}

		errors := params.Validate()
		file, _, _ := r.FormFile("image")
		if file != nil {
			ValidateImage(file, "image", "Image", errors, 600, 600)
			file.Seek(0, os.SEEK_SET)
		}

		if len(errors) > 0 {
			book.Title = params.Title
			book.Author = params.Author
			book.Subtitle = params.Subtitle
			book.Description = params.Description
			book.Publisher = params.Publisher
			book.PageCount = params.PageCount
			return Render("layout", "books/new", map[string]interface{}{
				"current_user": actor,
				"user":         user,
				"book":         book,
				"csrf":         csrf.TemplateField(r),
				"errors":       errors,
			})
		}

		err = queries.UpdateBook(r.Context(), params)
		if err != nil {
			return InternalServerError(err)
		}

		if file != nil {
			name, err := UploadImage(file, BOOK_COVER_PATH, 432, 576)
			if err != nil {
				return InternalServerError(err)
			}

			oldname := path.Join(BOOK_COVER_PATH, book.Image.String)
			if book.Image.Valid && len(book.Image.String) > 0 { // if image is set
				if _, err = os.Stat(oldname); err == nil { // and it exists
					os.Remove(oldname) // delete it
				}
			}

			err = queries.UpdateBookImage(r.Context(), UpdateBookImageParams{
				Image: NullString(name),
				ID:    book.ID,
			})
			if err != nil {
				return InternalServerError(err)
			}
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

		images, err := queries.HighlightsWithImages(r.Context(), book.ID)
		if err != nil {
			return InternalServerError(err)
		}
		for _, v := range images {
			os.Remove(path.Join(HIGHLIGHT_IMAGE_PATH, v.String))
		}

		if book.Image.Valid && len(book.Image.String) > 0 {
			os.Remove(path.Join(BOOK_COVER_PATH, book.Image.String))
		}

		err = queries.DeleteBook(r.Context(), book.ID)
		if err != nil {
			return InternalServerError(err)
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

		shelf, err := queries.ShelfByIdAndUser(r.Context(), ShelfByIdAndUserParams{
			UserID: user.ID,
			ID:     atoi64(r.FormValue("shelf_id")),
		})
		if err == nil && !can(actor, "edit", shelf) {
			return Unauthorized
		}

		err = queries.MoveBookToShelf(r.Context(), MoveBookToShelfParams{
			ShelfID: sql.NullInt64{Int64: shelf.ID, Valid: err == nil},
			ID:      book.ID,
		})
		if err != nil {
			return InternalServerError(err)
		}

		return Redirect(fmt.Sprintf("/users/%s/books/%s", user.Slug, vars["isbn"]))
	}, loggedinMiddleware)

	GET("/users/{user}/shelves", func(w Response, r Request) Output {
		actor := current_user(r)
		vars := mux.Vars(r)

		user, err := queries.UserBySlug(r.Context(), vars["user"])
		if err != nil {
			return NotFound
		}

		if !can(actor, "show_shelves", user) {
			return Unauthorized
		}

		shelves, err := queries.Shelves(r.Context(), user.ID)
		if err != nil {
			return InternalServerError(err)
		}

		return Render("layout", "shelves/index", map[string]interface{}{
			"current_user": actor,
			"user":         user,
			"shelves":      shelves,
			"errors":       ValidationErrors{},
			"csrf":         csrf.TemplateField(r),
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

		params := NewShelfParams{
			Name:   r.FormValue("name"),
			UserID: user.ID,
		}

		errors := params.Validate()
		if len(errors) > 0 {
			shelves, err := queries.Shelves(r.Context(), user.ID)
			if err != nil {
				return InternalServerError(err)
			}

			return Render("layout", "shelves/index", map[string]interface{}{
				"current_user": actor,
				"user":         user,
				"shelves":      shelves,
				"shelf":        params,
				"csrf":         csrf.TemplateField(r),
				"errors":       errors,
			})
		}

		err = queries.NewShelf(r.Context(), params)
		if err != nil {
			return InternalServerError(err)
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

		shelf, err := queries.ShelfByIdAndUser(r.Context(), ShelfByIdAndUserParams{
			UserID: user.ID,
			ID:     atoi64(vars["shelf"]),
		})

		if !can(actor, "edit", shelf) {
			return Unauthorized
		}

		return Render("layout", "shelves/edit", map[string]interface{}{
			"current_user": actor,
			"user":         user,
			"shelf":        shelf,
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

		params := UpdateShelfParams{
			Name: r.FormValue("name"),
			ID:   shelf.ID,
		}

		errors := params.Validate()
		if len(errors) > 0 {
			return Render("layout", "shelves/edit", map[string]interface{}{
				"current_user": actor,
				"user":         user,
				"shelf":        params,
				"csrf":         csrf.TemplateField(r),
				"errors":       errors,
			})
		}

		err = queries.UpdateShelf(r.Context(), params)
		if err != nil {
			return InternalServerError(err)
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

		if !can(actor, "up", shelf) {
			return Unauthorized
		}

		err = queries.MoveShelfUp(r.Context(), shelf.ID)
		if err != nil {
			return InternalServerError(err)
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

		if !can(actor, "down", shelf) {
			return Unauthorized
		}

		err = queries.MoveShelfDown(r.Context(), shelf.ID)
		if err != nil {
			return InternalServerError(err)
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

		err = queries.RemoveShelf(r.Context(), shelf.ID)
		if err != nil {
			return InternalServerError(err)
		}

		err = queries.DeleteShelf(r.Context(), shelf.ID)
		if err != nil {
			return InternalServerError(err)
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
			"book":         book,
			"user":         user,
			"errors":       ValidationErrors{},
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

		r.ParseMultipartForm(MB * 10)

		params := NewHighlightParams{
			BookID:  book.ID,
			Page:    atoi32(r.FormValue("page")),
			Content: r.FormValue("content"),
		}
		errors := params.Validate()

		file, _, _ := r.FormFile("image")
		if file != nil {
			ValidateImage(file, "image", "Image", errors, 1000, 1000)
			file.Seek(0, os.SEEK_SET)
		}

		if len(errors) > 0 {
			return Render("layout", "highlights/new", map[string]interface{}{
				"current_user": actor,
				"book":         book,
				"user":         user,
				"highlight":    params,
				"errors":       errors,
				"csrf":         csrf.TemplateField(r),
			})
		}

		highlight, err := queries.NewHighlight(r.Context(), params)
		if err != nil {
			return InternalServerError(err)
		}

		if file != nil {
			name, err := UploadImage(file, HIGHLIGHT_IMAGE_PATH, 600, 600)
			if err != nil {
				return InternalServerError(err)
			}

			err = queries.UpdateHighlightImage(r.Context(), UpdateHighlightImageParams{
				Image: NullString(name),
				ID:    highlight.ID,
			})
			if err != nil {
				return InternalServerError(err)
			}
		}

		return Redirect(fmt.Sprintf("/users/%s/books/%s", user.Slug, book.Isbn))
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

		return Render("layout", "highlights/new", map[string]interface{}{
			"current_user": actor,
			"book":         book,
			"user":         user,
			"highlight":    highlight,
			"errors":       ValidationErrors{},
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

		params := UpdateHighlightParams{
			Page:    atoi32(r.FormValue("page")),
			Content: r.FormValue("content"),
			ID:      highlight.ID,
		}
		errors := params.Validate()

		file, _, _ := r.FormFile("image")
		if file != nil {
			ValidateImage(file, "image", "Image", errors, 1000, 1000)
			file.Seek(0, os.SEEK_SET)
		}

		if len(errors) > 0 {
			highlight.Content = params.Content
			highlight.Page = params.Page
			return Render("layout", "highlights/new", map[string]interface{}{
				"current_user": actor,
				"user":         user,
				"book":         book,
				"highlight":    highlight,
				"errors":       errors,
				"csrf":         csrf.TemplateField(r),
			})
		}

		err = queries.UpdateHighlight(r.Context(), params)
		if err != nil {
			return InternalServerError(err)
		}

		if file != nil {
			name, err := UploadImage(file, HIGHLIGHT_IMAGE_PATH, 1000, 1000)
			if err != nil {
				return InternalServerError(err)
			}

			err = queries.UpdateHighlightImage(r.Context(), UpdateHighlightImageParams{
				Image: NullString(name),
				ID:    highlight.ID,
			})
			if err != nil {
				return InternalServerError(err)
			}
		}

		return Redirect(fmt.Sprintf("/users/%s/books/%s", user.Slug, book.Isbn))
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

		highlight, err := queries.HighlightByIDAndBook(r.Context(), HighlightByIDAndBookParams{
			ID:     atoi64(vars["id"]),
			BookID: book.ID,
		})
		if err != nil {
			return NotFound
		}

		if !can(actor, "delete_highlight", book) {
			return Unauthorized
		}

		if highlight.Image.Valid && len(highlight.Image.String) > 0 {
			os.Remove(path.Join(HIGHLIGHT_IMAGE_PATH, highlight.Image.String))
		}

		err = queries.DeleteHighlight(r.Context(), highlight.ID)
		if err != nil {
			return InternalServerError(err)
		}

		return Redirect(fmt.Sprintf("/users/%s/books/%s", user.Slug, vars["isbn"]))
	}, loggedinMiddleware)

	Helpers()
	Start()
}
