package main

import (
	"context"
	"database/sql"
)

func (u *User) Books() []Book {
	books, _ := queries.UserBooks(context.Background(), sql.NullInt32{
		Int32: int32(u.ID),
		Valid: true,
	})

	return books
}

func (b Book) User() *User {
	user, _ := queries.User(context.Background(), int64(b.UserID.Int32))
	return &user
}
