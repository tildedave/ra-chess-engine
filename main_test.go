package main

import (
	"log"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	logger = log.New(os.Stdout, "", log.LstdFlags)
	os.Exit(m.Run())
}
