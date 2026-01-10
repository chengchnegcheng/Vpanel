// Package main is the entry point for the V Panel application.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"v/internal/auth"
	"v/internal/config"
	"v/internal/database"
	"v/internal/database/repository"
	"v/internal/logger"
	"v/internal/proxy"
	"v/internal/proxy/protocols/shadowsocks"
	"v/internal/proxy/protocols/trojan"
	"v/internal/proxy/protocols/vless"
	"v/internal/proxy/protocols/vmess"
	"v/internal/server"
)

var (
	version   = "dev"
	buildTime = "unknown"
	gitCommit = "unknown"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "configs/config.yaml", "path to config file")
	showVersion := flag.Bool("version", false, "show version information")
	flag.Parse()

	// Show version and exit
	if *showVersion {
		fmt.Printf("V Panel %s\n", version)
		fmt.Printf("Build Time: %s\n", buildTime)
		fmt.Printf("Git Commit: %s\n", gitCommit)
		os.Exit(0)
	}

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}
	cfg.Version = version

	// Initialize logger
	log := logger.New(logger.Config{
		Level:  cfg.Log.Level,
		Format: cfg.Log.Format,
		Output: cfg.Log.Output,
	})

	log.Info("starting V Panel",
		logger.F("version", version),
		logger.F("config", *configPath),
	)

	// Initialize database
	db, err := database.New(&database.Config{
		Driver: cfg.Database.Driver,
		DSN:    cfg.Database.DSN,
	})
	if err != nil {
		log.Error("failed to initialize database", logger.F("error", err))
		os.Exit(1)
	}
	defer db.Close()

	// Run migrations
	if err := db.AutoMigrate(); err != nil {
		log.Error("failed to run migrations", logger.F("error", err))
		os.Exit(1)
	}

	// Initialize repositories
	repos := repository.NewRepositories(db.DB())

	// Initialize auth service
	authService := auth.NewService(auth.Config{
		JWTSecret:           cfg.Auth.JWTSecret,
		TokenExpiry:         cfg.Auth.TokenExpiry,
		RefreshTokenExpiry:  cfg.Auth.RefreshTokenExpiry,
	})

	// Initialize proxy manager
	proxyManager := proxy.NewManager(repos.Proxy)

	// Register protocols
	proxyManager.RegisterProtocol(vmess.New())
	proxyManager.RegisterProtocol(vless.New())
	proxyManager.RegisterProtocol(trojan.New())
	proxyManager.RegisterProtocol(shadowsocks.New())

	// Create and start server
	srv := server.New(cfg, log, authService, proxyManager, repos)

	if err := srv.Start(); err != nil {
		log.Error("failed to start server", logger.F("error", err))
		os.Exit(1)
	}

	log.Info("server started",
		logger.F("address", fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)),
	)

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	log.Info("shutdown signal received", logger.F("signal", sig.String()))

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Stop(ctx); err != nil {
		log.Error("server shutdown error", logger.F("error", err))
		os.Exit(1)
	}

	log.Info("server stopped gracefully")
}
