package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "social/docs" // 导入生成的docs包
	"social/internal/config"
	"social/internal/handlers"
	"social/internal/middleware"
	"social/internal/platforms"
	"social/internal/storage"
	"social/pkg/logger"
)

// @title Social Media Platform API
// @version 1.0
// @description 多平台社交媒体授权分享API
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8084
// @BasePath /
// @schemes http https
func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	appLogger := logger.NewLogger()

	// Initialize Redis storage
	redisStorage, err := storage.NewRedisStorage(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		log.Fatalf("Failed to initialize Redis storage: %v", err)
	}
	defer func() {
		if err := redisStorage.Close(); err != nil {
			log.Printf("Failed to close Redis storage: %v", err)
		}
	}()

	// Initialize platform registry
	platformRegistry := platforms.NewRegistry()

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(cfg, redisStorage, platformRegistry, appLogger)
	shareHandler := handlers.NewShareHandler(cfg, redisStorage, platformRegistry, appLogger)
	healthHandler := handlers.NewHealthHandler(redisStorage, appLogger)

	// Initialize request middleware
	requestMiddleware := middleware.NewRequestMiddleware(appLogger)

	// Setup Gin router
	router := setupRouter(authHandler, shareHandler, healthHandler, requestMiddleware)

	// Create HTTP server
	server := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		appLogger.Info(context.Background(), "Starting server", "addr", server.Addr, "base_url", cfg.Server.BaseURL)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info(context.Background(), "Shutting down server...")

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	appLogger.Info(context.Background(), "Server exited")
}

// setupRouter configures the Gin router with all routes
func setupRouter(authHandler *handlers.AuthHandler, shareHandler *handlers.ShareHandler, healthHandler *handlers.HealthHandler, requestMiddleware *middleware.RequestMiddleware) *gin.Engine {
	// Set Gin mode based on environment
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Add middleware
	router.Use(gin.Recovery())
	router.Use(requestMiddleware.RequestID()) // 添加request ID中间件

	// Health check endpoint
	router.GET("/health", healthHandler.Health)

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Static files and test pages
	router.Static("/static", "./static")
	router.GET("/test", func(c *gin.Context) {
		c.File("./static/test.html")
	})
	router.GET("/callback", func(c *gin.Context) {
		c.File("./static/callback.html")
	})

	// OAuth endpoints
	router.POST("/auth/start", authHandler.StartAuth)
	router.POST("/auth/callback", authHandler.Callback)
	router.POST("/auth/is-authorized", authHandler.IsAuthorized)
	router.POST("/auth/user-info", authHandler.GetUserInfo)
	router.POST("/auth/refresh-token", authHandler.RefreshToken)

	// API endpoints - RESTful design
	api := router.Group("/api")
	{
		// Legacy endpoints for backward compatibility
		api.POST("/share", shareHandler.Share)
		api.POST("/stats", shareHandler.GetStats)

		// Recent posts endpoints
		api.POST("/recent-posts", shareHandler.GetRecentPosts)
		api.POST("/batch-recent-posts", shareHandler.BatchGetRecentPosts)
	}

	return router
}
