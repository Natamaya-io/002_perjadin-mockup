package dalkot

import (
	"encoding/json"
	"errors"
	"mime/multipart"
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

func (h *Handler) List(c echo.Context) error {
	protocolUserID, err :=
		protocolUserIDFromContext(c)

	if err != nil {
		return err
	}

	result, err := h.service.List(
		c.Request().Context(),
		protocolUserID,
	)
	if err != nil {
		return writeError(c, err)
	}

	return c.JSON(
		http.StatusOK,
		result,
	)
}

func (h *Handler) GetRecap(c echo.Context) error {
	filter, err := parseRecapFilter(c, true)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": err.Error()})
	}

	result, err := h.service.GetRecap(c.Request().Context(), filter)
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) GetMonthlySPJRecap(c echo.Context) error {
	filter, err := parseRecapFilter(c, false)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": err.Error()})
	}

	result, err := h.service.GetMonthlySPJRecap(c.Request().Context(), filter)
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

func parseRecapFilter(c echo.Context, includeOfficer bool) (RecapFilter, error) {
	now := time.Now()
	month := int(now.Month())
	year := now.Year()

	if value := strings.TrimSpace(c.QueryParam("month")); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			return RecapFilter{}, errors.New("Bulan rekap Dalkot tidak valid")
		}
		month = parsed
	}
	if value := strings.TrimSpace(c.QueryParam("year")); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			return RecapFilter{}, errors.New("Tahun rekap Dalkot tidak valid")
		}
		year = parsed
	}

	filter := RecapFilter{Month: month, Year: year}
	if includeOfficer {
		if value := strings.TrimSpace(c.QueryParam("officerId")); value != "" {
			id, err := uuid.Parse(value)
			if err != nil || id == uuid.Nil {
				return RecapFilter{}, errors.New("ID Petugas Dalkot tidak valid")
			}
			filter.OfficerID = &id
		}
	}

	return filter, nil
}

func (h *Handler) GetMaster(c echo.Context) error {
	result, err := h.service.GetMaster(c.Request().Context())
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) GetByID(
	c echo.Context,
) error {
	id, err := parseDalkotID(
		c.Param("id"),
	)
	if err != nil {
		return c.JSON(
			http.StatusBadRequest,
			map[string]string{
				"message": err.Error(),
			},
		)
	}

	protocolUserID, err :=
		protocolUserIDFromContext(c)

	if err != nil {
		return err
	}

	result, err := h.service.GetByID(
		c.Request().Context(),
		id,
		protocolUserID,
	)
	if err != nil {
		return writeError(c, err)
	}

	return c.JSON(
		http.StatusOK,
		result,
	)
}

func (h *Handler) Create(
	c echo.Context,
) error {
	protocolUserID, err :=
		protocolUserIDFromContext(c)

	if err != nil {
		return err
	}

	input, err := parseFormInput(c)
	if err != nil {
		return c.JSON(
			http.StatusBadRequest,
			map[string]string{
				"message": err.Error(),
			},
		)
	}

	result, err := h.service.Create(
		c.Request().Context(),
		input,
		protocolUserID,
	)
	if err != nil {
		return writeError(c, err)
	}

	return c.JSON(
		http.StatusCreated,
		result,
	)
}

func (h *Handler) Update(
	c echo.Context,
) error {
	id, err := parseDalkotID(
		c.Param("id"),
	)
	if err != nil {
		return c.JSON(
			http.StatusBadRequest,
			map[string]string{
				"message": err.Error(),
			},
		)
	}

	input, err := parseFormInput(c)
	if err != nil {
		return c.JSON(
			http.StatusBadRequest,
			map[string]string{
				"message": err.Error(),
			},
		)
	}

	protocolUserID, err :=
		protocolUserIDFromContext(c)

	if err != nil {
		return err
	}

	result, err := h.service.Update(
		c.Request().Context(),
		id,
		input,
		protocolUserID,
	)
	if err != nil {
		return writeError(c, err)
	}

	return c.JSON(
		http.StatusOK,
		result,
	)
}

