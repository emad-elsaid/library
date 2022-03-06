package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

const (
	BOOK_COVER_PATH      = "public/books/image"
	HIGHLIGHT_IMAGE_PATH = "public/highlights/image"
)

func main() {
	google := &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("DOMAIN") + "/auth/google/callback",
		Scopes:       []string{"email", "profile"},
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://accounts.google.com/o/oauth2/auth",
			TokenURL:  "https://oauth2.googleapis.com/token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}

	GET("/", func(w Response, r Request) Output {
		user := current_user(r)
		if user != nil {
			return Redirect(fmt.Sprintf("/users/%s", user.Slug))
		}

		return Render("wide_layout", "index", Locals{"csrf": CSRF(r)})
	})

	GET("/privacy", func(w Response, r Request) Output {
		return Render("layout", "privacy", Locals{
			"current_user": current_user(r),
			"csrf":         CSRF(r),
		})
	})

	POST("/auth/google", func(w Response, r Request) Output {
		state := uuid.New().String()
		s := SESSION(r)
		s.Values["state"] = state
		if err := s.Save(r, w); err != nil {
			return InternalServerError(err)
		}

		return Redirect(google.AuthCodeURL(state))
	})

	GET("/auth/google/callback", func(w Response, r Request) Output {
		state := SESSION(r).Values["state"]
		param := r.FormValue("state")
		if state != param {
			return BadRequest
		}

		tok, err := google.Exchange(context.Background(), r.URL.Query().Get("code"))
		if err != nil {
			return BadRequest
		}

		client := google.Client(context.Background(), tok)
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

		if err = json.Unmarshal(body, &user); err != nil {
			return InternalServerError(err)
		}

		u, err := Q.Signup(r.Context(), SignupParams{
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
		vars := VARS(r)

		user, err := Q.UserBySlug(r.Context(), vars["user"])
		if err != nil {
			return NotFound
		}

		data := Locals{
			"title":        user.Name.String,
			"csrf":         CSRF(r),
			"current_user": current_user(r),
			"user":         user,
		}

		unshelved_books, err := Q.UserUnshelvedBooks(r.Context(), user.ID)
		if err != nil {
			return InternalServerError(err)
		}
		if len(unshelved_books) > 0 {
			data["unshelved_books"] = unshelved_books
		}

		data["shelves"], err = Q.Shelves(r.Context(), user.ID)

		return Render("layout", "users/show", data)
	})

	GET("/users/{user}/edit", func(w Response, r Request) Output {
		actor := current_user(r)
		vars := VARS(r)

		user, err := Q.UserBySlug(r.Context(), vars["user"])
		if err != nil {
			return NotFound
		}

		if !can(actor, "edit", user) {
			return Unauthorized
		}

		return Render("layout", "users/edit", Locals{
			"current_user": actor,
			"user":         user,
			"errors":       ValidationErrors{},
			"csrf":         CSRF(r),
		})
	}, loggedinMiddleware)

	POST("/users/{user}", func(w Response, r Request) Output {
		actor := current_user(r)
		vars := VARS(r)

		user, err := Q.UserBySlug(r.Context(), vars["user"])
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
			return Render("layout", "users/edit", Locals{
				"current_user": actor,
				"user":         user,
				"errors":       errors,
				"csrf":         CSRF(r),
			})
		}

		if err = Q.UpdateUser(r.Context(), params); err != nil {
			return InternalServerError(err)
		}

		return Redirect(fmt.Sprintf("/users/%s", user.Slug))
	}, loggedinMiddleware)

	GET("/users/{user}/books/new", func(w Response, r Request) Output {
		actor := current_user(r)
		vars := VARS(r)

		user, err := Q.UserBySlug(r.Context(), vars["user"])
		if err != nil {
			return NotFound
		}

		if !can(actor, "create_book", user) {
			return Unauthorized
		}

		return Render("layout", "books/new", Locals{
			"current_user": actor,
			"user":         user,
			"errors":       ValidationErrors{},
			"csrf":         CSRF(r),
		})
	}, loggedinMiddleware)

	POST("/users/{user}/books", func(w Response, r Request) Output {
		actor := current_user(r)
		vars := VARS(r)

		user, err := Q.UserBySlug(r.Context(), vars["user"])
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
			PageRead:      atoi32(r.FormValue("page_read")),
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
			return Render("layout", "books/new", Locals{
				"book":         params,
				"current_user": actor,
				"user":         user,
				"errors":       errors,
				"csrf":         CSRF(r),
			})
		}

		book, err := Q.NewBook(r.Context(), params)
		if err != nil {
			return InternalServerError(err)
		}

		if file != nil {
			name, err := UploadImage(file, BOOK_COVER_PATH, 432, 576)
			if err != nil {
				return InternalServerError(err)
			}

			err = Q.UpdateBookImage(r.Context(), UpdateBookImageParams{
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
		vars := VARS(r)

		user, err := Q.UserBySlug(r.Context(), vars["user"])
		if err != nil {
			return InternalServerError(err)
		}

		book, err := Q.BookByIsbnAndUser(r.Context(), BookByIsbnAndUserParams{
			UserID: user.ID,
			Isbn:   vars["isbn"],
		})
		if err != nil {
			return NotFound
		}

		highlights, err := Q.Highlights(r.Context(), book.ID)
		if err != nil {
			return InternalServerError(err)
		}

		shelves, err := Q.Shelves(r.Context(), user.ID)
		if err != nil {
			return InternalServerError(err)
		}

		return Render("layout", "books/show", Locals{
			"current_user": current_user(r),
			"user":         user,
			"title":        book.Title,
			"book":         book,
			"shelves":      shelves,
			"highlights":   highlights,
			"csrf":         CSRF(r),
			"meta": map[string]string{
				"og:title":       book.Title,
				"author":         book.Author,
				"description":    book.Description,
				"og:description": book.Description,
				"og:type":        "article",
				"og:image":       book_cover(book.Image.String, book.GoogleBooksID.String),
				"twitter:image":  book_cover(book.Image.String, book.GoogleBooksID.String),
				"twitter:card":   "summary",
				"twitter:title":  book.Title,
			},
		})
	})

	GET("/users/{user}/books/{isbn}/edit", func(w Response, r Request) Output {
		actor := current_user(r)
		vars := VARS(r)

		user, err := Q.UserBySlug(r.Context(), vars["user"])
		if err != nil {
			return NotFound
		}

		book, err := Q.BookByIsbnAndUser(r.Context(), BookByIsbnAndUserParams{
			UserID: user.ID,
			Isbn:   vars["isbn"],
		})
		if err != nil {
			return NotFound
		}

		if !can(actor, "edit", book) {
			return Unauthorized
		}

		return Render("layout", "books/new", Locals{
			"current_user": actor,
			"user":         user,
			"book":         book,
			"csrf":         CSRF(r),
			"errors":       ValidationErrors{},
		})
	}, loggedinMiddleware)

	POST("/users/{user}/books/{isbn}", func(w Response, r Request) Output {
		actor := current_user(r)
		vars := VARS(r)

		user, err := Q.UserBySlug(r.Context(), vars["user"])
		if err != nil {
			return NotFound
		}

		book, err := Q.BookByIsbnAndUser(r.Context(), BookByIsbnAndUserParams{
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
			PageRead:    atoi32(r.FormValue("page_read")),
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
			book.PageRead = params.PageRead
			return Render("layout", "books/new", Locals{
				"current_user": actor,
				"user":         user,
				"book":         book,
				"csrf":         CSRF(r),
				"errors":       errors,
			})
		}

		if err = Q.UpdateBook(r.Context(), params); err != nil {
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

			err = Q.UpdateBookImage(r.Context(), UpdateBookImageParams{
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
		vars := VARS(r)

		user, err := Q.UserBySlug(r.Context(), vars["user"])
		if err != nil {
			return NotFound
		}

		book, err := Q.BookByIsbnAndUser(r.Context(), BookByIsbnAndUserParams{
			UserID: user.ID,
			Isbn:   vars["isbn"],
		})
		if err != nil {
			return NotFound
		}

		if !can(actor, "delete", book) {
			return Unauthorized
		}

		images, err := Q.HighlightsWithImages(r.Context(), book.ID)
		if err != nil {
			return InternalServerError(err)
		}
		for _, v := range images {
			os.Remove(path.Join(HIGHLIGHT_IMAGE_PATH, v.String))
		}

		if book.Image.Valid && len(book.Image.String) > 0 {
			os.Remove(path.Join(BOOK_COVER_PATH, book.Image.String))
		}

		if err = Q.DeleteBook(r.Context(), book.ID); err != nil {
			return InternalServerError(err)
		}

		return Redirect(fmt.Sprintf("/users/%s", user.Slug))
	}, loggedinMiddleware)

	POST("/users/{user}/books/{isbn}/shelf", func(w Response, r Request) Output {
		actor := current_user(r)
		vars := VARS(r)

		user, err := Q.UserBySlug(r.Context(), vars["user"])
		if err != nil {
			return NotFound
		}

		book, err := Q.BookByIsbnAndUser(r.Context(), BookByIsbnAndUserParams{
			UserID: user.ID,
			Isbn:   vars["isbn"],
		})
		if err != nil {
			return NotFound
		}

		if !can(actor, "edit", book) {
			return Unauthorized
		}

		shelf, err := Q.ShelfByIdAndUser(r.Context(), ShelfByIdAndUserParams{
			UserID: user.ID,
			ID:     atoi64(r.FormValue("shelf_id")),
		})
		if err == nil && !can(actor, "edit", shelf) {
			return Unauthorized
		}

		err = Q.MoveBookToShelf(r.Context(), MoveBookToShelfParams{
			ShelfID: sql.NullInt64{Int64: shelf.ID, Valid: err == nil},
			ID:      book.ID,
		})
		if err != nil {
			return InternalServerError(err)
		}

		return Redirect(fmt.Sprintf("/users/%s/books/%s", user.Slug, vars["isbn"]))
	}, loggedinMiddleware)

	POST("/users/{user}/books/{isbn}/complete", func(w Response, r Request) Output {
		actor := current_user(r)
		vars := VARS(r)

		user, err := Q.UserBySlug(r.Context(), vars["user"])
		if err != nil {
			return NotFound
		}

		book, err := Q.BookByIsbnAndUser(r.Context(), BookByIsbnAndUserParams{
			UserID: user.ID,
			Isbn:   vars["isbn"],
		})
		if err != nil {
			return NotFound
		}

		if !can(actor, "edit", book) {
			return Unauthorized
		}

		err = Q.CompleteBook(r.Context(), book.ID)
		if err != nil {
			return InternalServerError(err)
		}

		return Redirect(fmt.Sprintf("/users/%s/books/%s", user.Slug, book.Isbn))
	}, loggedinMiddleware)

	GET("/users/{user}/shelves", func(w Response, r Request) Output {
		actor := current_user(r)
		vars := VARS(r)

		user, err := Q.UserBySlug(r.Context(), vars["user"])
		if err != nil {
			return NotFound
		}

		if !can(actor, "show_shelves", user) {
			return Unauthorized
		}

		shelves, err := Q.Shelves(r.Context(), user.ID)
		if err != nil {
			return InternalServerError(err)
		}

		return Render("layout", "shelves/index", Locals{
			"current_user": actor,
			"user":         user,
			"shelves":      shelves,
			"errors":       ValidationErrors{},
			"csrf":         CSRF(r),
		})
	}, loggedinMiddleware)

	POST("/users/{user}/shelves", func(w Response, r Request) Output {
		actor := current_user(r)
		vars := VARS(r)

		user, err := Q.UserBySlug(r.Context(), vars["user"])
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
			shelves, err := Q.Shelves(r.Context(), user.ID)
			if err != nil {
				return InternalServerError(err)
			}

			return Render("layout", "shelves/index", Locals{
				"current_user": actor,
				"user":         user,
				"shelves":      shelves,
				"shelf":        params,
				"csrf":         CSRF(r),
				"errors":       errors,
			})
		}

		if err = Q.NewShelf(r.Context(), params); err != nil {
			return InternalServerError(err)
		}

		return Redirect(fmt.Sprintf("/users/%s/shelves", user.Slug))
	}, loggedinMiddleware)

	GET("/users/{user}/shelves/{shelf}/edit", func(w Response, r Request) Output {
		actor := current_user(r)
		vars := VARS(r)

		user, err := Q.UserBySlug(r.Context(), vars["user"])
		if err != nil {
			return NotFound
		}

		shelf, err := Q.ShelfByIdAndUser(r.Context(), ShelfByIdAndUserParams{
			UserID: user.ID,
			ID:     atoi64(vars["shelf"]),
		})

		if !can(actor, "edit", shelf) {
			return Unauthorized
		}

		return Render("layout", "shelves/edit", Locals{
			"current_user": actor,
			"user":         user,
			"errors":       ValidationErrors{},
			"shelf":        shelf,
			"csrf":         CSRF(r),
		})
	}, loggedinMiddleware)

	POST("/users/{user}/shelves/{shelf}", func(w Response, r Request) Output {
		actor := current_user(r)
		vars := VARS(r)

		user, err := Q.UserBySlug(r.Context(), vars["user"])
		if err != nil {
			return NotFound
		}

		shelf, err := Q.ShelfByIdAndUser(r.Context(), ShelfByIdAndUserParams{
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
			return Render("layout", "shelves/edit", Locals{
				"current_user": actor,
				"user":         user,
				"shelf":        params,
				"csrf":         CSRF(r),
				"errors":       errors,
			})
		}

		if err = Q.UpdateShelf(r.Context(), params); err != nil {
			return InternalServerError(err)
		}

		return Redirect(fmt.Sprintf("/users/%s/shelves", user.Slug))
	}, loggedinMiddleware)

	POST("/users/{user}/shelves/{shelf}/up", func(w Response, r Request) Output {
		actor := current_user(r)
		vars := VARS(r)

		user, err := Q.UserBySlug(r.Context(), vars["user"])
		if err != nil {
			return NotFound
		}

		shelf, err := Q.ShelfByIdAndUser(r.Context(), ShelfByIdAndUserParams{
			UserID: user.ID,
			ID:     atoi64(vars["shelf"]),
		})
		if err != nil {
			return NotFound
		}

		if !can(actor, "up", shelf) {
			return Unauthorized
		}

		if err = Q.MoveShelfUp(r.Context(), shelf.ID); err != nil {
			return InternalServerError(err)
		}

		return Redirect(fmt.Sprintf("/users/%s/shelves", user.Slug))
	}, loggedinMiddleware)

	POST("/users/{user}/shelves/{shelf}/down", func(w Response, r Request) Output {
		actor := current_user(r)
		vars := VARS(r)

		user, err := Q.UserBySlug(r.Context(), vars["user"])
		if err != nil {
			return NotFound
		}

		shelf, err := Q.ShelfByIdAndUser(r.Context(), ShelfByIdAndUserParams{
			UserID: user.ID,
			ID:     atoi64(vars["shelf"]),
		})
		if err != nil {
			return NotFound
		}

		if !can(actor, "down", shelf) {
			return Unauthorized
		}

		if err = Q.MoveShelfDown(r.Context(), shelf.ID); err != nil {
			return InternalServerError(err)
		}

		return Redirect(fmt.Sprintf("/users/%s/shelves", user.Slug))
	}, loggedinMiddleware)

	DELETE("/users/{user}/shelves/{shelf}", func(w Response, r Request) Output {
		actor := current_user(r)
		vars := VARS(r)

		user, err := Q.UserBySlug(r.Context(), vars["user"])
		if err != nil {
			return NotFound
		}

		shelf, err := Q.ShelfByIdAndUser(r.Context(), ShelfByIdAndUserParams{
			UserID: user.ID,
			ID:     atoi64(vars["shelf"]),
		})
		if err != nil {
			return NotFound
		}

		if !can(actor, "delete", shelf) {
			return Unauthorized
		}

		if err = Q.RemoveShelf(r.Context(), shelf.ID); err != nil {
			return InternalServerError(err)
		}

		if err = Q.DeleteShelf(r.Context(), shelf.ID); err != nil {
			return InternalServerError(err)
		}

		return Redirect(fmt.Sprintf("/users/%s/shelves", user.Slug))
	}, loggedinMiddleware)

	GET("/users/{user}/books/{isbn}/highlights/new", func(w Response, r Request) Output {
		actor := current_user(r)
		vars := VARS(r)

		user, err := Q.UserBySlug(r.Context(), vars["user"])
		if err != nil {
			return NotFound
		}

		book, err := Q.BookByIsbnAndUser(r.Context(), BookByIsbnAndUserParams{
			UserID: user.ID,
			Isbn:   vars["isbn"],
		})
		if err != nil {
			return NotFound
		}

		if !can(actor, "create_highlight", book) {
			return Unauthorized
		}

		return Render("layout", "highlights/new", Locals{
			"current_user": actor,
			"book":         book,
			"user":         user,
			"errors":       ValidationErrors{},
			"csrf":         CSRF(r),
		})
	}, loggedinMiddleware)

	POST("/users/{user}/books/{isbn}/highlights", func(w Response, r Request) Output {
		actor := current_user(r)
		vars := VARS(r)

		user, err := Q.UserBySlug(r.Context(), vars["user"])
		if err != nil {
			return NotFound
		}

		book, err := Q.BookByIsbnAndUser(r.Context(), BookByIsbnAndUserParams{
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
			return Render("layout", "highlights/new", Locals{
				"current_user": actor,
				"book":         book,
				"user":         user,
				"highlight":    params,
				"errors":       errors,
				"csrf":         CSRF(r),
			})
		}

		highlight, err := Q.NewHighlight(r.Context(), params)
		if err != nil {
			return InternalServerError(err)
		}

		if file != nil {
			name, err := UploadImage(file, HIGHLIGHT_IMAGE_PATH, 600, 600)
			if err != nil {
				return InternalServerError(err)
			}

			err = Q.UpdateHighlightImage(r.Context(), UpdateHighlightImageParams{
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
		vars := VARS(r)

		user, err := Q.UserBySlug(r.Context(), vars["user"])
		if err != nil {
			return NotFound
		}

		book, err := Q.BookByIsbnAndUser(r.Context(), BookByIsbnAndUserParams{
			UserID: user.ID,
			Isbn:   vars["isbn"],
		})
		if err != nil {
			return NotFound
		}

		highlight, err := Q.HighlightByIDAndBook(r.Context(), HighlightByIDAndBookParams{
			ID:     atoi64(vars["id"]),
			BookID: book.ID,
		})
		if err != nil {
			return NotFound
		}

		if !can(actor, "edit_highlight", book) {
			return Unauthorized
		}

		return Render("layout", "highlights/new", Locals{
			"current_user": actor,
			"book":         book,
			"user":         user,
			"highlight":    highlight,
			"errors":       ValidationErrors{},
			"csrf":         CSRF(r),
		})
	}, loggedinMiddleware)

	POST("/users/{user}/books/{isbn}/highlights/{id}", func(w Response, r Request) Output {
		actor := current_user(r)
		vars := VARS(r)

		user, err := Q.UserBySlug(r.Context(), vars["user"])
		if err != nil {
			return NotFound
		}

		book, err := Q.BookByIsbnAndUser(r.Context(), BookByIsbnAndUserParams{
			UserID: user.ID,
			Isbn:   vars["isbn"],
		})
		if err != nil {
			return NotFound
		}

		highlight, err := Q.HighlightByIDAndBook(r.Context(), HighlightByIDAndBookParams{
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
			return Render("layout", "highlights/new", Locals{
				"current_user": actor,
				"user":         user,
				"book":         book,
				"highlight":    highlight,
				"errors":       errors,
				"csrf":         CSRF(r),
			})
		}

		if err = Q.UpdateHighlight(r.Context(), params); err != nil {
			return InternalServerError(err)
		}

		if file != nil {
			name, err := UploadImage(file, HIGHLIGHT_IMAGE_PATH, 1000, 1000)
			if err != nil {
				return InternalServerError(err)
			}

			err = Q.UpdateHighlightImage(r.Context(), UpdateHighlightImageParams{
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
		vars := VARS(r)

		user, err := Q.UserBySlug(r.Context(), vars["user"])
		if err != nil {
			return NotFound
		}

		book, err := Q.BookByIsbnAndUser(r.Context(), BookByIsbnAndUserParams{
			UserID: user.ID,
			Isbn:   vars["isbn"],
		})
		if err != nil {
			return NotFound
		}

		highlight, err := Q.HighlightByIDAndBook(r.Context(), HighlightByIDAndBookParams{
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

		if err = Q.DeleteHighlight(r.Context(), highlight.ID); err != nil {
			return InternalServerError(err)
		}

		return Redirect(fmt.Sprintf("/users/%s/books/%s", user.Slug, vars["isbn"]))
	}, loggedinMiddleware)

	Helpers()
	Start()
}
