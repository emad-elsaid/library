package main

import (
	"context"
	"errors"
	"strconv"
	"strings"
)

func (b Book) User() *User {
	user, _ := queries.User(context.Background(), int64(b.UserID))
	return &user
}

func (u User) Shelves() ([]ShelvesRow, error) {
	return queries.Shelves(context.Background(), u.ID)
}

func (n NewBookParams) Validate() (res ValidationErrors) {
	res = ValidationErrors{}

	if !StringPresent(n.Title) {
		res.Add("title", errors.New("Title can't be empty"))
	}

	if !StringPresent(n.Author) {
		res.Add("author", errors.New("Author name can't be empty"))
	}

	if !StringNumeric(n.Isbn) {
		res.Add("isbn", errors.New("ISBN has to consist of numbers"))
	}

	if !StringLength(n.GoogleBooksID.String, 0, 30) {
		res.Add("google_books_id", errors.New("Google Books ID has to be betwee 0 and 30"))
	}

	if !ISBN13(n.Isbn) {
		res.Add("isbn", errors.New("ISBN is not valid"))
	}

	if n.UserID == 0 {
		res.Add("user", errors.New("User has to be set for book"))
	}

	return
}

// VALIDATIONS ===================

type ValidationErrors map[string][]error

func (v ValidationErrors) Add(field string, err error) {
	v[field] = append(v[field], err)
}

func StringPresent(s string) bool {
	return len(strings.TrimSpace(s)) > 0
}

func StringLength(s string, min, max int) bool {
	l := len(s)
	return l >= min && l <= max
}

func StringNumeric(s string) bool {
	for _, c := range s {
		if !strings.ContainsRune("0123456789", c) {
			return false
		}
	}

	return true
}

func ISBN13(s string) bool {
	sum := 0
	for i, s := range s {
		digit, _ := strconv.Atoi(string(s))
		if i%2 == 0 {
			sum += digit
		} else {
			sum += digit * 3
		}
	}

	return sum%10 == 0
}
