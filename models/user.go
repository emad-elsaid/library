package models

import (
	"context"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

var ctx = context.Background()

var ALLOWED_IMAGES_TYPES = [...]string{"gif", "png", "jpeg"}

type User struct {
	Name        string
	Email       string
	Image       string
	Slug        string
	Description string
	Facebook    string
	Twitter     string
	Linkedin    string
	Instagram   string
	Phone       string
	Whatsapp    string
	Telegram    string

	CreatedAt time.Time
	UpdatedAt time.Time
}

func UserByID(db *pgxpool.Pool, id int64) (user User, err error) {
	return user, db.
		QueryRow(ctx, "SELECT name FROM users WHERE id=$1", id).
		Scan(&user.Name)
}

type Book struct {
	ShelfID int64
	UserID  int64

	Title         string
	Subtitle      string
	Description   string
	Author        string
	Image         string
	ISBN          string
	PageCount     int64
	Publisher     string
	GoogleBooksID string

	CreatedAt time.Time
	UpdatedAt time.Time
}

type Shelf struct {
	UserID int64

	Name     string
	Position int64

	CreatedAt time.Time
	UpdatedAt time.Time
}

type Borrow struct {
	UserID  int64
	BookID  int64
	OwnerID int64
	Days    int64

	BorrowedAt time.Time
	ReturnedAt time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type Access struct {
	UserID  int64
	OwnerID int64

	AcceptedAt int64
	RejectedAt int64

	CreatedAt time.Time
	UpdatedAt time.Time
}

type Email struct {
	UserID        int64
	EmailableType string
	EmailableID   int64

	About string

	CreatedAt time.Time
	UpdatedAt time.Time
}
