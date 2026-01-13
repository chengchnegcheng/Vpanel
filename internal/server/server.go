// Package server provides HTTP server with graceful shutdown.
package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"v/internal/api"
	"v/internal/auth"
	"v/internal/config"
	"v/internal/database/repository"
	logservice "v/internal/log"
	"v/internal/logger"
	"v/internal/proxy"
)

// Server represents the HTTP server.
type Server struct {
	config       *config.Config
	logger       logger.Logger
	httpServer   *http.Server
	router       *api.Router
	authService  *auth.Service
	proxyManager proxy.Manager
	repos        *repository.Repositories
	logService   *logservice.Service
}

// New creates a new Server.
func New(
	cfg *config.Config,
	log logger.Logger,
	authService *auth.Service,
	proxyManager proxy.Manager,
	repos *repository.Repositories,
	logService *logservice.Service,
) *Server {
	return &Server{
		config:       cfg,
		logger:       log,
		authService:  authService,
		proxyManager: proxyManager,
		repos:        repos,
		logService:   logService,
	}
}

// Start starts the HTTP server.
func (s *Server) Start() error {
	// Create router
	s.router = api.NewRouter(
		s.config,
		s.logger,
		s.authService,
		s.proxyManager,
		s.repos,
		s.logService,
	)
	s.router.Setup()

	// Create HTTP server
	addr := fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.Port)
	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      s.router.Engine(),
		ReadTimeout:  s.config.Server.ReadTimeout,
		WriteTimeout: s.config.Server.WriteTimeout,
		IdleTimeout:  s.config.Server.IdleTimeout,
	}

	// Start server in goroutine
	go func() {
		s.logger.Info("starting HTTP server",
			logger.F("address", addr),
			logger.F("mode", s.config.Server.Mode),
		)

		var err error
		if s.config.Server.TLSCert != "" && s.config.Server.TLSKey != "" {
			err = s.httpServer.ListenAndServeTLS(s.config.Server.TLSCert, s.config.Server.TLSKey)
		} else {
			err = s.httpServer.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			s.logger.Error("HTTP server error", logger.F("error", err))
		}
	}()

	return nil
}

// Stop stops the HTTP server gracefully.
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("stopping HTTP server")

	if s.httpServer == nil {
		return nil
	}

	// Shutdown with context timeout
	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.Error("HTTP server shutdown error", logger.F("error", err))
		return err
	}

	s.logger.Info("HTTP server stopped")
	return nil
}

// Run starts the server and waits for shutdown signal.
func (s *Server) Run() error {
	// Start server
	if err := s.Start(); err != nil {
		return err
	}

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	s.logger.Info("shutdown signal received")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), s.config.Server.ShutdownTimeout)
	defer cancel()

	// Stop server
	return s.Stop(ctx)
}

// GracefulShutdown performs graceful shutdown with the given timeout.
func (s *Server) GracefulShutdown(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return s.Stop(ctx)
}

// Address returns the server address.
func (s *Server) Address() string {
	if s.httpServer == nil {
		return ""
	}
	return s.httpServer.Addr
}

// IsRunning returns true if the server is running.
func (s *Server) IsRunning() bool {
	return s.httpServer != nil
}
