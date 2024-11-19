//go:build consumer
// +build consumer

package main

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// Remove specific files before running tests
	if err := os.RemoveAll("pacts"); err != nil {
		panic(err)
	}

	// Run tests
	code := m.Run()

	// Exit with the test result code
	os.Exit(code)
}
