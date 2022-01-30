//go:generate sqlc generate
package main

// This is an experiment

// This file should be copied to new projects that needs connection to DB, HTTP
// server, views and helpers setup. So it should do as much work as possible
// just by including it in the project. I imagine that I will use it by copying
// the code instead of referencing it. and changing the constants to what I
// think is suitable for the new project.

// HOW TO USE

// 1. Copy the file to your project. set environment variables. check the constants values.
// 2. Make sure you have sqlc.yaml
// 3. Write queries in query.sql
// 4. Everytime you edit query.sql run `go generate`
// 5. Use `router` to add your gorilla routes, or shorthand methods GET, POST, DELETE...etc
// 6. Add Helpers to `helpers` map
// 7. call `Start()` to start the server

// ENV Variables
// =============
// DATABASE_URL : postgres database URL

import (
	"bytes"
	"context"
	"database/sql"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "embed"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

const (
	MAX_DB_OPEN_CONNECTIONS = 5
	MAX_DB_IDLE_CONNECTIONS = 5
	HTTP_ROOT_PATH          = "/"
	STATIC_DIR_PATH         = "public"
	BIND_ADDRESS            = "127.0.0.1:3000"
	VIEWS_EXTENSION         = ".html"
	SESSION_COOKIE_NAME     = "library"
)

var (
	// queries functions as a result to sqlc compilation
	queries *Queries
	router  *mux.Router
	session *sessions.CookieStore
)

type QueryLogger struct {
	db     *sqlx.DB
	logger *log.Logger
}

func (p QueryLogger) ExecContext(ctx context.Context, q string, args ...interface{}) (sql.Result, error) {
	p.logger.Print(q)
	p.logger.Print(args...)
	return p.db.ExecContext(ctx, q, args...)
}
func (p QueryLogger) PrepareContext(ctx context.Context, q string) (*sql.Stmt, error) {
	p.logger.Print(q)
	return p.db.PrepareContext(ctx, q)
}
func (p QueryLogger) QueryContext(ctx context.Context, q string, args ...interface{}) (*sql.Rows, error) {
	p.logger.Print(q)
	p.logger.Print(args...)
	return p.db.QueryContext(ctx, q, args...)
}
func (p QueryLogger) QueryRowContext(ctx context.Context, q string, args ...interface{}) *sql.Row {
	p.logger.Print(q)
	p.logger.Print(args...)
	return p.db.QueryRowContext(ctx, q, args...)
}

func init() {
	db, err := connectToDB()
	if err != nil {
		log.Fatal(err)
	}

	ql := QueryLogger{db, log.Default()}
	queries = New(ql)

	router = mux.NewRouter()
	session = sessions.NewCookieStore([]byte(os.Getenv("SESSION_SECRET")))
}

func connectToDB() (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(MAX_DB_OPEN_CONNECTIONS)
	db.SetMaxIdleConns(MAX_DB_IDLE_CONNECTIONS)

	return db, err
}

func Start() {
	compileViews()

	router.PathPrefix("/").Handler(staticWithoutDirectoryListingHandler())

	http.Handle(HTTP_ROOT_PATH, router)

	srv := &http.Server{
		Handler:      router,
		Addr:         BIND_ADDRESS,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Printf("Starting server: %s", BIND_ADDRESS)
	log.Fatal(srv.ListenAndServe())
}

func staticWithoutDirectoryListingHandler() http.Handler {
	dir := http.Dir(STATIC_DIR_PATH)
	server := http.FileServer(dir)
	handler := http.StripPrefix("/", server)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}

		handler.ServeHTTP(w, r)
	})
}

// ROUTES HELPERS ==========================================

func GET(path string, handler http.HandlerFunc) {
	router.Methods("GET").Path(path).HandlerFunc(handler)
}

func POST(path string, handler http.HandlerFunc) {
	router.Methods("POST").Path(path).HandlerFunc(handler)
}

func DELETE(path string, handler http.HandlerFunc) {
	router.Methods("DELETE").Path(path).HandlerFunc(handler)
}

// VIEWS ====================

//go:embed views
var views embed.FS
var templates map[string]*template.Template = map[string]*template.Template{}
var helpers = template.FuncMap{}

func compileViews() {
	fs.WalkDir(views, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(path, VIEWS_EXTENSION) && d.Type().IsRegular() {
			name := strings.TrimPrefix(path, "views/")
			name = strings.TrimSuffix(name, VIEWS_EXTENSION)
			log.Printf("Parsing view: %s -> %s", path, name)

			c, err := fs.ReadFile(views, path)
			if err != nil {
				return err
			}

			templates[name] = template.Must(template.New(name).Funcs(helpers).Parse(string(c)))
		}

		return nil
	})
}

func partial(path string, data interface{}) string {
	v, ok := templates[path]
	if !ok {
		return "view %s not found"
	}

	w := bytes.NewBufferString("")
	err := v.Execute(w, data)
	if err != nil {
		return "rendering error " + path + " " + err.Error()
	}

	return w.String()
}

func render(w http.ResponseWriter, path string, view string, data map[string]interface{}) {
	v, ok := templates[path]
	if !ok {
		fmt.Fprintln(w, "layout %s not found", path)
	}

	data["yield"] = template.HTML(partial(view, nil))

	err := v.Execute(w, data)
	if err != nil {
		fmt.Fprintf(w, "rendering layout error %s", err.Error())
	}
}

// SESSION =================================

func SESSION(r *http.Request) *sessions.Session {
	s, _ := session.Get(r, SESSION_COOKIE_NAME)
	return s
}
