package chatbot

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func parsePeriod(c echo.Context) (PeriodFilter, error) {
	now := time.Now()
	filter := PeriodFilter{Year: now.Year(), Month: 0}

	if value := strings.TrimSpace(c.QueryParam("year")); value != "" {
		year, err := strconv.Atoi(value)
		if err != nil {
			return PeriodFilter{}, errors.New("tahun analisis tidak valid")
		}
		filter.Year = year
	}
	if value := strings.TrimSpace(c.QueryParam("month")); value != "" {
		month, err := strconv.Atoi(value)
		if err != nil {
			return PeriodFilter{}, errors.New("bulan analisis tidak valid")
		}
		filter.Month = month
	}
	return filter, nil
}

func contextIdentity(c echo.Context) (role, userID string) {
	role, _ = c.Get("role").(string)
	userID, _ = c.Get("user_id").(string)
	return role, userID
}

// SendMessage is the primary endpoint for the actual chatbot UI.
// It interprets a user command first. Analytics are queried only when the command requires them.
func (h *Handler) SendMessage(c echo.Context) error {
	var request ChatMessageRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "format pesan chatbot tidak valid"})
	}
	role, userID := contextIdentity(c)
	result, err := h.service.HandleMessage(c.Request().Context(), request.Message, role, userID)
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

// Legacy analytical endpoints are retained for internal compatibility.
func (h *Handler) GetGUP(c echo.Context) error {
	filter, err := parsePeriod(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": err.Error()})
	}
	result, err := h.service.GetGUPAnalysis(c.Request().Context(), filter)
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) GetOutsideCity(c echo.Context) error {
	filter, err := parsePeriod(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": err.Error()})
	}
	role, userID := contextIdentity(c)
	result, err := h.service.GetOutsideCityAnalysis(c.Request().Context(), filter, role, userID)
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) GetInsideCity(c echo.Context) error {
	filter, err := parsePeriod(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": err.Error()})
	}
	role, userID := contextIdentity(c)
	result, err := h.service.GetInsideCityAnalysis(c.Request().Context(), filter, role, userID)
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

func writeError(c echo.Context, err error) error {
	if errors.Is(err, ErrForbidden) {
		return c.JSON(http.StatusForbidden, map[string]string{"message": err.Error()})
	}
	if strings.Contains(strings.ToLower(err.Error()), "tidak valid") {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": err.Error()})
	}
	return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
}
