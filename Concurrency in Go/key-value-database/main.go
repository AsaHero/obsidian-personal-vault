package main

import (
	"bufio"
	"context"
	"fmt"

	"log"
	"log/slog"
	"os"
	"time"
)

func main() {
	// 1. Configure the Handler Options (Set minimum log level to Debug)
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug, // Default is slog.LevelInfo
	}

	// 2. Initialize a JSON Handler and set it as the global default logger
	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(handler)

	computeLayer := compute.NewParser(logger)
	inmemoryEngine := in_memory.NewEngine(logger)
	storageLayer, err := storage.NewStorage(inmemoryEngine, logger)
	if err != nil {
		log.Fatalf("failed to init storage layer: %s", err.Error())
	}
	database, err := NewDatabase(computeLayer, storageLayer, logger)
	if err != nil {
		log.Fatalf("failed to init database layer: %s", err.Error())
	}

	stdin := bufio.NewScanner(os.Stdin)
	for stdin.Scan() {
		request := stdin.Text()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
		fmt.Println(database.HandleQuery(ctx, request))
		cancel()
	}

	if err := stdin.Err(); err != nil {
		log.Fatalf("Error reading standard input: %v", err)
	}
}
