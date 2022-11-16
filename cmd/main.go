package main

import (
	"fmt"
	"os"
	"regexp"

	"github.com/joho/godotenv"
)

const projectDir = "dev-forum_article"

func loadEnv() {
	re := regexp.MustCompile(`^(.*` + projectDir + `)`)
	cwd, _ := os.Getwd()
	rootPath := re.Find([]byte(cwd))

	err := godotenv.Load(string(rootPath) + `/.env`)
	if err != nil {
		panic(fmt.Sprintf("Failed to load .env, err: %s", err))
	}
}

func main() {
	loadEnv()
}
