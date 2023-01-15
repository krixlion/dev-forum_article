package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/krixlion/dev_forum-article/cmd/service"
	"github.com/krixlion/dev_forum-article/pkg/env"
	"github.com/krixlion/dev_forum-article/pkg/logging"
	"github.com/krixlion/dev_forum-article/pkg/tracing"
)

// Hardcoded root dir name.
const projectDir = "app"

var port int

func init() {
	portFlag := flag.Int("p", 50051, "The gRPC server port")
	flag.Parse()
	port = *portFlag
	env.Load(projectDir)
}

func main() {
	shutdownTracing, err := tracing.InitProvider()
	if err != nil {
		logging.Log("Failed to initialize tracing", "err", err)
	}

	service := service.NewArticleService(port, getServiceDependencies())
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	service.Run(ctx)

	<-ctx.Done()
	logging.Log("Service shutting down")

	defer func() {
		cancel()
		shutdownTracing()
		err := service.Close()
		if err != nil {
			logging.Log("Failed to shutdown service", "err", err)
		} else {
			logging.Log("Service shutdown properly")
		}
	}()
}
