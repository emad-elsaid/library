package main

import "net/http"

func loggedinMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !loggedin(r) {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		next(w, r)
	}
}
