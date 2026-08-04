package gupreport

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

func (h *Handler) GetReport(c echo.Context) error {
	year, err := parseYear(c.QueryParam("year"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Tahun tidak valid")
	}

	result, err := h.service.GetReport(c.Request().Context(), year)
	if err != nil {
		return handleError(err)
	}

	return c.JSON(http.StatusOK, result)
}

func (h *Handler) SaveBudget(c echo.Context) error {
	var request SaveBudgetRequest
	if err := c.Bind(&request); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Data anggaran tidak valid")
	}

	if request.Year == 0 {
		request.Year = time.Now().Year()
	}

	result, err := h.service.SaveBudget(c.Request().Context(), request)
	if err != nil {
		return handleError(err)
	}

	return c.JSON(http.StatusOK, result)
}

func handleError(err error) error {
	if errors.Is(err, ErrValidation) {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
	}
	return echo.NewHTTPError(
		http.StatusInternalServerError,
		"Laporan GUP tidak dapat diproses",
	).SetInternal(err)
}
