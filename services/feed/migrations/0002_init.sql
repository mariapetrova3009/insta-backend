-- +goose Up
CREATE TABLE IF NOT EXISTS posts (
  id              uuid PRIMARY KEY,
  author_id       uuid NOT NULL,
  caption         text,
  media_path      text NOT NULL,
  mime            text NOT NULL,
  likes_count     int  NOT NULL DEFAULT 0,
  comments_count  int  NOT NULL DEFAULT 0,
  created_at      timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_posts_author_created
  ON posts (author_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_posts_created
  ON posts (created_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_posts_created;
DROP INDEX IF EXISTS idx_posts_author_created;
DROP TABLE IF EXISTS posts;
