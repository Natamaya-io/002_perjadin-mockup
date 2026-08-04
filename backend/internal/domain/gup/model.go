package gup

import (
	"errors"
	"mime/multipart"
	"time"

	"github.com/google/uuid"
)

const MaxDocumentSize int64 = 5 * 1024 * 1024

var ErrNotFound = errors.New("data GUP tidak ditemukan")

type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

type Transaction struct {
	ID                   uuid.UUID  `json:"id"`
	BusinessID           string     `json:"businessId"`
	PaymentDescription   string     `json:"paymentDescription"`
	ProcurementTypeID    uuid.UUID  `json:"procurementTypeId"`
	ProcurementTypeName  string     `json:"procurementTypeName"`
	AccountCode          string     `json:"accountCode"`
	MAK                  string     `json:"mak"`
	FundingSourceID      *uuid.UUID `json:"fundingSourceId,omitempty"`
	FundingSourceLabel   string     `json:"fundingSourceLabel"`
	ValueAmount          int64      `json:"valueAmount"`
	PaidAmount           int64      `json:"paidAmount"`
	TaxAmount            int64      `json:"taxAmount"`
	DifferenceAmount     int64      `json:"differenceAmount"`
	ReceiptDate          string     `json:"receiptDate"`
	Recipient            string     `json:"recipient"`
	PUM                  string     `json:"pum"`
	DocumentFilename     string     `json:"-"`
	DocumentOriginalName string     `json:"documentOriginalName"`
	DocumentMIME         string     `json:"documentMime"`
	DocumentSize         int64      `json:"documentSize"`
	DocumentURL          string     `json:"documentUrl"`
	CreatedAt            time.Time  `json:"createdAt"`
	UpdatedAt            time.Time  `json:"updatedAt"`
}

type ProcurementType struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	AccountCode string    `json:"accountCode"`
	MAK         string    `json:"mak"`
}

type FundingSource struct {
	ID          uuid.UUID `json:"id"`
	Year        int       `json:"year"`
	MonthNumber int       `json:"monthNumber"`
	MonthName   string    `json:"monthName"`
	GUPLabel    string    `json:"gupLabel"`
}

type MasterData struct {
	ProcurementTypes []ProcurementType `json:"procurementTypes"`
	FundingSources   []FundingSource   `json:"fundingSources"`
}

type FormInput struct {
	PaymentDescription string
	ProcurementTypeID  uuid.UUID
	FundingSourceID    *uuid.UUID
	ValueAmount        int64
	PaidAmount         int64
	TaxAmount          int64
	ReceiptDate        time.Time
	Recipient          string
	PUM                string
	Document           *multipart.FileHeader
}

type DocumentMetadata struct {
	Filename     string
	OriginalName string
	MIME         string
	Size         int64
}

type SaveInput struct {
	BusinessID          string
	PaymentDescription  string
	ProcurementTypeID   uuid.UUID
	FundingSourceID     *uuid.UUID
	ValueAmount         int64
	PaidAmount          int64
	TaxAmount           int64
	ReceiptDate         time.Time
	Recipient           string
	PUM                 string
	DocumentFilename    string
	DocumentOriginal    string
	DocumentMIME        string
	DocumentSize        int64
}