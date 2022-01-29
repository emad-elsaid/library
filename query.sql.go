// Code generated by sqlc. DO NOT EDIT.
// source: query.sql

package main

import (
	"context"
)

const getUser = `-- name: GetUser :one
SELECT id, name, email, image, created_at, updated_at, slug, description, facebook, twitter, linkedin, instagram, phone, whatsapp, telegram, amazon_associates_id
  FROM users
 WHERE id = $1
 LIMIT 1
`

func (q *Queries) GetUser(ctx context.Context, id int64) (User, error) {
	row := q.db.QueryRowContext(ctx, getUser, id)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Email,
		&i.Image,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Slug,
		&i.Description,
		&i.Facebook,
		&i.Twitter,
		&i.Linkedin,
		&i.Instagram,
		&i.Phone,
		&i.Whatsapp,
		&i.Telegram,
		&i.AmazonAssociatesID,
	)
	return i, err
}
