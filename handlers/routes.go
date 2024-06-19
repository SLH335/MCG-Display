package handlers

import "github.com/labstack/echo/v4"

func SetupRoutes(e *echo.Echo) {
	e.GET("/", Index)
}
