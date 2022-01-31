package main

import (
	"context"
)

func (b Book) User() *User {
	user, _ := queries.User(context.Background(), int64(b.UserID))
	return &user
}
