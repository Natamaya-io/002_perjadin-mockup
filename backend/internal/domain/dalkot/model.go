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

	CategoryJamKerja  = "Jam Kerja"
	CategoryOvertime  = "Overtime"
	CategoryHariLibur = "Hari Libur"

	OfficerGroupSPJ  = "SPJ"
	OfficerGroupRiil = "RIIL"

	StatusDraft   = "Draft"
	StatusPending = "Pending"
	StatusSukses  = "Sukses"

	SPJAmount       int64 = 170000
	JamKerjaAmount  int64 = 100000
	OvertimeAmount  int64 = 150000
	HariLiburAmount int64 = 250000

	MonthlySPJAccount = "026.01.WA.2158.EBA.994.002.S.524113"
	MonthlySPJTitle   = "Penerimaan biaya perjalanan dinas dalam kota protokol yang telah melaksanakan persiapan kegiatan Menteri Ketenagakerjaan dan Wakil Menteri Ketenagakerjaan RI"

	MaxDocumentSize int64 = 5 * 1024 * 1024
)

var (
	ErrNotFound     = errors.New("data Dalkot tidak ditemukan")
	ErrInvalidInput = errors.New("data Pengajuan Dalkot tidak valid")
)

type UserOption struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	NIP      string    `json:"nip"`
	Position string    `json:"position"`
	Grade    string    `json:"grade"`
	Role     string    `json:"role"`
	IsActive bool      `json:"isActive"`
}

type MasterOption struct {
	Value  string `json:"value"`
	Label  string `json:"label"`
	Amount int64  `json:"amount,omitempty"`
}

type MasterData struct {
	Users           []UserOption   `json:"users"`
	Categories      []MasterOption `json:"categories"`
	Officials       []MasterOption `json:"officials"`
	Locations       []MasterOption `json:"locations"`
	SPJAmount       int64          `json:"spjAmount"`
	JamKerjaAmount  int64          `json:"jamKerjaAmount"`
	OvertimeAmount  int64          `json:"overtimeAmount"`
	HariLiburAmount int64          `json:"hariLiburAmount"`
}

type Officer struct {
	ID           uuid.UUID  `json:"id"`
	DalkotID     uuid.UUID  `json:"dalkotId"`
	EmployeeID   *uuid.UUID `json:"employeeId,omitempty"`
	OfficerGroup string     `json:"officerGroup"`
	Name         string     `json:"name"`
	NIP          string     `json:"nip"`
	Position     string     `json:"position"`
	Grade        string     `json:"grade"`
	Amount       int64      `json:"amount"`
	Status       string     `json:"status"`
	SortOrder    int        `json:"sortOrder"`
	CreatedAt    time.Time  `json:"createdAt"`
}

type Record struct {
	ID                     uuid.UUID `json:"id"`
	BusinessID             string    `json:"businessId"`
	ActiveEmail            string    `json:"-"`
	ServiceDate            string    `json:"serviceDate"`
	Category               string    `json:"category"`
	AccompaniedOfficial    string    `json:"accompaniedOfficial"`
	DalkotType             string    `json:"dalkotType"`
	AssignmentLetterNumber string    `json:"assignmentLetterNumber"`
	AssignmentLetterDate   *string   `json:"assignmentLetterDate"`
	ActivityName           string    `json:"activityName"`
	Location               string    `json:"location"`
	ReportContent          string    `json:"reportContent"`
	Status                 string    `json:"status"`
	TotalSPJAmount         int64     `json:"totalSPJAmount"`
	TotalRiilAmount        int64     `json:"totalRiilAmount"`
	DocumentFilename       string    `json:"documentFilename"`
	DocumentOriginalName   string    `json:"documentOriginalName"`
	DocumentMIME           string    `json:"documentMime"`
	DocumentSize           int64     `json:"documentSize"`
	DocumentURL            string    `json:"documentUrl"`
	OfficerCount           int       `json:"officerCount"`
	Officers               []Officer `json:"officers"`
	CreatedAt              time.Time `json:"createdAt"`
	UpdatedAt              time.Time `json:"updatedAt"`
}

