package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/krixlion/dev-forum_article/cmd/service"
	"github.com/krixlion/dev-forum_article/pkg/env"
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
	service := service.NewArticleService(port)
	service.Run()

	sigExitC := make(chan os.Signal, 1)
	signal.Notify(sigExitC, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	<-sigExitC
	log.Println("Service shutting down")

	defer func() {
		err := service.Close()
		if err != nil {
			log.Println("Failed to shutdown service")
		} else {
			log.Println("Service Exited Properly")
		}
	}()
}
