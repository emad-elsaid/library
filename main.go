package main

import (
	"fmt"
	"net/http"
)

func main() {
	router.Methods("GET").Path("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, render("layout", "index", map[string]interface{}{
			"meta": map[string]string{},
		}))
	})

	router.Methods("POST").Path("/auth/google").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	router.Methods("POST").Path("/auth/google/callback").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	router.Methods("GET").Path("/logout").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	router.Methods("GET").Path("/users/{user}").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	router.Methods("GET").Path("/users/{user}/edit").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	router.Methods("POST").Path("/users/{user}").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	router.Methods("GET").Path("/users/{user}/books/new").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	router.Methods("POST").Path("/users/{user}/books").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	router.Methods("GET").Path("/users/{user}/books/{isbn}").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	router.Methods("POST").Path("/users/{user}/books/{isbn}").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	router.Methods("GET").Path("/users/{user}/books/{isbn}/edit").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	router.Methods("DELETE").Path("/users/{user}/books/{isbn}").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	router.Methods("POST").Path("/users/{user}/books/{isbn}/shelf").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	router.Methods("GET").Path("/users/{user}/books/{isbn}/image").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	router.Methods("POST").Path("/users/{user}/books/{isbn}/image").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	router.Methods("GET").Path("/users/{user}/shelves/new").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	router.Methods("GET").Path("/users/{user}/shelves").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	router.Methods("POST").Path("/users/{user}/shelves").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	router.Methods("GET").Path("/users/{user}/shelves/{shelf}/edit").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	router.Methods("POST").Path("/users/{user}/shelves/{shelf}").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	router.Methods("POST").Path("/users/{user}/shelves/{shelf}/up").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	router.Methods("POST").Path("/users/{user}/shelves/{shelf}/down").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	router.Methods("DELETE").Path("/users/{user}/shelves/{shelf}").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	Helpers()
	Start()
}
