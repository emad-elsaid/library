package main

import (
	"context"
	"database/sql"
)

func (u *User) Books() []Book {
	books, _ := queries.UsreBooks(context.Background(), sql.NullInt32{
		Int32: int32(u.ID),
		Valid: true,
	})

	return books
}
