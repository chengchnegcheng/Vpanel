// Package api provides HTTP API routes and handlers for the V Panel application.
package api

import (
	"github.com/gin-gonic/gin"

	"v/internal/api/handlers"
	"v/internal/api/middleware"
	"v/internal/auth"
	"v/internal/config"
	"v/internal/database/repository"
	"v/internal/logger"
	"v/internal/proxy"
)

// Router manages API routes.
type Router struct {
	engine       *gin.Engine
	config       *config.Config
	logger       logger.Logger
	authService  *auth.Service
	proxyManager proxy.Manager
	repos        *repository.Repositories
}

// NewRouter creates a new API router.
func NewRouter(
	cfg *config.Config,
	log logger.Logger,
	authService *auth.Service,
	proxyManager proxy.Manager,
	repos *repository.Repositories,
) *Router {
	// Set Gin mode based on config
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()

	return &Router{
		engine:       engine,
		config:       cfg,
		logger:       log,
		authService:  authService,
		proxyManager: proxyManager,
		repos:        repos,
	}
}

// Setup configures all routes and middleware.
func (r *Router) Setup() {
	// Global middleware
	r.engine.Use(middleware.Recovery(r.logger))
	r.engine.Use(middleware.Logger(r.logger))
	r.engine.Use(middleware.CORS(r.config.Server.CORSOrigins))
	r.engine.Use(middleware.RequestID())

	// Create handlers
	authHandler := handlers.NewAuthHandler(r.authService, r.repos.User, r.logger)
	proxyHandler := handlers.NewProxyHandler(r.proxyManager, r.repos.Proxy, r.logger)
	systemHandler := handlers.NewSystemHandler(r.config, r.logger)
	healthHandler := handlers.NewHealthHandler(r.repos, r.logger)

	// Auth middleware
	authMiddleware := middleware.NewAuthMiddleware(r.authService, r.logger)

	// Public routes
	r.engine.GET("/health", healthHandler.Health)
	r.engine.GET("/ready", healthHandler.Ready)

	// API routes
	api := r.engine.Group("/api")
	{
		// Auth routes (public)
		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
		}

		// Protected routes
		protected := api.Group("")
		protected.Use(authMiddleware.Authenticate())
		{
			// Auth routes (protected)
			protected.POST("/auth/logout", authHandler.Logout)
			protected.GET("/auth/me", authHandler.GetCurrentUser)
			protected.PUT("/auth/password", authHandler.ChangePassword)

			// Proxy routes
			proxies := protected.Group("/proxies")
			{
				proxies.GET("", proxyHandler.List)
				proxies.POST("", proxyHandler.Create)
				proxies.GET("/:id", proxyHandler.Get)
				proxies.PUT("/:id", proxyHandler.Update)
				proxies.DELETE("/:id", proxyHandler.Delete)
				proxies.GET("/:id/link", proxyHandler.GetShareLink)
				proxies.POST("/:id/toggle", proxyHandler.Toggle)
			}

			// System routes
			system := protected.Group("/system")
			{
				system.GET("/info", systemHandler.GetInfo)
				system.GET("/status", systemHandler.GetStatus)
				system.GET("/stats", systemHandler.GetStats)
			}

			// User management (admin only)
			users := protected.Group("/users")
			users.Use(authMiddleware.RequireRole("admin"))
			{
				users.GET("", authHandler.ListUsers)
				users.POST("", authHandler.CreateUser)
				users.GET("/:id", authHandler.GetUser)
				users.PUT("/:id", authHandler.UpdateUser)
				users.DELETE("/:id", authHandler.DeleteUser)
			}
		}
	}

	// Static files for frontend (if enabled)
	if r.config.Server.StaticPath != "" {
		r.engine.Static("/static", r.config.Server.StaticPath)
		r.engine.NoRoute(func(c *gin.Context) {
			c.File(r.config.Server.StaticPath + "/index.html")
		})
	}
}

// Engine returns the underlying Gin engine.
func (r *Router) Engine() *gin.Engine {
	return r.engine
}
