-- +goose Up

-- Подписки: кто на кого подписан
CREATE TABLE IF NOT EXISTS follows (
  follower_id uuid NOT NULL,
  followee_id uuid NOT NULL,
  created_at  timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (follower_id, followee_id)
);
-- Быстрый поиск подписчиков пользователя
CREATE INDEX IF NOT EXISTS idx_follows_followee ON follows (followee_id);

-- Материализованная лента: какие посты у какого пользователя в ленте
CREATE TABLE IF NOT EXISTS feed_entries (
  user_id    uuid NOT NULL,
  post_id    uuid NOT NULL,
  created_at timestamptz NOT NULL,
  PRIMARY KEY (user_id, post_id)
);
-- Быстрая выдача ленты по user_id + пагинация по дате
CREATE INDEX IF NOT EXISTS idx_feed_entries_user_created
  ON feed_entries (user_id, created_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_feed_entries_user_created;
DROP TABLE IF EXISTS feed_entries;

DROP INDEX IF EXISTS idx_follows_followee;
DROP TABLE IF EXISTS follows;
