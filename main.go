package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/mcg-dallgow/mcg-display/handlers"
)

func main() {
	// Echo instance
	e := echo.New()

	// Static assets
	e.Static("/static", "static")

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	handlers.SetupRoutes(e)

	// Start server
	e.Logger.Fatal(e.Start(":3000"))
}
