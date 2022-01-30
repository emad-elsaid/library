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
			Name:  sql.NullString{String: user.Name},
			Image: sql.NullString{String: user.Picture},
			Slug:  uuid.New().String(),
			Email: sql.NullString{String: user.Email},
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

	GET("/logout", func(w http.ResponseWriter, r *http.Request) {})

	GET("/users/{user}", func(w http.ResponseWriter, r *http.Request) {})
	GET("/users/{user}/edit", func(w http.ResponseWriter, r *http.Request) {})
	POST("/users/{user}", func(w http.ResponseWriter, r *http.Request) {})

	GET("/users/{user}/books/new", func(w http.ResponseWriter, r *http.Request) {})
	POST("/users/{user}/books", func(w http.ResponseWriter, r *http.Request) {})
	GET("/users/{user}/books/{isbn}", func(w http.ResponseWriter, r *http.Request) {})
	POST("/users/{user}/books/{isbn}", func(w http.ResponseWriter, r *http.Request) {})
	GET("/users/{user}/books/{isbn}/edit", func(w http.ResponseWriter, r *http.Request) {})
	DELETE("/users/{user}/books/{isbn}", func(w http.ResponseWriter, r *http.Request) {})

	POST("/users/{user}/books/{isbn}/shelf", func(w http.ResponseWriter, r *http.Request) {})

	GET("/users/{user}/books/{isbn}/image", func(w http.ResponseWriter, r *http.Request) {})
	POST("/users/{user}/books/{isbn}/image", func(w http.ResponseWriter, r *http.Request) {})

	GET("/users/{user}/shelves/new", func(w http.ResponseWriter, r *http.Request) {})
	GET("/users/{user}/shelves", func(w http.ResponseWriter, r *http.Request) {})
	POST("/users/{user}/shelves", func(w http.ResponseWriter, r *http.Request) {})
	GET("/users/{user}/shelves/{shelf}/edit", func(w http.ResponseWriter, r *http.Request) {})
	POST("/users/{user}/shelves/{shelf}", func(w http.ResponseWriter, r *http.Request) {})
	POST("/users/{user}/shelves/{shelf}/up", func(w http.ResponseWriter, r *http.Request) {})
	POST("/users/{user}/shelves/{shelf}/down", func(w http.ResponseWriter, r *http.Request) {})
	DELETE("/users/{user}/shelves/{shelf}", func(w http.ResponseWriter, r *http.Request) {})

	GET("/users/{user}/books/{isbn}/highlights/new", func(w http.ResponseWriter, r *http.Request) {})
	POST("/users/{user}/books/{isbn}/highlights", func(w http.ResponseWriter, r *http.Request) {})
	GET("/users/{user}/books/{isbn}/highlights/{id}/edit", func(w http.ResponseWriter, r *http.Request) {})
	POST("/users/{user}/books/{isbn}/highlights/{id}", func(w http.ResponseWriter, r *http.Request) {})
	DELETE("/users/{user}/books/{isbn}/highlights/{id}", func(w http.ResponseWriter, r *http.Request) {})
	GET("/users/{user}/books/{isbn}/highlights/{id}/image", func(w http.ResponseWriter, r *http.Request) {})
	POST("/users/{user}/books/{isbn}/highlights/{id}/image", func(w http.ResponseWriter, r *http.Request) {})

	Helpers()
	Start()
}
