package main

import (
	"context"
	"log"
	"net/http"

	"os"

	"github.com/emad-elsaid/library/models"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	db, err := pgxpool.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer db.Close()

	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Logger())

	e.GET("/", func(c echo.Context) error {
		user, err := models.UserByID(db, 1)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
		}

		return c.String(http.StatusOK, user.Name)
	})

	e.Logger.Fatal(e.Start(":3000"))
}
