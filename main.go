package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

	GET("/", func(w http.ResponseWriter, r *http.Request) {
		user := current_user(r)
		if user != nil {
			http.Redirect(w, r, fmt.Sprintf("/users/%s", user.Slug), http.StatusFound)
			return
		} else {
			render(w, "layout", "index", map[string]interface{}{
				"meta": map[string]string{},
			})
		}
	})

	GET("/privacy", func(w http.ResponseWriter, r *http.Request) {
		render(w, "layout", "privacy", map[string]interface{}{
			"meta": map[string]string{},
		})
	})

	POST("/auth/google", func(w http.ResponseWriter, r *http.Request) {
		url := google_conf.AuthCodeURL("state")
		http.Redirect(w, r, url, http.StatusFound)
	})

	GET("/auth/google/callback", func(w http.ResponseWriter, r *http.Request) {
		tok, err := google_conf.Exchange(oauth2.NoContext, r.URL.Query().Get("code"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		client := google_conf.Client(oauth2.NoContext, tok)
		resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		user := struct {
			Name    string `json:"name"`
			Email   string `json:"email"`
			Picture string `json:"picture"`
		}{}
		err = json.Unmarshal(body, &user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		u, err := queries.Signup(context.Background(), SignupParams{
			Name:  sql.NullString{String: user.Name, Valid: true},
			Image: sql.NullString{String: user.Picture, Valid: true},
			Slug:  uuid.New().String(),
			Email: sql.NullString{String: user.Email, Valid: true},
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		s := SESSION(r)
		s.Values["current_user"] = u
		if err = s.Save(r, w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusFound)
	})

	GET("/logout", func(w http.ResponseWriter, r *http.Request) {
		s := SESSION(r)
		s.Values = map[interface{}]interface{}{}
		s.Save(r, w)
		http.Redirect(w, r, "/", http.StatusFound)
	})

	GET("/users/{user}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		user, err := queries.UserBySlug(context.Background(), vars["user"])
		if err != nil {
			http.NotFound(w, r)
			return
		}

		render(w, "layout", "users/show", map[string]interface{}{
			"current_user": current_user(r),
			"user":         user,
		})
	})

	GET("/users/{user}/edit", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		user, err := queries.UserBySlug(context.Background(), vars["user"])
		if err != nil {
			http.NotFound(w, r)
			return
		}

		render(w, "layout", "users/edit", map[string]interface{}{
			"current_user": current_user(r),
			"user":         user,
		})
	}, loggedinMiddleware)

	POST("/users/{user}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		user, err := queries.UserBySlug(context.Background(), vars["user"])
		if err != nil {
			http.NotFound(w, r)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/users/%s", user.Slug), http.StatusFound)
	}, loggedinMiddleware)

	GET("/users/{user}/books/new", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		user, err := queries.UserBySlug(context.Background(), vars["user"])
		if err != nil {
			http.NotFound(w, r)
			return
		}

		render(w, "layout", "books/new", map[string]interface{}{
			"current_user": current_user(r),
			"user":         user,
			"csrf":         csrf.TemplateField(r),
		})
	}, loggedinMiddleware)

	POST("/users/{user}/books", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		user, err := queries.UserBySlug(context.Background(), vars["user"])
		if err != nil {
			http.NotFound(w, r)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/users/%s", user.Slug), http.StatusFound)
	})

	GET("/users/{user}/books/{isbn}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		user, err := queries.UserBySlug(context.Background(), vars["user"])
		if err != nil {
			http.NotFound(w, r)
			return
		}

		render(w, "layout", "books/show", map[string]interface{}{
			"current_user": current_user(r),
			"user":         user,
		})
	})

	POST("/users/{user}/books/{isbn}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		user, err := queries.UserBySlug(context.Background(), vars["user"])
		if err != nil {
			http.NotFound(w, r)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/users/%s/books/%s", user.Slug, vars["isbn"]), http.StatusFound)
	}, loggedinMiddleware)

	GET("/users/{user}/books/{isbn}/edit", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		user, err := queries.UserBySlug(context.Background(), vars["user"])
		if err != nil {
			http.NotFound(w, r)
			return
		}

		render(w, "layout", "books/edit", map[string]interface{}{
			"current_user": current_user(r),
			"user":         user,
		})
	}, loggedinMiddleware)

	DELETE("/users/{user}/books/{isbn}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		user, err := queries.UserBySlug(context.Background(), vars["user"])
		if err != nil {
			http.NotFound(w, r)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/users/%s", user.Slug), http.StatusFound)
	}, loggedinMiddleware)

	POST("/users/{user}/books/{isbn}/shelf", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		user, err := queries.UserBySlug(context.Background(), vars["user"])
		if err != nil {
			http.NotFound(w, r)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/users/%s/books/%s", user.Slug, vars["isbn"]), http.StatusFound)
	}, loggedinMiddleware)

	GET("/users/{user}/books/{isbn}/image", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		user, err := queries.UserBySlug(context.Background(), vars["user"])
		if err != nil {
			http.NotFound(w, r)
			return
		}

		render(w, "layout", "books/image", map[string]interface{}{
			"current_user": current_user(r),
			"user":         user,
		})
	}, loggedinMiddleware)

	POST("/users/{user}/books/{isbn}/image", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		user, err := queries.UserBySlug(context.Background(), vars["user"])
		if err != nil {
			http.NotFound(w, r)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/users/%s/books/%s", user.Slug, vars["isbn"]), http.StatusFound)
	}, loggedinMiddleware)

	GET("/users/{user}/shelves/new", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		user, err := queries.UserBySlug(context.Background(), vars["user"])
		if err != nil {
			http.NotFound(w, r)
			return
		}

		render(w, "layout", "shelves/new", map[string]interface{}{
			"current_user": current_user(r),
			"user":         user,
			"csrf":         csrf.TemplateField(r),
		})
	}, loggedinMiddleware)

	GET("/users/{user}/shelves", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		user, err := queries.UserBySlug(context.Background(), vars["user"])
		if err != nil {
			http.NotFound(w, r)
			return
		}

		render(w, "layout", "shelves/index", map[string]interface{}{
			"current_user": current_user(r),
			"user":         user,
		})
	}, loggedinMiddleware)

	POST("/users/{user}/shelves", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		user, err := queries.UserBySlug(context.Background(), vars["user"])
		if err != nil {
			http.NotFound(w, r)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/users/%s/shelves", user.Slug), http.StatusFound)
	}, loggedinMiddleware)

	GET("/users/{user}/shelves/{shelf}/edit", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		user, err := queries.UserBySlug(context.Background(), vars["user"])
		if err != nil {
			http.NotFound(w, r)
			return
		}

		render(w, "layout", "shelves/edit", map[string]interface{}{
			"current_user": current_user(r),
			"user":         user,
			"csrf":         csrf.TemplateField(r),
		})
	}, loggedinMiddleware)

	POST("/users/{user}/shelves/{shelf}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		user, err := queries.UserBySlug(context.Background(), vars["user"])
		if err != nil {
			http.NotFound(w, r)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/users/%s/shelves", user.Slug), http.StatusFound)
	}, loggedinMiddleware)

	POST("/users/{user}/shelves/{shelf}/up", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		user, err := queries.UserBySlug(context.Background(), vars["user"])
		if err != nil {
			http.NotFound(w, r)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/users/%s/shelves", user.Slug), http.StatusFound)
	}, loggedinMiddleware)

	POST("/users/{user}/shelves/{shelf}/down", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		user, err := queries.UserBySlug(context.Background(), vars["user"])
		if err != nil {
			http.NotFound(w, r)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/users/%s/shelves", user.Slug), http.StatusFound)
	}, loggedinMiddleware)

	DELETE("/users/{user}/shelves/{shelf}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		user, err := queries.UserBySlug(context.Background(), vars["user"])
		if err != nil {
			http.NotFound(w, r)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/users/%s/shelves", user.Slug), http.StatusFound)
	}, loggedinMiddleware)

	GET("/users/{user}/books/{isbn}/highlights/new", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		user, err := queries.UserBySlug(context.Background(), vars["user"])
		if err != nil {
			http.NotFound(w, r)
			return
		}

		render(w, "layout", "highlights/new", map[string]interface{}{
			"current_user": current_user(r),
			"user":         user,
			"csrf":         csrf.TemplateField(r),
		})
	}, loggedinMiddleware)

	POST("/users/{user}/books/{isbn}/highlights", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		user, err := queries.UserBySlug(context.Background(), vars["user"])
		if err != nil {
			http.NotFound(w, r)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/users/%s/books/%s", user.Slug, vars["isbn"]), http.StatusFound)
	}, loggedinMiddleware)

	GET("/users/{user}/books/{isbn}/highlights/{id}/edit", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		user, err := queries.UserBySlug(context.Background(), vars["user"])
		if err != nil {
			http.NotFound(w, r)
			return
		}

		render(w, "layout", "highlights/edit", map[string]interface{}{
			"current_user": current_user(r),
			"user":         user,
			"csrf":         csrf.TemplateField(r),
		})
	}, loggedinMiddleware)

	POST("/users/{user}/books/{isbn}/highlights/{id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		user, err := queries.UserBySlug(context.Background(), vars["user"])
		if err != nil {
			http.NotFound(w, r)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/users/%s/books/%s", user.Slug, vars["isbn"]), http.StatusFound)
	}, loggedinMiddleware)

	DELETE("/users/{user}/books/{isbn}/highlights/{id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		user, err := queries.UserBySlug(context.Background(), vars["user"])
		if err != nil {
			http.NotFound(w, r)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/users/%s/books/%s", user.Slug, vars["isbn"]), http.StatusFound)
	}, loggedinMiddleware)

	GET("/users/{user}/books/{isbn}/highlights/{id}/image", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		user, err := queries.UserBySlug(context.Background(), vars["user"])
		if err != nil {
			http.NotFound(w, r)
			return
		}

		render(w, "layout", "highlights/image", map[string]interface{}{
			"current_user": current_user(r),
			"user":         user,
		})
	}, loggedinMiddleware)

	POST("/users/{user}/books/{isbn}/highlights/{id}/image", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		user, err := queries.UserBySlug(context.Background(), vars["user"])
		if err != nil {
			http.NotFound(w, r)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/users/%s/books/%s", user.Slug, vars["isbn"]), http.StatusFound)
	}, loggedinMiddleware)

	Helpers()
	Start()
}
