-- +goose Up
CREATE TABLE users (
  id UUID PRIMARY KEY,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  email TEXT UNIQUE NOT NULL
);

-- +goose Down
DROP TABLE users;
