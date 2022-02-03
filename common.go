//go:generate sqlc generate
package main

// This is an experiment

// This file should be copied to new projects that needs connection to DB, HTTP
// server, views and helpers setup. So it should do as much work as possible
// just by including it in the project. I imagine that I will use it by copying
// the code instead of referencing it. and changing the constants to what I
// think is suitable for the new project.

// HOW TO USE

// 1. Copy common.go, .env.sample, sqlc.yaml
// 2. Write queries in query.sql and use `go generate` to generate functions with sqlc
// 3. Use `router` to add your gorilla routes, or shorthand methods GET, POST, DELETE...etc
// 4. Add Helpers to `helpers` map
// 5. call `Start()` to start the server

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
	"strconv"
	"strings"
	"time"

	_ "embed"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

const (
	MAX_DB_OPEN_CONNECTIONS = 5
	MAX_DB_IDLE_CONNECTIONS = 5
	STATIC_DIR_PATH         = "public"
	BIND_ADDRESS            = "127.0.0.1:3000"
	VIEWS_EXTENSION         = ".html"
	SESSION_COOKIE_NAME     = "library"
)

var (
	queries *Queries
	router  *mux.Router
	session *sessions.CookieStore
)

func init() {
	db, err := connectToDB()
	if err != nil {
		log.Fatal(err)
	}

	ql := QueryLogger{db, log.Default()}
	queries = New(ql)

	router = mux.NewRouter()
	session = sessions.NewCookieStore([]byte(os.Getenv("SESSION_SECRET")))
	csrf.TemplateTag = "csrf"
}

func Start() {
	compileViews()

	router.PathPrefix("/").Handler(staticWithoutDirectoryListingHandler())
	csrfMiddleware := csrf.Protect([]byte(os.Getenv("SESSION_SECRET")))

	http.Handle("/", csrfMiddleware(router))

	srv := &http.Server{
		Handler:      router,
		Addr:         BIND_ADDRESS,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Printf("Starting server: %s", BIND_ADDRESS)
	log.Fatal(srv.ListenAndServe())
}

// DATABASE CONNECTION ===================================

type QueryLogger struct {
	db     *sqlx.DB
	logger *log.Logger
}

func (p QueryLogger) ExecContext(ctx context.Context, q string, args ...interface{}) (sql.Result, error) {
	r, err := p.db.ExecContext(ctx, q, args...)
	a, _ := r.RowsAffected()

	p.logger.Print("DB Exec:", strings.ReplaceAll(q, "\n", " "))
	p.logger.Print(args...)
	p.logger.Printf("RowsAffected: %d", a)
	return r, err
}
func (p QueryLogger) PrepareContext(ctx context.Context, q string) (*sql.Stmt, error) {
	return p.db.PrepareContext(ctx, q)
}
func (p QueryLogger) QueryContext(ctx context.Context, q string, args ...interface{}) (*sql.Rows, error) {
	r, err := p.db.QueryContext(ctx, q, args...)

	p.logger.Print("DB Query: ", strings.ReplaceAll(q, "\n", " "))
	p.logger.Print(args...)
	if err != nil {
		p.logger.Printf("Error: %s", err.Error())
	}

	return r, err
}
func (p QueryLogger) QueryRowContext(ctx context.Context, q string, args ...interface{}) *sql.Row {
	r := p.db.QueryRowContext(ctx, q, args...)

	p.logger.Print("Query Row:", strings.ReplaceAll(q, "\n", " "))
	p.logger.Print(args...)
	err := r.Err()
	if err != nil {
		p.logger.Printf("Error: %s", err.Error())
	}

	return r
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

// ROUTES HELPERS ==========================================
type HandlerFunc func(http.ResponseWriter, *http.Request) http.HandlerFunc

func handlerFuncToHttpHandler(handler HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler(w, r)(w, r)
	}
}

func NotFound(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "", http.StatusNotFound)
}

func BadRequest(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "", http.StatusBadRequest)
}

func Unauthorized(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "", http.StatusBadRequest)
}

func InternalServerError(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "", http.StatusBadRequest)
}

func Redirect(url string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, url, http.StatusFound)
	}
}

func GET(path string, handler HandlerFunc, middlewares ...func(http.HandlerFunc) http.HandlerFunc) {
	router.Methods("GET").Path(path).HandlerFunc(applyMiddlewares(handlerFuncToHttpHandler(handler), middlewares...))
}

func POST(path string, handler HandlerFunc, middlewares ...func(http.HandlerFunc) http.HandlerFunc) {
	router.Methods("POST").Path(path).HandlerFunc(applyMiddlewares(handlerFuncToHttpHandler(handler), middlewares...))
}

func DELETE(path string, handler HandlerFunc, middlewares ...func(http.HandlerFunc) http.HandlerFunc) {
	router.Methods("DELETE").Path(path).HandlerFunc(applyMiddlewares(handlerFuncToHttpHandler(handler), middlewares...))
}

// VIEWS ====================

//go:embed views
var views embed.FS
var templates *template.Template
var helpers = template.FuncMap{}

func compileViews() {
	templates = template.New("")
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

			template.Must(templates.New(name).Funcs(helpers).Parse(string(c)))
		}

		return nil
	})
}

func partial(path string, data interface{}) string {
	v := templates.Lookup(path)
	if v == nil {
		return "view %s not found"
	}

	w := bytes.NewBufferString("")
	err := v.Execute(w, data)
	if err != nil {
		return "rendering error " + path + " " + err.Error()
	}

	return w.String()
}

func Render(path string, view string, data map[string]interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data["view"] = view
		fmt.Fprint(w, partial(path, data))
	}
}

// SESSION =================================

func SESSION(r *http.Request) *sessions.Session {
	s, _ := session.Get(r, SESSION_COOKIE_NAME)
	return s
}

// MIDDLEWARES =============================

func applyMiddlewares(handler http.HandlerFunc, middlewares ...func(http.HandlerFunc) http.HandlerFunc) http.HandlerFunc {
	for _, h := range middlewares {
		handler = h(handler)
	}
	return handler
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

// CONTEXTS ================================

func DbCtx() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
	return ctx
}

// HELPERS FUNCTIONS ======================

func atoi32(s string) int32 {
	i, _ := strconv.ParseInt(s, 10, 32)
	return int32(i)
}
