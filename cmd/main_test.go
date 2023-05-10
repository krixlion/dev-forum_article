package main

import (
	"testing"

	"go.uber.org/goleak"
)

// Avoid flags being parsed by the 'go test' before they are defined in the main's init func.
var _ = func() bool {
	testing.Init()
	return true
}()

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}
