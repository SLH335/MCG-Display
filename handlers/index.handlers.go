package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/mcg-dallgow/mcg-display/components"
	"github.com/mcg-dallgow/mcg-display/services"
)

func Index(c echo.Context) error {
	return c.HTML(http.StatusOK, services.RenderComponent(components.Index()))
}
