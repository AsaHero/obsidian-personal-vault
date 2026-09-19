package network

import (
	"bufio"
	"context"
	"errors"
	"key-value-database/config"
	"key-value-database/database"
	"log"
	"log/slog"
	"net"
	"sync"
)

type Server struct {
	cfg      *config.Config
	logger   *slog.Logger
	database *database.Database
	listener net.Listener
}

func New(cfg *config.Config, logger *slog.Logger, database *database.Database) (*Server, error) {
	if cfg == nil {
		return nil, errors.New("config is required")
	}

	if logger == nil {
		return nil, errors.New("logger is required")
	}

	if database == nil {
		return nil, errors.New("database is required")
	}

	return &Server{
		cfg:      cfg,
		logger:   logger,
		database: database,
	}, nil
}

func (s *Server) Start(ctx context.Context) {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("error on Listen: %v", err.Error())
	}
	s.listener = listener

	wg := sync.WaitGroup{}
	semaphore := make(chan struct{}, s.cfg.Network.MaxConnections)
	for {
		select 

		conn, err := listener.Accept()
		if err != nil {
			s.logger.ErrorContext(ctx, "failed to accept connection", "error", err)
			continue
		}

		semaphore <- struct{}{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() { <-semaphore }()
			s.handle(ctx, conn)
		}()
	}
}

func (s *Server) Stop() {
}

func (s *Server) handle(ctx context.Context, conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		request, err := reader.ReadString('\n')
		if err != nil {
			s.logger.InfoContext(ctx, "client disconnected or error", "error", err)
			return
		}

		response := s.database.HandleQuery(ctx, request)

		_, err = conn.Write([]byte(response))
		if err != nil {
			s.logger.ErrorContext(ctx, "failed to write to socket", "error", err)
		}
	}
}
