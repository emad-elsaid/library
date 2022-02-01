-- name: User :one
SELECT *
  FROM users
 WHERE id = $1
 LIMIT 1;

-- name: UserBySlug :one
SELECT *
  FROM users
 WHERE slug = $1
 LIMIT 1;

-- name: Signup :one
INSERT
 INTO public.users(name, image, slug, email)
VALUES($1,$2,$3,$4)
       ON CONFLICT (email)
       DO UPDATE SET name = $1, image = $2
       RETURNING id;

-- name: UserUnshelvedBooks :many
SELECT books.id id, title, books.image image, google_books_id, slug, isbn
  FROM books, users
 WHERE users.id = books.user_id
   AND user_id = $1
   AND shelf_id IS NULL;

-- name: Shelves :many
SELECT id, name
  FROM shelves
 WHERE user_id = $1
 ORDER BY position;

-- name: ShelfBooks :many
SELECT books.id id, title, books.image image, google_books_id, slug, isbn
  FROM books, users
 WHERE users.id = books.user_id
   AND shelf_id = $1
 ORDER BY books.created_at DESC;

-- name: BookByIsbnAndUser :one
SELECT books.*, slug, shelves.name shelf_name
  FROM users, books
       LEFT JOIN shelves
           ON shelves.id = books.shelf_id
 WHERE users.id = books.user_id
   AND books.user_id = $1
   AND isbn = $2
 LIMIT 1;

-- name: Highlights :many
SELECT *
  FROM highlights
 WHERE book_id = $1;

-- name: NewBook :one
INSERT INTO public.books (title, isbn, author, subtitle, description, publisher, page_count, google_books_id, user_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
       RETURNING *;
