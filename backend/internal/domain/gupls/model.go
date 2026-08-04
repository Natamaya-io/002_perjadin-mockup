package gupls

import "github.com/google/uuid"

const LSAccountCode = "S.522141"

type Row struct {
	FundingSourceID  uuid.UUID `json:"fundingSourceId"`
	Year             int       `json:"year"`
	MonthNumber      int       `json:"monthNumber"`
	MonthName        string    `json:"monthName"`
	GUPLabel         string    `json:"gupLabel"`
	AccountCode      string    `json:"accountCode"`
	Amount           int64     `json:"amount"`
	Realization      int64     `json:"realization"`
	Remaining        int64     `json:"remaining"`
	TransactionCount int64     `json:"transactionCount"`
}

type Summary struct {
	TotalAmount      int64 `json:"totalAmount"`
	TotalRealization int64 `json:"totalRealization"`
	TotalRemaining   int64 `json:"totalRemaining"`
	FilledMonths     int   `json:"filledMonths"`
	TotalMonths      int   `json:"totalMonths"`
}

type Response struct {
	Year    int     `json:"year"`
	Rows    []Row   `json:"rows"`
	Summary Summary `json:"summary"`
}

type SaveRequest struct {
	Year            int       `json:"year"`
	FundingSourceID uuid.UUID `json:"fundingSourceId"`
	Amount          int64     `json:"amount"`
}
