package main

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"reflect"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/image/draw"
)

func Helpers() {
	helpers["partial"] = func(path string, data interface{}) (template.HTML, error) {
		return template.HTML(partial(path, data)), nil
	}

	helpers["meta_property"] = func(meta map[string]string, name string) template.HTML {
		if meta == nil {
			return ""
		}

		v, ok := meta[name]
		if !ok {
			return ""
		}

		tag := fmt.Sprintf(`<meta property="%s" value="%s"/>`, template.HTMLEscapeString(name), template.HTMLEscapeString(v))
		return template.HTML(tag)
	}

	helpers["meta_name"] = func(meta map[string]string, name string) template.HTML {
		if meta == nil {
			return ""
		}

		v, ok := meta[name]
		if !ok {
			return ""
		}

		tag := fmt.Sprintf(`<meta name="%s" value="%s"/>`, template.HTMLEscapeString(name), template.HTMLEscapeString(v))
		return template.HTML(tag)
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

	helpers["book_cover"] = book_cover

	helpers["simple_format"] = func(str string) (template.HTML, error) {
		return template.HTML(strings.ReplaceAll(template.HTMLEscapeString(str), "\n", "<br/>")), nil
	}

	helpers["shelf_books"] = func(shelfID int64) ([]ShelfBooksRow, error) {
		return queries.ShelfBooks(context.Background(), sql.NullInt64{Valid: true, Int64: shelfID})
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

	helpers["last"] = func(v interface{}) interface{} {
		rv := reflect.ValueOf(v)
		if rv.Kind() == reflect.Ptr {
			rv = rv.Elem()
		}

		if rv.Kind() != reflect.Slice {
			return false
		}

		if rv.Len() == 0 {
			return false
		}

		return rv.Index(rv.Len() - 1)
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

func book_cover(image, google_books_id string) string {
	if len(image) > 0 {
		return "/books/image/" + image
	}

	if len(google_books_id) > 0 {
		const googleBookURL = "https://books.google.com/books/content?id=%s&printsec=frontcover&img=1&zoom=1"
		return fmt.Sprintf(googleBookURL, google_books_id)
	}

	return "/default_book"
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
		case "create_book", "list_shelves", "edit", "create_shelf", "show_shelves":
			return who != nil && who.ID == w.ID
		default:
			log.Fatal(err)
		}

	case *User:
		switch do {
		case "create_book", "list_shelves":
			return who != nil && who.ID == w.ID
		default:
			log.Fatal(err)
		}

	case BookByIsbnAndUserRow:
		switch do {
		case "edit", "highlight", "create_highlight", "edit_highlight", "delete", "delete_highlight":
			return who != nil && who.ID == w.UserID
		default:
			log.Fatal(err)
		}

	case Shelf:
		switch do {
		case "edit", "delete":
			return who != nil && who.ID == w.UserID
		case "up":
			return who != nil && who.ID == w.UserID && w.Position > 1
		case "down":
			return who != nil && who.ID == w.UserID
		default:
			log.Fatal(err)
		}

	default:
		log.Fatal(err)
	}
	return true
}

func ImageResize(in io.Reader, out io.Writer, w, h int) error {
	src, _, err := image.Decode(in)
	if err != nil {
		return err
	}

	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.NearestNeighbor.Scale(dst, dst.Rect, src, src.Bounds(), draw.Over, nil)

	return jpeg.Encode(out, dst, &jpeg.Options{Quality: 90})
}

func UploadImage(in io.Reader, p string, w, h int) (string, error) {
	name := uuid.New().String()

	out, err := os.Create(path.Join(p, name))
	if err != nil {
		return "", err
	}

	err = ImageResize(in, out, w, h)
	if err != nil {
		return "", err
	}

	return name, nil
}
