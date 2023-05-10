package server

import (
	"context"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
)

func (server ArticleServer) AuthFunc(ctx context.Context) (context.Context, error) {
	token, err := grpc_auth.AuthFromMD(ctx, "Bearer")
	if err != nil {
		return nil, err
	}

	if err := server.tokenValidator.VerifyToken(token); err != nil {
		return nil, err
	}

	return ctx, nil
}
