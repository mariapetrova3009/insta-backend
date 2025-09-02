-- +goose Up
CREATE TABLE IF NOT EXISTS media (
  id         uuid PRIMARY KEY,
  path       text NOT NULL,
  mime       text NOT NULL,
  size       bigint NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

ALTER TABLE posts
  ADD COLUMN IF NOT EXISTS media_id uuid;


ALTER TABLE posts DROP COLUMN IF EXISTS media_id;
DROP TABLE IF EXISTS media;
