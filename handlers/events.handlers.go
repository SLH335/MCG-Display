package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/mcg-dallgow/mcg-display/components"
	"github.com/mcg-dallgow/mcg-display/services"
	. "github.com/mcg-dallgow/mcg-display/types"
)

func Events(c echo.Context) error {
	start := c.QueryParam("start")
	end := c.QueryParam("end")
	days := c.QueryParam("days")

	startDate, endDate, err := services.ParseDateRange(start, end, days)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: err.Error(),
		})
	}
	events, err := services.GetEvents(startDate, endDate)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: err.Error(),
		})
	}

	return c.HTML(http.StatusOK, services.RenderComponent(components.Events(events)))
}
