package repo

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Media struct {
	ID        uuid.UUID
	Path      string
	Mime      string
	Size      int64
	CreatedAt time.Time
}

type Post struct {
	ID        uuid.UUID
	AuthorID  uuid.UUID
	Caption   string
	Media     Media
	CreatedAt time.Time
}

type Repo struct {
	DB *sql.DB
}

func NewRepo(db *sql.DB) *Repo {
	return &Repo{DB: db}
}

func (r *Repo) CreateMedia(ctx context.Context, m *Media) error {
	stmt, err := r.DB.PrepareContext(ctx, `INSERT INTO media (id, path, mime, size, created_at) 
	VALUES ($1, $2, $3, $4, $5)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.ExecContext(ctx, m.ID, m.Path, m.Mime, m.Size, m.CreatedAt)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repo) CreatePost(ctx context.Context, p *Post) error {
	stmt, err := r.DB.PrepareContext(ctx, `insert into posts(id, author_id, caption, media_id, created_at) values($1,$2,$3,$4,$5)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.ExecContext(ctx, p.ID, p.AuthorID, p.Caption, p.Media.ID, p.CreatedAt)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repo) GetPost(ctx context.Context, id uuid.UUID) (*Post, error) {
	stmt, err := r.DB.PrepareContext(ctx,
		`SELECT p.id, p.author_id, p.caption, p.created_at, 
		m.id, m.path, m.mime, m.size, m.created_at 
	FROM posts p JOIN media m 
	ON m.id = p.media_id
	where p.id = $1`)
	var p Post
	p.Media = Media{}
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	row := stmt.QueryRowContext(ctx, id)
	if err := row.Scan(&p.ID, &p.AuthorID, &p.Caption, &p.CreatedAt,
		&p.Media.ID, &p.Media.Path, &p.Media.Mime, &p.Media.Size, &p.Media.CreatedAt); err != nil {
		return nil, err

	}

	return &p, nil
}
