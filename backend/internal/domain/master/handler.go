package master

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) GetProvinces(c echo.Context) error {
	provinces, err := h.svc.GetProvinces(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, provinces)
}

func (h *Handler) GetSBMRates(c echo.Context) error {
	rates, err := h.svc.GetSBMRates(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, rates)
}

func (h *Handler) GetSettings(c echo.Context) error {
	settings, err := h.svc.GetSettings(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, settings)
}

func (h *Handler) UpdateSettings(c echo.Context) error {
	var settings map[string]string
	if err := c.Bind(&settings); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := h.svc.UpdateSettings(c.Request().Context(), settings); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Settings updated successfully"})
}
