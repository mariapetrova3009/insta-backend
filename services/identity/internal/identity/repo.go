package identity

import (
	"context"
	"database/sql"
	"errors"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // драйвер pgx/stdlib для database/sql
)

type Repo struct {
	DB *sql.DB
}

// модели для чтения/записи
type DBUser struct {
	ID         string
	Email      string
	Username   string
	PassHash   string
	Bio        string
	AvatarPath string
	CreatedAt  time.Time
}

func (r *Repo) CreateUser(ctx context.Context, u DBUser) error {

	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO users (id, email, username, pass_hash, bio, avatar_path)
		VALUES ($1, $2, $3, $4, $5, COALESCE($6, ''))
	`, u.ID, u.Email, u.Username, u.PassHash, u.Bio, u.AvatarPath)
	return err
}

func (r *Repo) GetUserByEmailOrName(ctx context.Context, emailOrName string) (*DBUser, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, email, username, pass_hash, bio, avatar_path, created_at
		FROM users
		WHERE email = $1 OR username = $1
	`, emailOrName)

	var u DBUser
	if err := row.Scan(&u.ID, &u.Email, &u.Username, &u.PassHash, &u.Bio, &u.AvatarPath, &u.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *Repo) GetUserByID(ctx context.Context, id string) (*DBUser, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, email, username, pass_hash, bio, avatar_path, created_at
		FROM users WHERE id = $1
	`, id)
	var u DBUser
	if err := row.Scan(&u.ID, &u.Email, &u.Username, &u.PassHash, &u.Bio, &u.AvatarPath, &u.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}
