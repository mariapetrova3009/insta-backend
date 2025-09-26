package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"path/filepath"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/mariapetrova3009/insta-backend/services/content/internal/repo"
	"github.com/mariapetrova3009/insta-backend/services/content/internal/storage"
	"google.golang.org/grpc/metadata"

	"github.com/google/uuid"
	commonpb "github.com/mariapetrova3009/insta-backend/proto/common"
	contentpb "github.com/mariapetrova3009/insta-backend/proto/content"
)

type Server struct {
	contentpb.UnimplementedContentServiceServer
	log              *slog.Logger
	repo             *repo.Repo
	store            storage.Storage
	prod             *kafka.Producer
	topicPostCreated string
}

func New(log *slog.Logger, repo *repo.Repo, store storage.Storage, prod *kafka.Producer, topicPostCreated string) *Server {
	return &Server{log: log, repo: repo, store: store, prod: prod, topicPostCreated: topicPostCreated}
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

	var authorID uuid.UUID
	{
		// ключ "user-id" замени на тот, что реально прокидывает твой gateway/JWT-мидлварь
		md, _ := metadata.FromIncomingContext(ctx)
		if vals := md.Get("user-id"); len(vals) > 0 {
			if aid, err := uuid.Parse(vals[0]); err == nil {
				authorID = aid
			} else {
				s.log.Warn("invalid user-id in metadata", "value", vals[0], "err", err)
			}
		}
		if authorID == uuid.Nil {
			s.log.Warn("author_id is empty; feed may not be able to attribute the post")

		}
	}

	mediaID := uuid.MustParse(filepath.Base(filepath.Dir(in.GetMediaPath())))
	media := repo.Media{
		ID: mediaID,
	}
	id := uuid.New()
	p := repo.Post{
		ID:        id,
		AuthorID:  authorID,
		Caption:   in.Caption,
		Media:     media,
		CreatedAt: time.Now().UTC(),
	}
	if err := s.repo.CreatePost(ctx, &p); err != nil {
		return nil, err
	}

	// TODO: Kafka

	// событие
	evt := struct {
		PostID      string `json:"post_id"`
		AuthorID    string `json:"author_id,omitempty"`
		Caption     string `json:"caption"`
		MediaPath   string `json:"media_path"`
		Mime        string `json:"mime"`
		CreatedAtMs int64  `json:"created_at_ms"`
	}{
		PostID:      p.ID.String(),
		AuthorID:    authorID.String(),
		Caption:     p.Caption,
		MediaPath:   in.GetMediaPath(),
		Mime:        in.GetMime(),
		CreatedAtMs: time.Now().UTC().UnixMilli(),
	}

	payload, err := json.Marshal(evt)
	if err != nil {
		s.log.Error("marshal event failed", "err", err)
	} else {
		key := []byte(p.ID.String())
		topic := s.topicPostCreated

		// отправка в Kafka
		err = s.prod.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     &topic,
				Partition: kafka.PartitionAny,
			},
			Key:   key,
			Value: payload,
			Headers: []kafka.Header{
				{Key: "schema", Value: []byte("content.post.created.v1")},
				{Key: "content-type", Value: []byte("application/json")},
			},
		}, nil)

		if err != nil {
			s.log.Error("kafka produce failed", "err", err)
		}
	}

	//

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
