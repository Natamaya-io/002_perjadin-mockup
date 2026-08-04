package gup

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

func (h *Handler) List(c echo.Context) error {
	results, err := h.service.List(
		c.Request().Context(),
	)
	if err != nil {
		return handleServiceError(err)
	}

	return c.JSON(http.StatusOK, results)
}

func (h *Handler) GetByID(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"ID GUP tidak valid",
		)
	}

	result, err := h.service.GetByID(
		c.Request().Context(),
		id,
	)
	if err != nil {
		return handleServiceError(err)
	}

	return c.JSON(http.StatusOK, result)
}

func (h *Handler) GetMaster(c echo.Context) error {
	year := time.Now().Year()

	if value := strings.TrimSpace(
		c.QueryParam("year"),
	); value != "" {
		parsedYear, err := strconv.Atoi(value)
		if err != nil {
			return echo.NewHTTPError(
				http.StatusBadRequest,
				"Tahun tidak valid",
			)
		}

		year = parsedYear
	}

	result, err := h.service.GetMaster(
		c.Request().Context(),
		year,
	)
	if err != nil {
		return handleServiceError(err)
	}

	return c.JSON(http.StatusOK, result)
}

func (h *Handler) Create(c echo.Context) error {
	input, err := parseFormInput(c)
	if err != nil {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			err.Error(),
		)
	}

	result, err := h.service.Create(
		c.Request().Context(),
		input,
	)
	if err != nil {
		return handleServiceError(err)
	}

	return c.JSON(http.StatusCreated, result)
}

func (h *Handler) Update(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"ID GUP tidak valid",
		)
	}

	input, err := parseFormInput(c)
	if err != nil {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			err.Error(),
		)
	}

	result, err := h.service.Update(
		c.Request().Context(),
		id,
		input,
	)
	if err != nil {
		return handleServiceError(err)
	}

	return c.JSON(http.StatusOK, result)
}

func (h *Handler) Delete(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"ID GUP tidak valid",
		)
	}

	if err := h.service.Delete(
		c.Request().Context(),
		id,
	); err != nil {
		return handleServiceError(err)
	}

	return c.JSON(
		http.StatusOK,
		map[string]string{
			"message": "Data GUP berhasil dihapus",
		},
	)
}

func parseFormInput(
	c echo.Context,
) (FormInput, error) {
	var input FormInput

	if err := c.Request().ParseMultipartForm(
		MaxDocumentSize + (1 << 20),
	); err != nil {
		return input, errors.New(
			"form GUP tidak dapat dibaca",
		)
	}

	input.PaymentDescription = strings.TrimSpace(
		c.FormValue("paymentDescription"),
	)

	procurementTypeID, err := uuid.Parse(
		strings.TrimSpace(
			c.FormValue("procurementTypeId"),
		),
	)
	if err != nil {
		return input, errors.New(
			"jenis pengadaan tidak valid",
		)
	}

	input.ProcurementTypeID = procurementTypeID

	fundingValue := strings.TrimSpace(
		c.FormValue("fundingSourceId"),
	)

	if fundingValue != "" {
		fundingSourceID, err := uuid.Parse(
			fundingValue,
		)
		if err != nil {
			return input, errors.New(
				"sumber dana tidak valid",
			)
		}

		input.FundingSourceID = &fundingSourceID
	}

	input.ValueAmount, err = parseRupiah(
		c.FormValue("valueAmount"),
	)
	if err != nil {
		return input, errors.New(
			"nilai pengajuan tidak valid",
		)
	}

	input.PaidAmount, err = parseRupiah(
		c.FormValue("paidAmount"),
	)
	if err != nil {
		return input, errors.New(
			"jumlah dibayarkan tidak valid",
		)
	}

	input.TaxAmount, err = parseRupiah(
		c.FormValue("taxAmount"),
	)
	if err != nil {
		return input, errors.New(
			"nilai pajak tidak valid",
		)
	}

	receiptDate := strings.TrimSpace(
		c.FormValue("receiptDate"),
	)

	if receiptDate != "" {
		input.ReceiptDate, err = time.Parse(
			"2006-01-02",
			receiptDate,
		)
		if err != nil {
			return input, errors.New(
				"tanggal kwitansi tidak valid",
			)
		}
	}

	input.Recipient = strings.TrimSpace(
		c.FormValue("recipient"),
	)

	input.PUM = strings.TrimSpace(
		c.FormValue("pum"),
	)

	document, documentErr := c.FormFile("document")

	if errors.Is(documentErr, http.ErrMissingFile) {
		document, documentErr =
			c.FormFile("dokumen_gup")
	}

	if documentErr == nil {
		input.Document = document
	} else if !errors.Is(
		documentErr,
		http.ErrMissingFile,
	) {
		return input, errors.New(
			"dokumen GUP tidak dapat dibaca",
		)
	}

	return input, nil
}

func parseRupiah(value string) (int64, error) {
	cleanValue := strings.TrimSpace(value)

	replacer := strings.NewReplacer(
		"Rp", "",
		"rp", "",
		".", "",
		",", "",
		" ", "",
	)

	cleanValue = replacer.Replace(cleanValue)

	if cleanValue == "" {
		return 0, nil
	}

	return strconv.ParseInt(
		cleanValue,
		10,
		64,
	)
}

func handleServiceError(err error) error {
	if errors.Is(err, ErrNotFound) {
		return echo.NewHTTPError(
			http.StatusNotFound,
			err.Error(),
		)
	}

	var validationError *ValidationError

	if errors.As(err, &validationError) {
		return echo.NewHTTPError(
			http.StatusUnprocessableEntity,
			validationError.Message,
		)
	}

	return echo.NewHTTPError(
		http.StatusInternalServerError,
		"Terjadi kesalahan saat memproses data GUP",
	).SetInternal(err)
}