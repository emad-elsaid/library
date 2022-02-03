package main

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"reflect"
	"strings"
)

func Helpers() {
	helpers["partial"] = func(path string, data interface{}) (template.HTML, error) {
		return template.HTML(partial(path, data)), nil
	}

	helpers["meta_property"] = func(meta map[string]string, name string) string {
		if meta == nil {
			return ""
		}

		v, ok := meta[name]
		if !ok {
			return ""
		}

		return fmt.Sprintf(`<meta property="%s" value="%s"/>`, template.HTMLEscapeString(name), template.HTMLEscapeString(v))
	}

	helpers["meta_name"] = func(meta map[string]string, name string) string {
		if meta == nil {
			return ""
		}

		v, ok := meta[name]
		if !ok {
			return ""
		}

		return fmt.Sprintf(`<meta name="%s" value="%s"/>`, name, v)
	}

	helpers["can"] = can

	helpers["include"] = func(list []string, str string) bool {
		for _, i := range list {
			if i == str {
				return true
			}
		}

		return false
	}

	helpers["book_cover"] = func(image, google_books_id string) string {
		if len(image) > 0 {
			return "/books/image/" + image
		}

		if len(google_books_id) > 0 {
			const googleBookURL = "https://books.google.com/books/content?id=%s&printsec=frontcover&img=1&zoom=1"
			return fmt.Sprintf(googleBookURL, google_books_id)
		}

		return "/default_book"
	}

	helpers["simple_format"] = func(str string) (template.HTML, error) {
		return template.HTML(strings.ReplaceAll(template.HTMLEscapeString(str), "\n", "<br/>")), nil
	}

	helpers["shelf_books"] = func(shelfID int64) ([]ShelfBooksRow, error) {
		return queries.ShelfBooks(context.Background(), sql.NullInt32{Valid: true, Int32: int32(shelfID)})
	}

	helpers["has_field"] = func(v interface{}, name string) bool {
		rv := reflect.ValueOf(v)
		if rv.Kind() == reflect.Ptr {
			rv = rv.Elem()
		}
		if rv.Kind() == reflect.Struct {
			return rv.FieldByName(name).IsValid()
		}
		if rv.Kind() == reflect.Map {
			val := rv.MapIndex(reflect.ValueOf(name))
			return val.IsValid()
		}
		return false
	}
}

func loggedin(r *http.Request) bool {
	_, ok := SESSION(r).Values["current_user"]
	return ok
}

func current_user(r *http.Request) *User {
	id, ok := SESSION(r).Values["current_user"]
	if !ok {
		return nil
	}

	user_id, ok := id.(int64)
	if !ok {
		return nil
	}

	user, err := queries.User(r.Context(), user_id)
	if err != nil {
		return nil
	}

	return &user
}

func can(who *User, do string, what interface{}) bool {
	err := fmt.Sprintf("Verb %s not handled for %#v", do, what)

	switch w := what.(type) {
	case nil:
		switch do {
		case "login":
			return who == nil
		case "logout":
			return who != nil
		default:
			log.Fatal(err)
		}

	case User:
		switch do {
		case "create_book":
			return who != nil && who.ID == w.ID
		case "list_shelves":
			return who != nil && who.ID == w.ID
		default:
			log.Fatal(err)
		}

	case *User:
		switch do {
		case "create_book":
			return who != nil && who.ID == w.ID
		case "list_shelves":
			return who != nil && who.ID == w.ID
		default:
			log.Fatal(err)
		}

	case BookByIsbnAndUserRow:
		switch do {
		case "edit", "highlight":
			return who != nil && who.ID == w.UserID
		}

	default:
		log.Fatal(err)
	}
	return true
}
