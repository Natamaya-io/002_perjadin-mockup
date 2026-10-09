package fundsource

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func parseYear(value string) (int, error) {
	if strings.TrimSpace(value) == "" {
		return time.Now().Year(), nil
	}
	return strconv.Atoi(value)
}

func (h *Handler) GetDashboard(c echo.Context) error {
	year, err := parseYear(c.QueryParam("year"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Tahun tidak valid")
	}

	result, err := h.service.GetDashboard(c.Request().Context(), year)
	if err != nil {
		return handleError(err)
	}
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) GetSourceOptions(c echo.Context) error {
	year, err := parseYear(c.QueryParam("year"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Tahun tidak valid")
	}

	result, err := h.service.GetSourceOptions(
		c.Request().Context(),
		year,
		c.QueryParam("budgetType"),
		c.QueryParam("spdNumber"),
	)
	if err != nil {
		return handleError(err)
	}
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) CreateSource(c echo.Context) error {
	var request SaveSourceRequest
	if err := c.Bind(&request); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Data sumber dana tidak valid")
	}
	result, err := h.service.CreateSource(c.Request().Context(), request)
	if err != nil {
		return handleError(err)
	}
	return c.JSON(http.StatusCreated, result)
}

func (h *Handler) UpdateSource(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "ID sumber dana tidak valid")
	}

	var request SaveSourceRequest
	if err := c.Bind(&request); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Data sumber dana tidak valid")
	}
	result, err := h.service.UpdateSource(c.Request().Context(), id, request)
	if err != nil {
		return handleError(err)
	}
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) DeleteSource(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "ID sumber dana tidak valid")
	}
	if err := h.service.DeleteSource(c.Request().Context(), id); err != nil {
		return handleError(err)
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "Sumber dana berhasil dihapus"})
}

func (h *Handler) CreateAllocation(c echo.Context) error {
	var request SaveAllocationRequest
	if err := c.Bind(&request); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Data alokasi tidak valid")
	}
	result, err := h.service.CreateAllocation(c.Request().Context(), request)
	if err != nil {
		return handleError(err)
	}
	return c.JSON(http.StatusCreated, result)
}

func (h *Handler) DeleteAllocation(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "ID alokasi tidak valid")
	}
	if err := h.service.DeleteAllocation(c.Request().Context(), id); err != nil {
		return handleError(err)
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "Alokasi berhasil dihapus"})
}

func handleError(err error) error {
	switch {
	case errors.Is(err, ErrValidation):
		return echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
	case errors.Is(err, ErrNotFound):
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	case errors.Is(err, ErrDuplicate):
		return echo.NewHTTPError(http.StatusConflict, err.Error())
	case errors.Is(err, ErrSourceInUse):
		return echo.NewHTTPError(http.StatusConflict, err.Error())
	case errors.Is(err, ErrAllocationInUse):
		return echo.NewHTTPError(http.StatusConflict, err.Error())
	case errors.Is(err, ErrInsufficientFunds):
		return echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
	default:
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			"Sumber dana tidak dapat diproses",
		).SetInternal(err)
	}
}
