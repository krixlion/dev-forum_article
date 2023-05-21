package server

import (
	"context"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/krixlion/dev_forum-lib/tracing"
)

func (server ArticleServer) AuthFunc(ctx context.Context) (context.Context, error) {
	ctx, span := server.tracer.Start(ctx, "server.AuthFunc")
	defer span.End()

	token, err := grpc_auth.AuthFromMD(ctx, "Bearer")
	if err != nil {
		tracing.SetSpanErr(span, err)
		return nil, err
	}

	if err := server.tokenValidator.VerifyToken(token); err != nil {
		tracing.SetSpanErr(span, err)
		return nil, err
	}

	return ctx, nil
}
