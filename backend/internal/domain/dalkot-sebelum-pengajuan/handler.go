package dalkot

import (
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
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

func (h *Handler) List(c echo.Context) error {
	result, err := h.service.List(c.Request().Context())
	if err != nil {
		return writeError(c, err)
	}

	return c.JSON(http.StatusOK, result)
}

func (h *Handler) GetMaster(c echo.Context) error {
	result, err := h.service.GetMaster(c.Request().Context())
	if err != nil {
		return writeError(c, err)
	}

	return c.JSON(http.StatusOK, result)
}

func (h *Handler) GetByID(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": "ID Dalkot tidak valid",
		})
	}

	result, err := h.service.GetByID(c.Request().Context(), id)
	if err != nil {
		return writeError(c, err)
	}

	return c.JSON(http.StatusOK, result)
}

func (h *Handler) Create(c echo.Context) error {
	input, err := parseFormInput(c, false)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": err.Error(),
		})
	}

	result, err := h.service.Create(c.Request().Context(), input)
	if err != nil {
		return writeError(c, err)
	}

	return c.JSON(http.StatusCreated, result)
}

func (h *Handler) Update(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": "ID Dalkot tidak valid",
		})
	}

	input, err := parseFormInput(c, true)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": err.Error(),
		})
	}

	result, err := h.service.Update(c.Request().Context(), id, input)
	if err != nil {
		return writeError(c, err)
	}

	return c.JSON(http.StatusOK, result)
}

func (h *Handler) Delete(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": "ID Dalkot tidak valid",
		})
	}

	deleted, err := h.service.Delete(c.Request().Context(), id)
	if err != nil {
		return writeError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]any{
		"message": "Data Dalkot berhasil dihapus",
		"data":    deleted,
	})
}

func parseFormInput(
	c echo.Context,
	isUpdate bool,
) (FormInput, error) {
	serviceDate, err := time.Parse("2006-01-02", c.FormValue("serviceDate"))
	if err != nil {
		return FormInput{}, errors.New(
			"Tanggal Pelaksanaan Dinas tidak valid",
		)
	}

	var officers []OfficerInput
	if err := json.Unmarshal(
		[]byte(c.FormValue("officers")),
		&officers,
	); err != nil {
		return FormInput{}, errors.New("Data petugas tidak valid")
	}

	var document *multipart.FileHeader
	document, err = c.FormFile("document")

	if err != nil && !errors.Is(err, http.ErrMissingFile) {
		return FormInput{}, errors.New("Dokumentasi tidak dapat dibaca")
	}
	if errors.Is(err, http.ErrMissingFile) {
		document = nil
	}

	input := FormInput{
		ActiveEmail:         strings.TrimSpace(c.FormValue("activeEmail")),
		ServiceDate:         serviceDate,
		Category:            strings.TrimSpace(c.FormValue("category")),
		AccompaniedOfficial: strings.TrimSpace(c.FormValue("accompaniedOfficial")),
		DalkotType:          strings.TrimSpace(c.FormValue("dalkotType")),
		PolicyCategory:      strings.TrimSpace(c.FormValue("policyCategory")),
		ActivityName:        strings.TrimSpace(c.FormValue("activityName")),
		Location:            strings.TrimSpace(c.FormValue("location")),
		ReportContent:       strings.TrimSpace(c.FormValue("reportContent")),
		Status:              strings.TrimSpace(c.FormValue("status")),
		Officers:            officers,
		Document:            document,
	}

	_ = isUpdate
	return input, nil
}

func writeError(c echo.Context, err error) error {
	if IsNotFound(err) {
		return c.JSON(http.StatusNotFound, map[string]string{
			"message": ErrNotFound.Error(),
		})
	}

	message := err.Error()
	statusCode := http.StatusInternalServerError

	validationMessages := []string{
		"wajib",
		"tidak valid",
		"tidak boleh",
		"maksimal",
		"hanya boleh",
		"Pilih minimal",
	}

	for _, marker := range validationMessages {
		if strings.Contains(message, marker) {
			statusCode = http.StatusBadRequest
			break
		}
	}

	return c.JSON(statusCode, map[string]string{
		"message": message,
	})
}
