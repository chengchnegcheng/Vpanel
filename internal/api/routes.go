// Package api provides HTTP API routes and handlers for the V Panel application.
package api

import (
	"context"

	"github.com/gin-gonic/gin"

	"v/internal/api/handlers"
	"v/internal/api/middleware"
	"v/internal/auth"
	"v/internal/config"
	"v/internal/database/repository"
	"v/internal/logger"
	"v/internal/proxy"
	"v/internal/settings"
	"v/internal/xray"
)

// Router manages API routes.
type Router struct {
	engine          *gin.Engine
	config          *config.Config
	logger          logger.Logger
	authService     *auth.Service
	proxyManager    proxy.Manager
	repos           *repository.Repositories
	settingsService *settings.Service
	xrayManager     xray.Manager
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

	// Create settings service
	settingsService := settings.NewService(repos.Settings)

	// Create Xray manager
	xrayManager := xray.NewManager(xray.Config{
		BinaryPath: cfg.Xray.BinaryPath,
		ConfigPath: cfg.Xray.ConfigPath,
		BackupDir:  cfg.Xray.BackupDir,
	}, log)

	return &Router{
		engine:          engine,
		config:          cfg,
		logger:          log,
		authService:     authService,
		proxyManager:    proxyManager,
		repos:           repos,
		settingsService: settingsService,
		xrayManager:     xrayManager,
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
	authHandler := handlers.NewAuthHandler(r.authService, r.repos.User, r.repos.LoginHistory, r.logger)
	proxyHandler := handlers.NewProxyHandlerWithTraffic(r.proxyManager, r.repos.Proxy, r.repos.Traffic, r.logger)
	systemHandler := handlers.NewSystemHandler(r.config, r.logger)
	healthHandler := handlers.NewHealthHandler(r.repos, r.logger, r.xrayManager, nil)
	roleHandler := handlers.NewRoleHandler(r.logger, r.repos.Role)
	statsHandler := handlers.NewStatsHandler(r.logger, r.repos, nil)
	settingsHandler := handlers.NewSettingsHandler(r.logger, r.settingsService)
	xrayHandler := handlers.NewXrayHandler(r.xrayManager, r.logger)
	certificatesHandler := handlers.NewCertificatesHandler(r.logger)

	// Initialize system roles
	ctx := context.Background()
	if err := roleHandler.InitSystemRoles(ctx); err != nil {
		r.logger.Error("Failed to initialize system roles", logger.F("error", err))
	}

	// Auth middleware
	authMiddleware := middleware.NewAuthMiddleware(r.authService, r.logger)

	// Access control middleware (checks traffic limits and expiration)
	accessControlMiddleware := middleware.NewAccessControlMiddleware(r.repos.User, r.logger)

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

		// SSE endpoint (placeholder - returns 204 No Content to avoid HTML fallback)
		api.GET("/sse/xray-events", func(c *gin.Context) {
			c.Status(204)
		})

		// Protected routes
		protected := api.Group("")
		protected.Use(authMiddleware.Authenticate())
		{
			// Auth routes (protected)
			protected.POST("/auth/logout", authHandler.Logout)
			protected.GET("/auth/me", authHandler.GetCurrentUser)
			protected.PUT("/auth/password", authHandler.ChangePassword)

			// Proxy routes - with access control for traffic limits and expiration
			proxies := protected.Group("/proxies")
			proxies.Use(accessControlMiddleware.CheckProxyAccess())
			{
				proxies.GET("", proxyHandler.List)
				proxies.POST("", proxyHandler.Create)
				proxies.POST("/batch", proxyHandler.BatchOperation)
				proxies.GET("/:id", proxyHandler.Get)
				proxies.PUT("/:id", proxyHandler.Update)
				proxies.DELETE("/:id", proxyHandler.Delete)
				proxies.GET("/:id/link", proxyHandler.GetShareLink)
				proxies.POST("/:id/toggle", proxyHandler.Toggle)
				proxies.POST("/:id/start", proxyHandler.Start)
				proxies.POST("/:id/stop", proxyHandler.Stop)
				proxies.GET("/:id/stats", proxyHandler.GetStats)
			}

			// System routes
			system := protected.Group("/system")
			{
				system.GET("/info", systemHandler.GetInfo)
				system.GET("/status", systemHandler.GetDetailedStatus)
				system.GET("/stats", systemHandler.GetStats)
			}

			// Role routes
			roles := protected.Group("/roles")
			{
				roles.GET("", roleHandler.ListRoles)
				roles.POST("", roleHandler.CreateRole)
				roles.GET("/:id", roleHandler.GetRole)
				roles.PUT("/:id", roleHandler.UpdateRole)
				roles.DELETE("/:id", roleHandler.DeleteRole)
			}

			// Permissions route
			protected.GET("/permissions", roleHandler.GetPermissions)

			// Stats routes
			stats := protected.Group("/stats")
			{
				stats.GET("/dashboard", statsHandler.GetDashboardStats)
				stats.GET("/protocol", statsHandler.GetProtocolStats)
				stats.GET("/traffic", statsHandler.GetTrafficStats)
				stats.GET("/user", statsHandler.GetUserStats)
				stats.GET("/detailed", statsHandler.GetDetailedStats)
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
				users.POST("/:id/enable", authHandler.EnableUser)
				users.POST("/:id/disable", authHandler.DisableUser)
				users.POST("/:id/reset-password", authHandler.ResetPassword)
				users.GET("/:id/login-history", authHandler.GetLoginHistory)
				users.DELETE("/:id/login-history", authHandler.ClearLoginHistory)
			}

			// Settings routes (admin only)
			settingsRoutes := protected.Group("/settings")
			settingsRoutes.Use(authMiddleware.RequireRole("admin"))
			{
				settingsRoutes.GET("", settingsHandler.GetSettings)
				settingsRoutes.PUT("", settingsHandler.UpdateSettings)
				settingsRoutes.POST("/backup", settingsHandler.BackupSettings)
				settingsRoutes.POST("/restore", settingsHandler.RestoreSettings)
				settingsRoutes.GET("/xray", settingsHandler.GetXraySettings)
				settingsRoutes.POST("/xray", settingsHandler.UpdateXraySettings)
				settingsRoutes.GET("/protocols", settingsHandler.GetProtocolSettings)
				settingsRoutes.POST("/protocols", settingsHandler.UpdateProtocolSettings)
			}

			// Xray routes (admin only)
			xrayRoutes := protected.Group("/xray")
			xrayRoutes.Use(authMiddleware.RequireRole("admin"))
			{
				xrayRoutes.GET("/status", xrayHandler.GetStatus)
				xrayRoutes.POST("/start", xrayHandler.Start)
				xrayRoutes.POST("/stop", xrayHandler.Stop)
				xrayRoutes.POST("/restart", xrayHandler.Restart)
				xrayRoutes.GET("/config", xrayHandler.GetConfig)
				xrayRoutes.PUT("/config", xrayHandler.UpdateConfig)
				xrayRoutes.POST("/validate", xrayHandler.ValidateConfig)
				xrayRoutes.POST("/test-config", xrayHandler.TestConfig)
				xrayRoutes.GET("/version", xrayHandler.GetVersion)
				xrayRoutes.GET("/versions", xrayHandler.GetVersions)
				xrayRoutes.POST("/sync-versions", xrayHandler.SyncVersions)
				xrayRoutes.GET("/check-updates", xrayHandler.CheckUpdates)
				xrayRoutes.POST("/download", xrayHandler.Download)
				xrayRoutes.POST("/install", xrayHandler.Install)
				xrayRoutes.POST("/update", xrayHandler.Update)
				xrayRoutes.POST("/switch-version", xrayHandler.SwitchVersion)
			}

			// Certificates routes (admin only)
			certificatesRoutes := protected.Group("/certificates")
			certificatesRoutes.Use(authMiddleware.RequireRole("admin"))
			{
				certificatesRoutes.GET("", certificatesHandler.List)
				certificatesRoutes.POST("/apply", certificatesHandler.Apply)
				certificatesRoutes.POST("/upload", certificatesHandler.Upload)
				certificatesRoutes.POST("/:id/renew", certificatesHandler.Renew)
				certificatesRoutes.GET("/:id/validate", certificatesHandler.Validate)
				certificatesRoutes.DELETE("/:id", certificatesHandler.Delete)
				certificatesRoutes.PUT("/:id/auto-renew", certificatesHandler.UpdateAutoRenew)
			}
		}
	}

	// Static files for frontend (if enabled)
	if r.config.Server.StaticPath != "" {
		// Serve static assets (js, css, images, etc.)
		r.engine.Static("/assets", r.config.Server.StaticPath+"/assets")
		// Serve favicon
		r.engine.StaticFile("/favicon.ico", r.config.Server.StaticPath+"/favicon.ico")
		// SPA fallback - serve index.html for all other routes
		r.engine.NoRoute(func(c *gin.Context) {
			c.File(r.config.Server.StaticPath + "/index.html")
		})
	}
}

// Engine returns the underlying Gin engine.
func (r *Router) Engine() *gin.Engine {
	return r.engine
}
