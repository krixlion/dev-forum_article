package main

import (
	"log"
	"testing"

	"github.com/krixlion/dev_forum-article/pkg/storage/redis/testdata"
	"go.uber.org/goleak"
)

// Avoid flags being parsed by the 'go test' before they are defined in the main's init func.
var _ = func() bool {
	testing.Init()
	return true
}()

func TestMain(m *testing.M) {
	if !testing.Short() {
		if err := testdata.Seed(); err != nil {
			log.Fatalf("Failed to seed before the tests: %v", err)
		}
	}
	goleak.VerifyTestMain(m)
}
