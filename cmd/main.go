package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/krixlion/dev-forum_article/cmd/service"
	"github.com/krixlion/dev-forum_article/pkg/env"
)

// Hardcoded root dir name.
const projectDir = "app"

func main() {
	env.Load(projectDir)
	service.Run()

	signalChannel := make(chan os.Signal, 2)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

	sig := <-signalChannel
	switch sig {
	case os.Interrupt:
		//handle SIGINT
	case syscall.SIGTERM:
		//handle SIGTERM
	case syscall.SIGKILL:
		service.Close()
		os.Exit(0)
	}
}
