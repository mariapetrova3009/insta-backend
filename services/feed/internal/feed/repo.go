package feed

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Repo struct {
	DB *sql.DB
}

func NewRepo(db *sql.DB) *Repo {
	return &Repo{DB: db}
}

type DBFeed struct {
	userId    string
	postId    string
	createdAt time.Time
}

func (r *Repo) FanoutPost(ctx context.Context, authorID, postID string, createdAt time.Time) error {
	// transaction
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// получаем подписчика
	q1 := `SELECT follower_id FROM follows WHERE folowee_id = $1`
	rows, err := tx.QueryContext(ctx, q1, authorID)
	if err != nil {
		return err
	}
	defer rows.Close()

	// вставляем подписчика и указатель на пост
	q2 := `INSERT INTO feed_entries (user_id, post_id, created_at)
	VALUES ($1, $2, $3)
	ON CONFLICT (user_id, post_id) DO NOTHING`
	stmt, err := tx.PrepareContext(ctx, q2)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for rows.Next() {
		var follower string
		if err := rows.Scan(&follower); err != nil {
			return err
		}
		if _, err := stmt.ExecContext(ctx, follower, postID, createdAt); err != nil {
			return err
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return tx.Commit()

}

type EntryLow struct {
	UserID    string
	PostID    string
	CrearedAt time.Time
}

func (r *Repo) GetFeed(ctx context.Context, userID string, limit uint32, offset int) ([]EntryLow, error) {
	q := `SELECT * FROM feed_entries %s ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	args := []any{limit, offset}

	where := ""
	if userID != "" {
		where = "WHERE user_id = $3"
		args = []any{limit, offset, userID}
	}
	rows, err := r.DB.QueryContext(ctx, fmt.Sprintf(q, where), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]EntryLow, 0, limit)
	for rows.Next() {
		var e EntryLow
		if err := rows.Scan(&e.UserID, &e.PostID, &e.CrearedAt); err != nil {
			return nil, err
		}
		out = append(out, e)

	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
