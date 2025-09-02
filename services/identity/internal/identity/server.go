package identity

import (
	"context"
	"errors"
	"log/slog"
	"net/mail"
	"strings"
	"time"

	cfgpkg "github.com/mariapetrova3009/insta-backend/pkg/config"
	cmpb "github.com/mariapetrova3009/insta-backend/proto/common"
	idpb "github.com/mariapetrova3009/insta-backend/proto/identity"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// внутренняя модель пользователя (для in-memory варианта)

// gRPC server realisation
type Server struct {
	idpb.UnimplementedIdentityServiceServer
	log       *slog.Logger
	jwtSecret []byte
	accessTTL time.Duration
	repo      *Repo
}

func New(log *slog.Logger, cfg *cfgpkg.Config, repo *Repo) *Server {
	return &Server{
		log:       log,
		jwtSecret: []byte(cfg.JWT.Secret),
		accessTTL: cfg.JWT.TTL,
		repo:      repo,
	}
}

// валидация логина
// проверка, что наш токен не существует
// CreateUser в бд
// get access token
func (s *Server) Register(ctx context.Context, req *idpb.RegisterRequest) (*idpb.AuthResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if err := validateRegister(req); err != nil {
		return nil, err
	}

	if u, _ := s.repo.GetUserByEmailOrName(ctx, req.Email); u != nil {
		return nil, status.Error(codes.AlreadyExists, "email already registered")
	}
	if u, _ := s.repo.GetUserByEmailOrName(ctx, req.Username); u != nil {
		return nil, status.Error(codes.AlreadyExists, "username already taken")
	}

	h, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, status.Error(codes.Internal, "hash error")
	}

	id := uuid.NewString()
	// выполнение запроса к бд
	if err := s.repo.CreateUser(ctx, DBUser{
		ID: id, Email: req.Email, Username: req.Username, PassHash: string(h), Bio: req.Bio,
	}); err != nil {
		return nil, status.Error(codes.AlreadyExists, "email or username already taken")
	}

	access, err := s.issueAccessToken(id)
	if err != nil {
		return nil, status.Error(codes.Internal, "token error")
	}

	// читаем пользователя, чтобы отдать created_at
	dbu, _ := s.repo.GetUserByID(ctx, id)
	return &idpb.AuthResponse{
		AccessToken: access,
		User: &cmpb.User{
			Id: dbu.ID, Email: dbu.Email, Username: dbu.Username, Bio: dbu.Bio,
			CreatedAt: timestamppb.New(dbu.CreatedAt),
		},
	}, nil
}

func validateRegister(req *idpb.RegisterRequest) error {

	if req.Email == "" || !isEmail(req.Email) {
		return status.Error(codes.InvalidArgument, "invalid email")
	}
	if len(req.Password) < 6 {
		return status.Error(codes.InvalidArgument, "password too short (min 6)")
	}
	if req.Username == "" {
		return status.Error(codes.InvalidArgument, "username is required")
	}
	return nil
}

// validate login
// check passHash and password
// get access token
func (s *Server) Login(ctx context.Context, req *idpb.LoginRequest) (*idpb.AuthResponse, error) {
	if err := validateLogin(req); err != nil {
		return nil, err
	}

	u, err := s.repo.GetUserByEmailOrName(ctx, req.EmailOrUsername)
	if err != nil || u == nil {
		return nil, status.Error(codes.PermissionDenied, "invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PassHash), []byte(req.Password)); err != nil {
		return nil, status.Error(codes.PermissionDenied, "invalid credentials")
	}

	access, err := s.issueAccessToken(u.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, "token error")
	}

	return &idpb.AuthResponse{
		AccessToken: access,
		User: &cmpb.User{
			Id: u.ID, Email: u.Email, Username: u.Username, Bio: u.Bio,
			CreatedAt: timestamppb.New(u.CreatedAt),
		},
	}, nil
}

func validateLogin(req *idpb.LoginRequest) error {
	if req == nil || req.EmailOrUsername == "" || req.Password == "" {
		return status.Error(codes.InvalidArgument, "email_or_username and password are required")
	}
	return nil
}

// issue the JWT and return it to the client

func (s *Server) issueAccessToken(userID string) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   userID,
		ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTTL)),
		IssuedAt:  jwt.NewNumericDate(now),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString(s.jwtSecret)
}

func isEmail(s string) bool {
	_, err := mail.ParseAddress(s)
	return err == nil
}

func toPBUser(u *DBUser) *cmpb.User {
	if u == nil {
		return nil
	}
	return &cmpb.User{
		Id:         u.ID,
		Email:      u.Email,
		Username:   u.Username,
		Bio:        u.Bio,
		AvatarPath: u.AvatarPath,
		CreatedAt:  timestamppb.New(u.CreatedAt),
	}
}

func (s *Server) GetProfile(ctx context.Context, req *idpb.GetProfileRequest) (*idpb.GetProfileResponse, error) {
	var userID string
	if req != nil && req.UserId != "" {
		userID = req.UserId
	} else {
		uid, err := s.userIDFromAuth(ctx) // read JWT from metadata
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "missing or invalid token")
		}
		userID = uid
	}
	u, err := s.repo.GetUserByID(ctx, userID)
	if err != nil || u == nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	return &idpb.GetProfileResponse{User: &cmpb.User{
		Id: u.ID, Email: u.Email, Username: u.Username, Bio: u.Bio,
		CreatedAt: timestamppb.New(u.CreatedAt),
	}}, nil
}

func (s *Server) userIDFromAuth(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.New("no metadata")
	}
	vals := md.Get("authorization")
	if len(vals) == 0 {
		return "", errors.New("no authorization")
	}
	token := vals[0]
	if strings.HasPrefix(strings.ToLower(token), "bearer ") {
		token = token[7:]
	}

	claims := &jwt.RegisteredClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return s.jwtSecret, nil
	})
	if err != nil || claims.Subject == "" {
		return "", errors.New("invalid token")
	}
	return claims.Subject, nil
}
