package main

import (
	"github.com/RobsonDevCode/GoApi/cmd/api/internal"
	"github.com/RobsonDevCode/GoApi/cmd/api/settings/configuration"
	"log"
	"os"
)

func run() error {
	if err := configuration.SetEnvironmentSettings("development"); err != nil {
		log.Fatalf("Failed to set environment variables: %v", err)
		return err
	}

	err := routes.NewRouter()
	if err != nil {
		log.Fatal(err)
		return err
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
