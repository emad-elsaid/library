package main

import (
	"context"
	"crypto/sha256"
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
	"strconv"
	"archive/zip"
)


func Helpers() {
	HELPER("partial", func(path string, data interface{}) (template.HTML, error) {
		return template.HTML(partial(path, data)), nil
	})

	HELPER("meta_property", func(meta map[string]string, name string) template.HTML {
		if meta == nil {
			return ""
		}

		v, ok := meta[name]
		if !ok {
			return ""
		}

		tag := fmt.Sprintf(`<meta property="%s" value="%s"/>`, template.HTMLEscapeString(name), template.HTMLEscapeString(v))
		return template.HTML(tag)
	})

	HELPER("meta_name", func(meta map[string]string, name string) template.HTML {
		if meta == nil {
			return ""
		}

		v, ok := meta[name]
		if !ok {
			return ""
		}

		tag := fmt.Sprintf(`<meta name="%s" value="%s"/>`, template.HTMLEscapeString(name), template.HTMLEscapeString(v))
		return template.HTML(tag)
	})

	HELPER("can", can)

	HELPER("include", func(list []string, str string) bool {
		for _, i := range list {
			if i == str {
				return true
			}
		}

		return false
	})

	HELPER("book_cover", book_cover)

	HELPER("simple_format", func(str string) (template.HTML, error) {
		return template.HTML(strings.ReplaceAll(template.HTMLEscapeString(str), "\n", "<br/>")), nil
	})

	HELPER("shelf_books", func(shelfID int64) ([]ShelfBooksRow, error) {
		return Q.ShelfBooks(context.Background(), sql.NullInt64{Valid: true, Int64: shelfID})
	})

	HELPER("has_field", func(v interface{}, name string) bool {
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
	})

	HELPER("last", func(v interface{}) interface{} {
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
	})

	HELPER("books_count", func(u int64) int64 {
		c, _ := Q.BooksCount(context.Background(), u)
		return c
	})

	HELPER("sha256", func() interface{} {
		cache := map[string]string{}
		return func(p string) (string, error) {
			if v, ok := cache[p]; ok {
				return v, nil
			}

			f, err := os.Open(p)
			if err != nil {
				return "", err
			}

			d, err := io.ReadAll(f)
			if err != nil {
				return "", err
			}

			cache[p] = fmt.Sprintf("%x", sha256.Sum256(d))
			return cache[p], nil
		}
	}())
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

	user, err := Q.User(r.Context(), user_id)
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

func DownloadFile(w http.ResponseWriter, r *http.Request, file string) string {
	Openfile, err := os.Open(file) //Open the file to be downloaded later
	defer Openfile.Close() //Close after function return

	if err != nil {		
		http.Error(w, "File not found.", 404)
		return ""
	}

	tempBuffer := make([]byte, 512) //Create a byte array to read the file later
	Openfile.Read(tempBuffer) //Read the file into  byte
	FileContentType := http.DetectContentType(tempBuffer) //Get file header

	FileStat, _ := Openfile.Stat() //Get info from file
	FileSize := strconv.FormatInt(FileStat.Size(), 10) //Get file size as a string

	Filename := file

	//Set the headers
	w.Header().Set("Content-Disposition", "attachment; filename="+ Filename)
	w.Header().Set("Content-Type", FileContentType+";"+Filename)
	w.Header().Set("Content-Length", FileSize)

	Openfile.Seek(0, 0) //We read 512 bytes from the file already so we reset the offset back to 0
	io.Copy(w, Openfile) //'Copy' the file to the client
	return file
}

func DownloadZipFile(w http.ResponseWriter, r *http.Request, file string, newLocation string) string {
	// Openfile, err := os.Open(file) //Open the file to be downloaded later
	// defer Openfile.Close() //Close after function return

	// if err != nil {		
	// 	http.Error(w, "File not found.", 404) //return 404 if file is not found
	// 	return ""
	// }

	// tempBuffer := make([]byte, 512) //Create a byte array to read the file later
	// Openfile.Read(tempBuffer) //Read the file into  byte
	// FileContentType := http.DetectContentType(tempBuffer) //Get file header

	// FileStat, _ := Openfile.Stat() //Get info from file
	// FileSize := strconv.FormatInt(FileStat.Size(), 10) //Get file size as a string

	// Filename := file

	// //Set the headers
	// w.Header().Set("Content-Type", FileContentType+";"+Filename)
	// w.Header().Set("Content-Length", FileSize)

	// Openfile.Seek(0, 0) //We read 512 bytes from the file already so we reset the offset back to 0
	// io.Copy(w, Openfile) //'Copy' the file to the client
	// return file
	
	// move the zip file from the public folder to the export folder
	MoveFile(file, newLocation)
	w.Header().Set("Content-type", "application/zip")
	http.ServeFile(w, r, "export/library.zip")
	//defer os.Remove(file)	
	return ""

}

func appendFiles(filename string, zipw *zip.Writer) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("Failed to open %s: %s", filename, err)
	}
	defer file.Close()
 
	wr, err := zipw.Create(filename)
	if err != nil {
		msg := "Failed to create entry for %s in zip file: %s"
		return fmt.Errorf(msg, filename, err)
	}
 
	if _, err := io.Copy(wr, file); err != nil {
		return fmt.Errorf("Failed to write %s to zip: %s", filename, err)
	}
 
	return nil
}

func MoveFile(oldLocation string, newLocation string) string {
	err := os.Rename(oldLocation, newLocation)
	if err != nil {
		return ""
	}
	return ""
}
 
