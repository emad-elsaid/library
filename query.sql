-- name: User :one
SELECT * FROM users WHERE id = $1 LIMIT 1;

-- name: UserBySlug :one
SELECT * FROM users WHERE slug = $1 LIMIT 1;

-- name: Signup :one
INSERT
 INTO users(name, image, slug, email)
VALUES($1,$2,$3,$4)
       ON CONFLICT (email)
       DO UPDATE SET name = $1, image = $2, updated_at = CURRENT_TIMESTAMP
       RETURNING id;

-- name: UserUnshelvedBooks :many
SELECT books.id id, title, books.image image, google_books_id, slug, isbn
  FROM books, users
 WHERE users.id = books.user_id
   AND user_id = $1
   AND shelf_id IS NULL;

-- name: Shelves :many
SELECT * FROM shelves WHERE user_id = $1 ORDER BY position;

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
SELECT * FROM highlights WHERE book_id = $1 ORDER BY page;

-- name: NewBook :one
INSERT INTO books (title, isbn, author, subtitle, description, publisher, page_count, google_books_id, user_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
       RETURNING *;

-- name: UpdateBook :exec
UPDATE books
   SET title = $1,
       author = $2,
       subtitle = $3,
       description = $4,
       publisher = $5,
       page_count = $6,
       updated_at = CURRENT_TIMESTAMP
 WHERE id = $7;

-- name: UpdateBookImage :exec
UPDATE books SET image = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2;

-- name: ShelfByIdAndUser :one
SELECT * FROM shelves WHERE user_id = $1 AND id = $2 LIMIT 1;

-- name: HighlightByIDAndBook :one
SELECT * FROM highlights WHERE id = $1 AND book_id = $2 LIMIT 1;

-- name: UpdateUser :exec
UPDATE users
   SET description = $1,
       amazon_associates_id = $2,
       facebook = $3,
       twitter = $4,
       linkedin = $5,
       instagram = $6,
       phone = $7,
       whatsapp = $8,
       telegram = $9,
       updated_at = CURRENT_TIMESTAMP
 WHERE id = $10;

-- name: NewHighlight :one
INSERT INTO highlights (book_id, page, content) VALUES ($1, $2, $3) RETURNING *;

-- name: UpdateHighlightImage :exec
UPDATE highlights SET image = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2;

-- name: UpdateHighlight :exec
UPDATE highlights SET page = $1, content = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $3;

-- name: NewShelf :exec
INSERT INTO shelves (name, user_id, position)
VALUES ($1, $2, (
  SELECT coalesce(MAX(position), 0) + 1
    FROM shelves
   WHERE user_id = $2)
);

-- name: UpdateShelf :exec
UPDATE shelves SET name = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2;

-- name: DeleteBook :exec
DELETE FROM books WHERE id = $1;

-- name: DeleteHighlight :exec
DELETE FROM highlights WHERE id = $1;

-- name: HighlightsWithImages :many
SELECT image FROM highlights WHERE image IS NOT NULL AND length(image) > 0 AND book_id = $1;

-- name: RemoveShelf :exec
UPDATE shelves SET position = position - 1
 WHERE user_id = (SELECT user_id FROM shelves WHERE shelves.id = $1)
   AND position > (SELECT position FROM shelves WHERE shelves.id = $1);

-- name: DeleteShelf :exec
DELETE FROM shelves WHERE id = $1;

-- name: MoveShelfUp :exec
UPDATE shelves
   SET position = (
     CASE
     WHEN position = (SELECT position -1 FROM shelves WHERE shelves.id = $1) THEN position + 1
     WHEN position = (SELECT position FROM shelves WHERE shelves.id = $1) THEN position - 1
     END
   )
 WHERE user_id = (SELECT user_id FROM shelves WHERE shelves.id = $1)
   AND position IN (
     (SELECT position -1 FROM shelves WHERE shelves.id = $1),
     (SELECT position FROM shelves WHERE shelves.id = $1)
   );

-- name: MoveShelfDown :exec
UPDATE shelves
   SET position = (
     CASE
     WHEN position = (SELECT position FROM shelves WHERE shelves.id = $1) THEN position + 1
     WHEN position = (SELECT position + 1 FROM shelves WHERE shelves.id = $1) THEN position - 1
     END
   )
 WHERE user_id = (SELECT user_id FROM shelves WHERE shelves.id = $1)
   AND position IN (
     (SELECT position FROM shelves WHERE shelves.id = $1),
     (SELECT position + 1 FROM shelves WHERE shelves.id = $1)
   );

-- name: MoveBookToShelf :exec
UPDATE books SET shelf_id = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2;
