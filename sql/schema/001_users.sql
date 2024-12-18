-- +goose Up
CREATE TABLE users(id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
created_at TIMESTAMPTZ DEFAULT now(),
updated_at TIMESTAMPTZ DEFAULT now(),
email TEXT not null unique);

-- +goose Down
DROP TABLE users;





-- "postgres://mohamadhishmeh:@localhost:5432/chirpy"  this is connection string
