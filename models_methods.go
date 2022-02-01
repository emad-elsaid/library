package main

import (
	"context"
	"errors"
)

func (b Book) User() *User {
	user, _ := queries.User(context.Background(), int64(b.UserID))
	return &user
}

func (u User) Shelves() ([]ShelvesRow, error) {
	return queries.Shelves(context.Background(), u.ID)
}

func (n NewBookParams) Validate() (res map[string][]error) {
	res = map[string][]error{}

	if len(n.Title) == 0 {
		res["title"] = append(res["title"], errors.New("Title can't be empty"))
	}

	if len(n.Author) == 0 {
		res["author"] = append(res["author"], errors.New("Author name can't be empty"))
	}

	if len(n.Isbn) != 13 {
		res["isbn"] = append(res["isbn"], errors.New("ISBN has to be 13 digits"))
	}

	return
}
