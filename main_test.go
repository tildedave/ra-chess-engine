package main

import (
	"log"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	logger = log.New(os.Stdout, "", log.LstdFlags)
	InitializeMoveBitboards()
	os.Exit(m.Run())
}
