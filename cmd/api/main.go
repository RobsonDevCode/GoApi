package main

import (
	"context"
	"github.com/RobsonDevCode/GoApi/cmd/api/internal"
	"github.com/RobsonDevCode/GoApi/cmd/api/settings/configuration"
	"io"
	"log"
	"os"
)

func run(ctx context.Context, writer io.Writer, args []string) error {
	if err := configuration.SetEnvironmentSettings("development"); err != nil {
		log.Fatalf("Failed to set environment variables: %v", err)
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := routes.NewRouter()

	if err != nil {
		log.Fatal(err)
		return err
	}

	return nil
}

func main() {
	ctx := context.Background()

	if err := run(ctx, os.Stdout, os.Args[1:]); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