func (h *Handler) Delete(c echo.Context) error {
	id, err := parseDalkotID(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": err.Error()})
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

func protocolUserIDFromContext(
	c echo.Context,
) (*uuid.UUID, error) {
	role, ok := c.Get("role").(string)
	if !ok || strings.TrimSpace(role) == "" {
		return nil, echo.NewHTTPError(
			http.StatusUnauthorized,
			"Sesi pengguna tidak valid",
		)
	}

	if role != "protokol" {
		// Nil berarti admin dan kasubag
		// tidak mendapat filter user.
		return nil, nil
	}

	userIDText, ok :=
		c.Get("user_id").(string)

	if !ok {
		return nil, echo.NewHTTPError(
			http.StatusUnauthorized,
			"ID pengguna tidak valid",
		)
	}

	userID, err := uuid.Parse(
		strings.TrimSpace(userIDText),
	)

	if err != nil || userID == uuid.Nil {
		return nil, echo.NewHTTPError(
			http.StatusUnauthorized,
			"ID pengguna tidak valid",
		)
	}

	return &userID, nil
}

func parseDalkotID(value string) (uuid.UUID, error) {
	id, err := uuid.Parse(strings.TrimSpace(value))
	if err != nil || id == uuid.Nil {
		return uuid.Nil, errors.New("ID Dalkot tidak valid")
	}
	return id, nil
}

func parseFormInput(c echo.Context) (FormInput, error) {
	serviceDate, err := parseRequiredDate(
		c.FormValue("serviceDate"),
		"Tanggal Pelaksanaan Dinas",
	)
	if err != nil {
		return FormInput{}, err
	}

	assignmentLetterDate, err := parseOptionalDate(
		c.FormValue("assignmentLetterDate"),
		"Tanggal Surat Tugas",
	)
	if err != nil {
		return FormInput{}, err
	}

	spjOfficerIDs, err := parseUUIDListFormValue(
		c.FormValue("spjOfficerIds"),
		"Data Petugas SPJ tidak valid",
	)
	if err != nil {
		return FormInput{}, err
	}

	// riilOfficerIds adalah nama field final. rillOfficerIds tetap dibaca
	// sebagai kompatibilitas untuk form lama agar data tidak langsung rusak.
	riilValue := c.FormValue("riilOfficerIds")
	if strings.TrimSpace(riilValue) == "" {
		riilValue = c.FormValue("rillOfficerIds")
	}
	riilOfficerIDs, err := parseUUIDListFormValue(
		riilValue,
		"Data Petugas Rill tidak valid",
	)
	if err != nil {
		return FormInput{}, err
	}

	document, err := readOptionalDocument(c)
	if err != nil {
		return FormInput{}, err
	}

	return FormInput{
		ServiceDate:            serviceDate,
		Category:               strings.TrimSpace(c.FormValue("category")),
		AccompaniedOfficial:    strings.TrimSpace(c.FormValue("accompaniedOfficial")),
		AssignmentLetterNumber: strings.TrimSpace(c.FormValue("assignmentLetterNumber")),
		AssignmentLetterDate:   assignmentLetterDate,
		ActivityName:           strings.TrimSpace(c.FormValue("activityName")),
		Location:               strings.TrimSpace(c.FormValue("location")),
		ReportContent:          strings.TrimSpace(c.FormValue("reportContent")),
		Status:                 strings.TrimSpace(c.FormValue("status")),
		SPJOfficerIDs:          spjOfficerIDs,
		RiilOfficerIDs:         riilOfficerIDs,
		Document:               document,
	}, nil
}

func parseRequiredDate(value, label string) (time.Time, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return time.Time{}, errors.New(label + " wajib diisi")
	}
	parsed, err := time.Parse("2006-01-02", trimmed)
	if err != nil {
		return time.Time{}, errors.New(label + " tidak valid")
	}
	return parsed, nil
}

func parseOptionalDate(value, label string) (*time.Time, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil, nil
	}
	parsed, err := time.Parse("2006-01-02", trimmed)
	if err != nil {
		return nil, errors.New(label + " tidak valid")
	}
	return &parsed, nil
}

func parseUUIDListFormValue(value, errorMessage string) ([]uuid.UUID, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return []uuid.UUID{}, nil
	}

	var ids []uuid.UUID
	if err := json.Unmarshal([]byte(trimmed), &ids); err != nil {
		return nil, errors.New(errorMessage)
	}
	for _, id := range ids {
		if id == uuid.Nil {
			return nil, errors.New(errorMessage)
		}
	}
	return ids, nil
}

func readOptionalDocument(c echo.Context) (*multipart.FileHeader, error) {
	document, err := c.FormFile("document")
	if errors.Is(err, http.ErrMissingFile) {
		return nil, nil
	}
	if err != nil {
		return nil, errors.New("Dokumentasi tidak dapat dibaca")
	}
	return document, nil
}

func writeError(c echo.Context, err error) error {
	switch {
	case IsNotFound(err):
		return c.JSON(http.StatusNotFound, map[string]string{"message": ErrNotFound.Error()})
	case IsInvalidInput(err):
		return c.JSON(http.StatusBadRequest, map[string]string{"message": invalidInputMessage(err)})
	}

	var httpError *echo.HTTPError
	if errors.As(err, &httpError) {
		message := http.StatusText(httpError.Code)
		if text, ok := httpError.Message.(string); ok && strings.TrimSpace(text) != "" {
			message = text
		}
		return c.JSON(httpError.Code, map[string]string{"message": message})
	}

	c.Logger().Error(err)
	return c.JSON(http.StatusInternalServerError, map[string]string{
		"message": "Terjadi kesalahan pada server",
	})
}

func invalidInputMessage(err error) string {
	message := strings.TrimSpace(err.Error())
	prefix := ErrInvalidInput.Error() + ":"
	message = strings.TrimSpace(strings.TrimPrefix(message, prefix))
	if message == "" {
		return ErrInvalidInput.Error()
	}
	return message
}
