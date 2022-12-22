package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/krixlion/dev-forum_article/cmd/service"
	"github.com/krixlion/dev-forum_article/pkg/env"
	"github.com/krixlion/dev-forum_article/pkg/logging"
	"github.com/krixlion/dev-forum_article/pkg/tracing"
)

var port int

func init() {
	portFlag := flag.Int("p", 50051, "The gRPC server port")
	flag.Parse()
	port = *portFlag
}

// Hardcoded root dir name.
const projectDir = "app"

func main() {
	env.Load(projectDir)

	// Make InitProvider return err instead of calling log.Fatal()
	shutdownTracing, err := tracing.InitProvider()
	if err != nil {
		logging.Log("Failed to initialize tracing", "err", err)
	}

	service := service.NewArticleService(port)
	service.Run()

	sigExitC := make(chan os.Signal, 1)
	signal.Notify(sigExitC, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	<-sigExitC
	log.Println("Service shutting down")

	defer func() {
		shutdownTracing()
		err := service.Close()
		if err != nil {
			logging.Log("Failed to shutdown service", "err", err)
		} else {
			logging.Log("Service exited properly")
		}
	}()
}
