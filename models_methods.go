package main

import (
	"context"
)

func (u *User) Books() []Book {
	books, _ := queries.UserBooks(context.Background(), int32(u.ID))

	return books
}

func (b Book) User() *User {
	user, _ := queries.User(context.Background(), int64(b.UserID))
	return &user
}
