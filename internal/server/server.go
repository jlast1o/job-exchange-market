package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

type Config struct {
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

type Server struct {
	httpServer *http.Server
	logger     *zap.Logger
	config     Config
}

func New(cfg Config, logger *zap.Logger, router chi.Router) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:         fmt.Sprintf(":%d", cfg.Port),
			Handler:      router,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
		},
		logger: logger,
		config: cfg,
	}
}

func (s *Server) Run() error {
	quit := make(chan os.Signal, 1) // получаем сигналы от ОС
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		s.logger.Info("server starting", zap.Int("port", s.config.Port))
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Fatal("Server failed to start", zap.Error(err))
		}
	}()

	sig := <-quit

	s.logger.Info("Received signal, server is shutting down", zap.String("signal", sig.String()))

	ctx, cancel := context.WithTimeout(context.Background(), s.config.ShutdownTimeout)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.Error("error during shutting down server", zap.Error(err))
		return err
	}

	s.logger.Info("server successefully and gracefully stopped")
	return nil
}
