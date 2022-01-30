-- name: User :one
SELECT *
  FROM users
 WHERE id = $1
 LIMIT 1;

-- name: Signup :one
INSERT
 INTO public.users(name, image, slug, email)
VALUES($1,$2,$3,$4)
       ON CONFLICT (email)
       DO UPDATE SET name = $1, image = $2
       RETURNING id;
