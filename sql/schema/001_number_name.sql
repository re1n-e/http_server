-- +goose Up
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    email TEXT UNIQUE NOT NULL,
    hashed_password TEXT NOT NULL DEFAULT 'unset'
);

-- +goose Down
DROP TABLE users;
