package main

import (
	"fmt"
	"os"
	"regexp"

	"github.com/joho/godotenv"
	"github.com/krixlion/dev-forum_article/cmd/service"
)

// Hardcoded root dir name.
const projectDir = "app"

// LoadEnv assumes the root directory name is "app" and
// the .env file is located in the root directory.
// It panics if it cannot find the file named ".env" in the root dir.
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
	service.Run()
}
