package feed

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	cmpb "github.com/mariapetrova3009/insta-backend/proto/common"
	fdpb "github.com/mariapetrova3009/insta-backend/proto/feed"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type item struct {
	userID string
	post   *cmpb.Post
}

// gRPC server realisation
type Server struct {
	fdpb.UnimplementedFeedServiceServer
	log *slog.Logger

	repo *Repo
}

func New(log *slog.Logger, repo *Repo) *Server {
	return &Server{
		log:  log,
		repo: repo,
	}
}

func (s *Server) AddPost(authorID string, post *cmpb.Post) {
	if post == nil {
		return
	}

	ctx := context.Background()
	_ = s.repo.FanoutPost(ctx, authorID, post.Id, time.Now().UTC())
}

func (s *Server) GetFeed(ctx context.Context, req *fdpb.GetFeedRequest) (*fdpb.GetFeedResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	// 1) limit
	limit := uint32(20)
	if req.Page != nil && req.Page.Limit > 0 {
		limit = req.Page.Limit
		if limit > 100 {
			limit = 100
		}
	}

	// 2) cursor -> offset
	offset := 0
	if req.Page != nil && req.Page.Cursor != nil {
		if tok := strings.TrimSpace(req.Page.Cursor.Token); tok != "" {
			o, err := decodeCursor(tok) // уже есть у тебя
			if err != nil {
				return nil, status.Error(codes.InvalidArgument, "bad cursor")
			}
			if o > 0 {
				offset = o
			}
		}
	}

	rows, err := s.repo.GetFeed(ctx, "", limit, offset)
	if err != nil {
		return nil, status.Error(codes.Internal, "db error")
	}

	entries := make([]*fdpb.FeedEntry, 0, len(rows))
	for _, r := range rows {
		entries = append(entries, &fdpb.FeedEntry{
			UserId: r.UserID,

			Post: &cmpb.Post{Id: r.PostID},
		})
	}

	var next *cmpb.Cursor
	if uint32(len(rows)) == limit {
		nextOff := offset + int(limit)
		next = &cmpb.Cursor{Token: encodeCursor(nextOff)}
	}

	return &fdpb.GetFeedResponse{
		Entries: entries,
		PageInfo: &cmpb.PageInfo{
			HasMore:    next != nil,
			NextCursor: next,
		},
	}, nil
}

func encodeCursor(offset int) string {
	s := fmt.Sprintf("o:%d", offset)
	return base64.RawURLEncoding.EncodeToString([]byte(s))
}

func decodeCursor(cur string) (int, error) {
	b, err := base64.RawURLEncoding.DecodeString(cur)
	if err != nil {
		return 0, err
	}
	s := string(b)
	if !strings.HasPrefix(s, "o") {
		return 0, fmt.Errorf("bad prefix")
	}

	return strconv.Atoi(s[2:])
}
