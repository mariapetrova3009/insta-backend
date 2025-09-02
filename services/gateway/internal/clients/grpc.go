package clients

import (
	cfgpkg "github.com/mariapetrova3009/insta-backend/pkg/config"
	contentpb "github.com/mariapetrova3009/insta-backend/proto/content"
	feedpb "github.com/mariapetrova3009/insta-backend/proto/feed"
	idpb "github.com/mariapetrova3009/insta-backend/proto/identity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Clients struct {
	Identity               idpb.IdentityServiceClient
	Content                contentpb.ContentServiceClient
	Feed                   feedpb.FeedServiceClient
	idConn, ctConn, fdConn *grpc.ClientConn
}

func MustInit(cfg *cfgpkg.Config) *Clients {
	idConn, _ := grpc.Dial(cfg.Identity.Endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	ctConn, _ := grpc.Dial(cfg.Content.Endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	fdConn, _ := grpc.Dial(cfg.Feed.Endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	return &Clients{
		Identity: idpb.NewIdentityServiceClient(idConn),
		Content:  contentpb.NewContentServiceClient(ctConn),
		Feed:     feedpb.NewFeedServiceClient(fdConn),
		idConn:   idConn, ctConn: ctConn, fdConn: fdConn,
	}
}
func (c *Clients) Close() {
	_ = c.idConn.Close()
	_ = c.ctConn.Close()
	_ = c.fdConn.Close()
}
