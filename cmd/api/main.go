package main

import (
	"context"
	routes "github.com/RobsonDevCode/GoApi/cmd/internal"
	"io"
	"log"
	"os"
)

func run(ctx context.Context, writer io.Writer, args []string) error {
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
