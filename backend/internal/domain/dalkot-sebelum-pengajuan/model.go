package dalkot

import (
	"errors"
	"mime/multipart"
	"time"

	"github.com/google/uuid"
)

const (
	TypeSPJRill           = "SPJ_RILL"
	TypeKebijakanProtokol = "KEBIJAKAN_PROTOKOL"

	PolicyJamKerja  = "JAM_KERJA"
	PolicyOvertime  = "OVERTIME"
	PolicyHariLibur = "HARI_LIBUR"

	StatusDraft      = "DRAFT"
	StatusDiajukan   = "DIAJUKAN"
	StatusSelesai    = "SELESAI"
	StatusDibatalkan = "DIBATALKAN"

	SPJRillAmount   int64 = 170000
	JamKerjaAmount  int64 = 100000
	OvertimeAmount  int64 = 150000
	HariLiburAmount int64 = 250000

	MaxDocumentSize int64 = 5 * 1024 * 1024
)

var ErrNotFound = errors.New("data Dalkot tidak ditemukan")

type Officer struct {
	ID         uuid.UUID  `json:"id"`
	DalkotID   uuid.UUID  `json:"dalkotId"`
	EmployeeID *uuid.UUID `json:"employeeId,omitempty"`
	Name       string     `json:"name"`
	NIP        string     `json:"nip"`
	Position   string     `json:"position"`
	Amount     int64      `json:"amount"`
	SortOrder  int        `json:"sortOrder"`
	CreatedAt  time.Time  `json:"createdAt"`
}

type OfficerMaster struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	NIP       string    `json:"nip"`
	Position  string    `json:"position"`
	SortOrder int       `json:"sortOrder"`
}

type MasterOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

type MasterData struct {
	Officers         []OfficerMaster `json:"officers"`
	Categories       []MasterOption  `json:"categories"`
	Officials        []MasterOption  `json:"officials"`
	Locations        []MasterOption  `json:"locations"`
	PolicyCategories []MasterOption  `json:"policyCategories"`
	SPJRillAmount    int64           `json:"spjRillAmount"`
	JamKerjaAmount   int64           `json:"jamKerjaAmount"`
	OvertimeAmount   int64           `json:"overtimeAmount"`
	HariLiburAmount  int64           `json:"hariLiburAmount"`
}

type Record struct {
	ID                   uuid.UUID `json:"id"`
	BusinessID           string    `json:"businessId"`
	ActiveEmail          string    `json:"activeEmail"`
	ServiceDate          string    `json:"serviceDate"`
	Category             string    `json:"category"`
	AccompaniedOfficial  string    `json:"accompaniedOfficial"`
	DalkotType           string    `json:"dalkotType"`
	PolicyCategory       string    `json:"policyCategory"`
	ActivityName         string    `json:"activityName"`
	Location             string    `json:"location"`
	ReportContent        string    `json:"reportContent"`
	Status               string    `json:"status"`
	TotalAmount          int64     `json:"totalAmount"`
	DocumentFilename     string    `json:"documentFilename"`
	DocumentOriginalName string    `json:"documentOriginalName"`
	DocumentMIME         string    `json:"documentMime"`
	DocumentSize         int64     `json:"documentSize"`
	DocumentURL          string    `json:"documentUrl"`
	OfficerCount         int       `json:"officerCount"`
	Officers             []Officer `json:"officers"`
	CreatedAt            time.Time `json:"createdAt"`
	UpdatedAt            time.Time `json:"updatedAt"`
}

type Summary struct {
	TotalRecords   int   `json:"totalRecords"`
	TotalOfficers  int   `json:"totalOfficers"`
	TotalAmount    int64 `json:"totalAmount"`
	PendingRecords int   `json:"pendingRecords"`
}

type ListResponse struct {
	Rows    []Record `json:"rows"`
	Summary Summary  `json:"summary"`
}

type OfficerInput struct {
	EmployeeID *uuid.UUID `json:"employeeId,omitempty"`
	Name       string     `json:"name"`
	NIP        string     `json:"nip"`
	Position   string     `json:"position"`
}

type FormInput struct {
	ActiveEmail         string
	ServiceDate         time.Time
	Category            string
	AccompaniedOfficial string
	DalkotType          string
	PolicyCategory      string
	ActivityName        string
	Location            string
	ReportContent       string
	Status              string
	Officers            []OfficerInput
	Document            *multipart.FileHeader
}

type SaveInput struct {
	BusinessID           string
	ActiveEmail          string
	ServiceDate          time.Time
	Category             string
	AccompaniedOfficial  string
	DalkotType           string
	PolicyCategory       string
	ActivityName         string
	Location             string
	ReportContent        string
	Status               string
	TotalAmount          int64
	DocumentFilename     string
	DocumentOriginalName string
	DocumentMIME         string
	DocumentSize         int64
	Officers             []OfficerSaveInput
}

type OfficerSaveInput struct {
	EmployeeID *uuid.UUID
	Name       string
	NIP        string
	Position   string
	Amount     int64
	SortOrder  int
}

type DocumentMetadata struct {
	Filename     string
	OriginalName string
	MIME         string
	Size         int64
}
