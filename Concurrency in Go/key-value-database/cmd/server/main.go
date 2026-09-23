package main

import (
	"context"
	"flag"
	"key-value-database/config"
	"key-value-database/database"
	"key-value-database/database/compute"
	"key-value-database/database/storage"
	"key-value-database/database/storage/engine/in_memory"
	logger_pkg "key-value-database/logger"
	"key-value-database/network"
	"log"
	"os"
	"os/signal"
	"time"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "", "pass yaml configuration file location")
	flag.Parse()

	cfg, err := config.New(configPath)
	if err != nil {
		log.Fatalf("failed to init config: %v", err.Error())
	}

	logger, err := logger_pkg.New(cfg.Logging.Output, cfg.Logging.Level)
	if err != nil {
		log.Fatalf("failed to init logger: %v", err.Error())
	}

	computeLayer := compute.NewParser(logger)
	inMemoryEngine := in_memory.NewEngine(logger)
	storageLayer, err := storage.NewStorage(inMemoryEngine, logger)
	if err != nil {
		log.Fatalf("failed to init storage: %v", err.Error())
	}

	database, err := database.NewDatabase(computeLayer, storageLayer, logger)
	if err != nil {
		log.Fatalf("failed to init database: %v", err.Error())
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	tcpServer, err := network.New(cfg, logger, database)
	if err != nil {
		log.Fatalf("failed to init tcp server: %v", err.Error())
	}

	go func() {
		logger.InfoContext(ctx, "Starting server on address", "address", cfg.Network.Address)
		if err := tcpServer.Start(ctx); err != nil {
			logger.ErrorContext(ctx, "failed to start tcp server", "error", err)
			return
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = tcpServer.Stop(shutdownCtx)
	log.Println("application shutdown gracefully...")
}
