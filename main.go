package main

import (
	"log"
	"os"

	"github.com/Oluwaseun241/mura/cmd/api"
	"github.com/Oluwaseun241/mura/cmd/client"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Load Env variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// initialize client connection
	client.Init()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	e := echo.New()

	// Middleware
	e.Use(middleware.CORS())
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.POST("/detect-food", api.FoodHandler)
	e.POST("/detect", api.IngredientHandler)
	e.POST("/recipe", api.RecipeHandler)
	e.POST("/test", api.YtHandler)

	e.Logger.Fatal(e.Start(":" + port))
}
