package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Oluwaseun241/mura/cmd/api"
	"github.com/Oluwaseun241/mura/cmd/client"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

type Config struct {
	Port            string
	Environment     string
	ShutdownTimeout time.Duration
}

func loadConfig() Config {
	// Load Env variables for local development
	if os.Getenv("RUN_ENV") != "production" {
		if err := godotenv.Load(); err != nil {
			log.Printf("Warning: Error loading .env file: %v", err)
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	env := os.Getenv("RUN_ENV")
	if env == "" {
		env = "development"
	}

	return Config{
		Port:            port,
		Environment:     env,
		ShutdownTimeout: 10 * time.Second,
	}
}

func main() {
	// Load configuration
	cfg := loadConfig()

	// Initialize client connection
	client.Init()

	// Create Echo instance
	e := echo.New()

	// Configure logger
	e.Logger.SetLevel(log.INFO)
	if cfg.Environment == "development" {
		e.Logger.SetLevel(log.DEBUG)
	}

	// Middleware
	e.Use(middleware.CORS())
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: 30 * time.Second,
	}))

	// Health check endpoint
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "healthy"})
	})

	// Routes
	e.POST("/detect-food", api.FoodHandler)
	e.POST("/detect", api.IngredientHandler)
	e.POST("/recipe", api.RecipeHandler)

	// Start server in a goroutine
	go func() {
		if err := e.Start(":" + cfg.Port); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	// Shutdown the server
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}

	e.Logger.Info("Server gracefully stopped")
}
