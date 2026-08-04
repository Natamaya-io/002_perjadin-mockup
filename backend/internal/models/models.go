package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Base model for UUID support
type Base struct {
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	ID        uuid.UUID `json:"id"`
}

type User struct {
	NomorHP      string `json:"nomorHp"`
	Email        string `json:"email"`
	Password     string `json:"-"`
	Name         string `json:"name"`
	Role         string `json:"role"`
	NIP          string `json:"nip"`
	Pangkat      string `json:"pangkat"`
	Golongan     string `json:"golongan"`
	Jabatan      string `json:"jabatan"`
	TingkatBiaya string `json:"tingkatBiaya"`
	SessionID    string `json:"-"`
	DemoPassword string `json:"-"`
	Base
}

type TravelRecord struct {
	Base
	EndDate          time.Time        `json:"endDate"`
	SuratTugasDate   time.Time        `json:"suratTugasDate"`
	StartDate        time.Time        `json:"startDate"`
	Report           *TravelReport    `json:"reportData,omitempty"`
	Cost             *TravelCost      `json:"costs,omitempty"`
	Employee         User             `json:"employee"`
	Creator          User             `json:"creator"`
	Type             string           `json:"type"`
	SuratTugasPath   string           `json:"suratTugasPath"`
	Province         string           `json:"province"`
	SPDNumber        string           `json:"spd"`
	SuratTugasNumber string           `json:"suratTugasNumber"`
	Purpose          string           `json:"purpose"`
	Stakeholder      string           `json:"stakeholder"`
	Agenda           string           `json:"agenda"`
	Status           string           `json:"status"`
	Location         string           `json:"location"`
	ReportStatus     string           `json:"reportStatus"`
	PaymentStatus    string           `json:"paymentStatus"`
	EmployeeIDs      []uuid.UUID      `json:"employeeIds,omitempty"`
	Locations        []TravelLocation `json:"locations,omitempty"`
	TotalCost        float64          `json:"totalCost"`
	CreatorID        uuid.UUID        `json:"creatorId"`
	EmployeeID       uuid.UUID        `json:"employeeId"`
	IsViewed         bool             `json:"isViewed"`
	SequenceNumber   int              `json:"sequenceNumber"`
}

type TravelLocation struct {
	StartDate time.Time `json:"startDate"`
	EndDate   time.Time `json:"endDate"`
	Location  string    `json:"location"`
	Province  string    `json:"province"`
	Base
	TravelRecordID uuid.UUID `json:"travelRecordId"`
}

type TravelCost struct {
	OtherCostDesc      string          `json:"otherCostDesc"`
	TransportMode      string          `json:"transportMode"`
	Details            json.RawMessage `json:"details"`
	AdditionalCosts    json.RawMessage `json:"additionalCosts"`
	TransportFile      json.RawMessage `json:"transportFile"`
	HotelFile          json.RawMessage `json:"hotelFile"`
	BoardingPassFile   json.RawMessage `json:"boardingPassFile"`
	TicketBackFile     json.RawMessage `json:"ticketBackFile"`
	TicketGoFile       json.RawMessage `json:"ticketGoFile"`
	ReceiptFiles       json.RawMessage `json:"receiptFiles"`
	HotelDays          int             `json:"hotelDays"`
	OtherCost          float64         `json:"otherCost"`
	TransportAmount    float64         `json:"transportAmount"`
	RegionalTransport  float64         `json:"regionalTransport"`
	LocalTransport     float64         `json:"localTransport"`
	HotelRate          float64         `json:"hotelRate"`
	DailyAllowanceRate float64         `json:"dailyAllowanceRate"`
	DailyAllowanceDays int             `json:"dailyAllowanceDays"`
	TicketBack         float64         `json:"ticketBack"`
	TicketGo           float64         `json:"ticketGo"`
	TravelRecordID     uuid.UUID       `json:"-"`
}

type TravelReport struct {
	SubmittedAt    time.Time       `json:"submittedAt"`
	Text           string          `json:"text"`
	PPKName        string          `json:"ppkName"`
	PPKNIP         string          `json:"ppkNip"`
	BendaharaName  string          `json:"bendaharaName"`
	BendaharaNIP   string          `json:"bendaharaNip"`
	Files          json.RawMessage `json:"files"`
	SppdFile       json.RawMessage `json:"sppdFile"`
	SuratTugasFile json.RawMessage `json:"suratTugasFile"`
	TanggalMerah   json.RawMessage `json:"tanggalMerah"`
	TravelRecordID uuid.UUID       `json:"-"`
}

// ==========================================
// Master Data Models (SBM & Reference)
// ==========================================

type Province struct {
	Name string `json:"name"`
	Code string `json:"code"`
	Base
}

type SBMRate struct {
	Province Province `json:"province"`
	Base
	InsideCityRate  float64   `json:"insideCityRate"`
	Year            int       `json:"year"`
	FullboardRate   float64   `json:"fullboardRate"`
	FullhalfRate    float64   `json:"fullhalfRate"`
	OutsideCityRate float64   `json:"outsideCityRate"`
	DiklatRate      float64   `json:"diklatRate"`
	HotelEchelon1   float64   `json:"hotelEchelon1"`
	HotelEchelon2   float64   `json:"hotelEchelon2"`
	HotelEchelon3   float64   `json:"hotelEchelon3"`
	HotelEchelon4   float64   `json:"hotelEchelon4"`
	HotelStaff      float64   `json:"hotelStaff"`
	TaxiRate        float64   `json:"taxiRate"`
	ProvinceID      uuid.UUID `json:"provinceId"`
}

// ==========================================
// Dashboard Aggregation Models
// ==========================================

type DashboardSummary struct {
	RecentRecords    []TravelRecord    `json:"recentRecords"`
	Budgets          []DashboardBudget `json:"budgets"`
	TotalTrips       int64             `json:"totalTrips"`
	ActiveTrips      int64             `json:"activeTrips"`
	StatusCompleted  int64             `json:"statusCompleted"`
	StatusInProgress int64             `json:"statusInProgress"`
	StatusAssigned   int64             `json:"statusAssigned"`
	StatusRejected   int64             `json:"statusRejected"`
	ReportCompleted  int64             `json:"reportCompleted"`
	ReportPending    int64             `json:"reportPending"`
}

type DashboardBudget struct {
	Year  int     `json:"year"`
	Month int     `json:"month"`
	Total float64 `json:"total"`
}

type PaginatedParams struct {
	StartDate     *time.Time `query:"start_date"`
	EndDate       *time.Time `query:"end_date"`
	UserID        *uuid.UUID `query:"user_id"`
	Cursor        string     `query:"cursor"`
	Search        string     `query:"search"`
	Status        string     `query:"status"`
	ReportStatus  string     `query:"report_status"`
	PaymentStatus string     `query:"payment_status"`
	SortBy        string     `query:"sort_by"`
	Limit         int        `query:"limit"`
}

type PaginatedResponse struct {
	NextCursor   string         `json:"nextCursor"`
	Data         []TravelRecord `json:"data"`
	TotalItems   int64          `json:"totalItems"`
	TotalRecords int64          `json:"totalRecords"`
	Limit        int            `json:"limit"`
}
