package network

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"key-value-database/config"
	"key-value-database/database"
	"key-value-database/logger"
	"net"
	"sync"
)

type Server struct {
	cfg      *config.Config
	logger   *logger.Logger
	database *database.Database
	listener net.Listener
	wg       sync.WaitGroup
}

func New(cfg *config.Config, logger *logger.Logger, database *database.Database) (*Server, error) {
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

func (s *Server) Start(ctx context.Context) (err error) {
	s.listener, err = net.Listen("tcp", s.cfg.Network.Address)
	if err != nil {
		return fmt.Errorf("failed to listen tcp: %w", err)
	}

	semaphore := make(chan struct{}, s.cfg.Network.MaxConnections)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		conn, err := s.listener.Accept()
		if err != nil {
			if err == net.ErrClosed {
				return nil
			}
			s.logger.ErrorContext(ctx, "failed to accept connection", "error", err)
			continue
		}

		semaphore <- struct{}{}
		s.wg.Add(1)
		go func(conn net.Conn) {
			defer s.wg.Done()
			defer func() { <-semaphore }()
			defer func() {
				if err := recover(); err != nil {
					s.logger.ErrorContext(ctx, "panic recovered", "error", err)
				}
			}()
			s.handle(ctx, conn)
		}(conn)
	}
}

func (s *Server) Stop(ctx context.Context) error {
	_ = s.listener.Close()

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	default:
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return nil
	}

}

func (s *Server) handle(ctx context.Context, conn net.Conn) {
	defer conn.Close()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

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

		_, err = conn.Write([]byte(response + "\n"))
		if err != nil {
			s.logger.ErrorContext(ctx, "failed to write to socket", "error", err)
		}
	}
}
