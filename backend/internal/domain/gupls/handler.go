package gupls

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func parseYear(value string) (int, error) {
	if value == "" {
		return time.Now().Year(), nil
	}
	return strconv.Atoi(value)
}

func (h *Handler) Get(c echo.Context) error {
	year, err := parseYear(c.QueryParam("year"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Tahun tidak valid")
	}

	result, err := h.service.Get(c.Request().Context(), year)
	if err != nil {
		return handleError(err)
	}

	return c.JSON(http.StatusOK, result)
}

func (h *Handler) Save(c echo.Context) error {
	var request SaveRequest
	if err := c.Bind(&request); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Data LS tidak valid")
	}

	if request.Year == 0 {
		request.Year = time.Now().Year()
	}

	result, err := h.service.Save(c.Request().Context(), request)
	if err != nil {
		return handleError(err)
	}

	return c.JSON(http.StatusOK, result)
}

func handleError(err error) error {
	switch {
	case errors.Is(err, ErrValidation):
		return echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
	case errors.Is(err, ErrFundingSourceNotFound):
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	default:
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			"Data LS GUP tidak dapat diproses",
		).SetInternal(err)
	}
}
