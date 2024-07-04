-- name: GetPeople :many
SELECT id, name, email, created_at, updated_at, update_user
FROM person;

-- name: GetPersonById :one
SELECT id, name, email, created_at, updated_at, update_user
FROM person
WHERE id = $1;

-- name: UpdatePerson :execrows
UPDATE person SET
  "name" = $2,
  email = $3,
  created_at = $4,
  updated_at = $5,
  update_user = $6
where id = $1;

-- name: InsertPerson :one
INSERT INTO person (name, email, created_at, updated_at, update_user)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: DeletePerson :execrows
DELETE from person
WHERE id = $1;

-- name: PingDb :one
SELECT 1 as Result;