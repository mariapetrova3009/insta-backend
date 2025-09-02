-- +goose Up
CREATE TABLE IF NOT EXISTS follows (
  follower_id uuid NOT NULL,
  followee_id uuid NOT NULL,
  created_at  timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (follower_id, followee_id)
);
CREATE INDEX IF NOT EXISTS idx_follows_followee ON follows (followee_id);

CREATE TABLE IF NOT EXISTS feed_entries (
  user_id    uuid NOT NULL,
  post_id    uuid NOT NULL,
  created_at timestamptz NOT NULL,
  PRIMARY KEY (user_id, post_id)
);
CREATE INDEX IF NOT EXISTS idx_feed_entries_user_created
  ON feed_entries (user_id, created_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_feed_entries_user_created;
DROP TABLE IF EXISTS feed_entries;
DROP INDEX IF EXISTS idx_follows_followee;
DROP TABLE IF EXISTS follows;
