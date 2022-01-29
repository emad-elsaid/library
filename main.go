package main

import (
	"fmt"
	"net/http"
)

func main() {
	GET("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, render("layout", "index", map[string]interface{}{
			"meta": map[string]string{},
		}))
	})

	GET("/privacy", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, render("layout", "privacy", map[string]interface{}{
			"meta": map[string]string{},
		}))
	})

	POST("/auth/google", func(w http.ResponseWriter, r *http.Request) {})
	POST("/auth/google/callback", func(w http.ResponseWriter, r *http.Request) {})
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
