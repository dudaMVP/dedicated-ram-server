-- name: GetSingleChirp :one

SELECT * from chirps 
where id = $1;