// services/gateway/internal/http/handlers.go
package http

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"

	gatewayauth "github.com/mariapetrova3009/insta-backend/services/gateway/internal/auth"
	"github.com/mariapetrova3009/insta-backend/services/gateway/internal/clients"

	cmpb "github.com/mariapetrova3009/insta-backend/proto/common"
	contentpb "github.com/mariapetrova3009/insta-backend/proto/content"
	feedpb "github.com/mariapetrova3009/insta-backend/proto/feed"
	idpb "github.com/mariapetrova3009/insta-backend/proto/identity"
)

func Healthz() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}
}

// --------------------------------- AUTH --------------------------------------

func Register(cl *clients.Clients) http.HandlerFunc {
	type req struct {
		Email    string `json:"email"`
		Username string `json:"username"`
		Password string `json:"password"`
		Bio      string `json:"bio"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var in req
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			httpError(w, http.StatusBadRequest, "bad json")
			return
		}

		res, err := cl.Identity.Register(r.Context(), &idpb.RegisterRequest{
			Email:    in.Email,
			Username: in.Username,
			Password: in.Password,
			Bio:      in.Bio,
		})
		if err != nil {
			httpError(w, http.StatusBadGateway, err.Error())
			return
		}
		respondJSON(w, http.StatusOK, res)
	}

}

func Login(cl *clients.Clients) http.HandlerFunc {
	type req struct {
		EmailOrUsername string `json:"email_or_username"`
		Password        string `json:"password"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var in req
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			httpError(w, http.StatusBadRequest, "bad json")
			return
		}
		res, err := cl.Identity.Login(r.Context(), &idpb.LoginRequest{
			EmailOrUsername: in.EmailOrUsername,
			Password:        in.Password,
		})
		if err != nil {
			httpError(w, http.StatusBadGateway, err.Error())
		}
		respondJSON(w, http.StatusOK, res)
	}
}

func Me(cl *clients.Clients) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// достаём из http запроса заголовок Autorization и превращаем в gRPC metadata
		md := gatewayauth.MetadataFromHTTP(r)
		ctx := gatewayauth.Outgoing(r.Context(), md)
		res, err := cl.Identity.GetProfile(ctx, &idpb.GetProfileRequest{})
		if err != nil {
			httpError(w, http.StatusBadGateway, err.Error())
			return
		}
		respondJSON(w, http.StatusOK, res)
	}
}

// --------------------------------- POSTS -------------------------------------
func CreatePost(cl *clients.Clients) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// read autorization
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			httpError(w, http.StatusBadRequest, "multipart form required")
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			httpError(w, http.StatusBadRequest, "file is required")
			return
		}
		defer file.Close()

		// get caption and mime

		caption := r.FormValue("caption")
		mime := r.Header.Get("Content-Type")
		if mime == "" {
			mime = "application/octet-stream"
		}

		// read data
		data, err := readAll(file)
		if err != nil {
			httpError(w, http.StatusBadRequest, "read file error")
		}

		// upload media

		res, err := cl.Content.UploadMedia(r.Context(), &contentpb.UploadMediaRequest{
			Data: data,
			Mime: mime,
			Name: header.Filename,
		})
		if err != nil {
			httpError(w, http.StatusBadGateway, err.Error())
		}

		// create post

		cp, err := cl.Content.CreatePost(r.Context(), &contentpb.CreatePostRequest{
			Caption:   caption,
			MediaPath: res.MediaPath,
			Mime:      res.Mime,
		})
		if err != nil {
			httpError(w, http.StatusBadGateway, err.Error())
			return
		}

		respondJSON(w, http.StatusOK, cp)
	}
}

// ---------------------------------- FEED -------------------------------------

func GetFeed(cl *clients.Clients) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var page *cmpb.PageRequest
		limitStr := r.URL.Query().Get("limit")
		cursor := r.URL.Query().Get("cursor")

		if limitStr == "" || cursor == "" {
			page = &cmpb.PageRequest{}
		}
		if limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil {
				page.Limit = uint32(l)
			}
		}
		if cursor != "" {
			page.Cursor = &cmpb.Cursor{Token: cursor}
		}

		res, err := cl.Feed.GetFeed(r.Context(), &feedpb.GetFeedRequest{
			Page: page,
		})
		if err != nil {
			httpError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondJSON(w, http.StatusOK, res)
	}
}

// ------------------------------ small helpers --------------------------------

// readAll вынесен сюда, чтобы не тянуть лишние зависимости в responses.go
func readAll(f multipart.File) ([]byte, error) {
	b, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("readAll: %w", err)
	}
	return b, nil
}
