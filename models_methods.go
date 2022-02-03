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
		res.Add("google_books_id", errors.New("Google Books ID has to be between 0 and 30 characters"))
	}

	if !ISBN13(n.Isbn) {
		res.Add("isbn", errors.New("ISBN is not valid"))
	}

	if n.UserID == 0 {
		res.Add("user", errors.New("User has to be set for book"))
	}

	return
}

func (u UpdateUserParams) Validate() (res ValidationErrors) {
	res = ValidationErrors{}

	if !StringLength(u.Description.String, 0, 500) {
		res.Add("description", errors.New("Description has to be between 0 and 500 characters"))
	}

	if !StringLength(u.AmazonAssociatesID.String, 0, 50) {
		res.Add("amazon_associates_id", errors.New("Amazon Associates ID has to be between 0 and 50 characters"))
	}

	if !StringLength(u.Facebook.String, 0, 50) {
		res.Add("facebook", errors.New("Facebook has to be between 0 and 50 characters"))
	}

	if !StringLength(u.Twitter.String, 0, 50) {
		res.Add("twitter", errors.New("Twitter has to be between 0 and 50 characters"))
	}

	if !StringLength(u.Linkedin.String, 0, 50) {
		res.Add("linkedin", errors.New("Linkedin has to be between 0 and 50 characters"))
	}

	if !StringLength(u.Instagram.String, 0, 50) {
		res.Add("instagram", errors.New("Instagram has to be between 0 and 50 characters"))
	}

	if !StringLength(u.Phone.String, 0, 50) {
		res.Add("phone", errors.New("Phone has to be between 0 and 50 characters"))
	}

	if !StringLength(u.Whatsapp.String, 0, 50) {
		res.Add("whatsapp", errors.New("Whatsapp has to be between 0 and 50 characters"))
	}

	if !StringLength(u.Telegram.String, 0, 50) {
		res.Add("telegram", errors.New("Telegram has to be between 0 and 50 characters"))
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
