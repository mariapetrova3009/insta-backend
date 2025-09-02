package server

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"time"

	"github.com/mariapetrova3009/insta-backend/services/content/internal/repo"
	"github.com/mariapetrova3009/insta-backend/services/content/internal/storage"

	"github.com/google/uuid"
	commonpb "github.com/mariapetrova3009/insta-backend/proto/common"
	contentpb "github.com/mariapetrova3009/insta-backend/proto/content"
)

type Server struct {
	contentpb.UnimplementedContentServiceServer
	log   *slog.Logger
	repo  *repo.Repo
	store storage.Storage
}

func New(log *slog.Logger, repo *repo.Repo) *Server {
	return &Server{log: log, repo: repo}
}

func (s *Server) UploadMedia(ctx context.Context, in *contentpb.UploadMediaRequest) (*contentpb.UploadMediaResponse, error) {
	// uniq id
	id := uuid.New()
	keyName := filepath.Join(id.String(), in.GetName())

	res, err := s.store.Put(keyName, in.Data, in.Mime)
	if err != nil {
		s.log.Error("upload failed", "err", err)
		return nil, fmt.Errorf("failed to store file: %w", err)
	}

	media := &repo.Media{
		ID:        id,
		Path:      res.Key, // путь в хранилище (может быть URL или просто key)
		Mime:      in.GetMime(),
		Size:      res.Size,
		CreatedAt: time.Now().UTC(),
	}

	if err := s.repo.CreateMedia(ctx, media); err != nil {
		s.log.Error("create media failed", "err", err)
		return nil, fmt.Errorf("failed to save metadata: %w", err)
	}

	return &contentpb.UploadMediaResponse{
		MediaPath: res.Key,
		Name:      in.GetName(),
		Mime:      in.GetMime(),
	}, nil
}

func (s *Server) CreatePost(ctx context.Context, in *contentpb.CreatePostRequest) (*contentpb.PostResponse, error) {

	mediaID := uuid.MustParse(filepath.Base(filepath.Dir(in.GetMediaPath())))
	media := repo.Media{
		ID: mediaID,
	}
	id := uuid.New()
	p := repo.Post{
		ID:        id,
		Caption:   in.Caption,
		Media:     media,
		CreatedAt: time.Now().UTC(),
	}
	if err := s.repo.CreatePost(ctx, &p); err != nil {
		return nil, err
	}
	post := &commonpb.Post{
		Id:        p.ID.String(),
		Caption:   p.Caption,
		MediaPath: in.GetMediaPath(),
		Mime:      in.GetMime(),
	}
	return &contentpb.PostResponse{Post: post}, nil
}

func (s *Server) GetPost(ctx context.Context, in *contentpb.GetPostRequest) (*contentpb.PostResponse, error) {
	pid, err := uuid.Parse(in.GetPostId())
	if err != nil {
		return nil, err
	}
	p, err := s.repo.GetPost(ctx, pid)
	if err != nil {
		return nil, err
	}
	post := &commonpb.Post{
		Id:        p.ID.String(),
		Caption:   p.Caption,
		MediaPath: p.Media.Path,
		Mime:      p.Media.Mime,
	}
	return &contentpb.PostResponse{Post: post}, nil

}
