package main

import (
	"log"

	"github.com/Oluwaseun241/mura/cmd/api"
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

	e := echo.New()

	// Middleware
	e.Use(middleware.CORS())
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.POST("/detect", api.IngredientHandler)
	e.POST("/get-recipe", api.RecipeHandler)

	e.Logger.Fatal(e.Start(":3000"))
}
