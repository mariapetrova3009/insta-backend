-- +goose Up
CREATE TABLE IF NOT EXISTS users (
  id           uuid PRIMARY KEY,
  email        text NOT NULL UNIQUE,
  username     text NOT NULL UNIQUE,
  pass_hash    text NOT NULL,
  bio          text DEFAULT '',
  avatar_path  text DEFAULT '',
  created_at   timestamptz NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS users;