type Summary struct {
	TotalRecords    int   `json:"totalRecords"`
	TotalOfficers   int   `json:"totalOfficers"`
	TotalSPJAmount  int64 `json:"totalSPJAmount"`
	TotalRiilAmount int64 `json:"totalRiilAmount"`
}

type ListResponse struct {
	Rows    []Record `json:"rows"`
	Summary Summary  `json:"summary"`
}

type RecapFilter struct {
	Month     int
	Year      int
	OfficerID *uuid.UUID
}

type RecapRow struct {
	BusinessID          string `json:"businessId"`
	ServiceDate         string `json:"serviceDate"`
	AccompaniedOfficial string `json:"accompaniedOfficial"`
	ActivityName        string `json:"activityName"`
	OfficerName         string `json:"officerName"`
	Category            string `json:"category"`
	DalkotType          string `json:"dalkotType"`
	Nominal             int64  `json:"nominal"`
}

type RecapSummary struct {
	TotalRows       int   `json:"totalRows"`
	TotalNominal    int64 `json:"totalNominal"`
	TotalSPJAmount  int64 `json:"totalSPJAmount"`
	TotalRiilAmount int64 `json:"totalRiilAmount"`
	Balance         int64 `json:"balance"`
	BalanceMinus    bool  `json:"balanceMinus"`
}

type RecapResponse struct {
	Month       int          `json:"month"`
	Year        int          `json:"year"`
	OfficerID   *uuid.UUID   `json:"officerId,omitempty"`
	OfficerName string       `json:"officerName"`
	Rows        []RecapRow   `json:"rows"`
	Summary     RecapSummary `json:"summary"`
}

type MonthlySPJRecapRow struct {
	OfficerID               *uuid.UUID `json:"officerId,omitempty"`
	Name                    string     `json:"name"`
	Grade                   string     `json:"grade"`
	Dates                   []string   `json:"dates"`
	Rate                    int64      `json:"rate"`
	TotalDays               int        `json:"totalDays"`
	TotalAmount             int64      `json:"totalAmount"`
	AssignmentLetterNumbers []string   `json:"assignmentLetterNumbers"`
	AssignmentLetterDates   []string   `json:"assignmentLetterDates"`
}

type MonthlySPJRecapSummary struct {
	TotalOfficers int   `json:"totalOfficers"`
	TotalDays     int   `json:"totalDays"`
	TotalAmount   int64 `json:"totalAmount"`
}

type MonthlySPJRecapResponse struct {
	Month   int                    `json:"month"`
	Year    int                    `json:"year"`
	Account string                 `json:"account"`
	Title   string                 `json:"title"`
	Rows    []MonthlySPJRecapRow   `json:"rows"`
	Summary MonthlySPJRecapSummary `json:"summary"`
}

type FormInput struct {
	ServiceDate            time.Time
	Category               string
	AccompaniedOfficial    string
	AssignmentLetterNumber string
	AssignmentLetterDate   *time.Time
	ActivityName           string
	Location               string
	ReportContent          string
	Status                 string
	SPJOfficerIDs          []uuid.UUID
	RiilOfficerIDs         []uuid.UUID
	Document               *multipart.FileHeader
}

type SaveInput struct {
	BusinessID             string
	ServiceDate            time.Time
	Category               string
	AccompaniedOfficial    string
	DalkotType             string
	PolicyCategory         string
	AssignmentLetterNumber string
	AssignmentLetterDate   *time.Time
	ActivityName           string
	Location               string
	ReportContent          string
	Status                 string
	TotalSPJAmount         int64
	TotalRiilAmount        int64
	DocumentFilename       string
	DocumentOriginalName   string
	DocumentMIME           string
	DocumentSize           int64
	Officers               []OfficerSaveInput
}

type OfficerSaveInput struct {
	EmployeeID   uuid.UUID
	OfficerGroup string
	Name         string
	NIP          string
	Position     string
	Grade        string
	Amount       int64
	Status       string
	SortOrder    int
}

type DocumentMetadata struct {
	Filename     string
	OriginalName string
	MIME         string
	Size         int64
}
