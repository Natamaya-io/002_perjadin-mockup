package gupreport

import "github.com/google/uuid"

type ReportRow struct {
	ProcurementTypeID   uuid.UUID `json:"procurementTypeId"`
	ProcurementTypeName string    `json:"procurementTypeName"`
	AccountCode         string    `json:"accountCode"`
	MAK                 string    `json:"mak"`
	TransactionCount    int64     `json:"transactionCount"`
	Realization         int64     `json:"realization"`
	SubmissionValue     int64     `json:"submissionValue"`
	Budget              int64     `json:"budget"`
	Remaining           int64     `json:"remaining"`
	Absorption          float64   `json:"absorption"`
}

type Summary struct {
	TotalBudget          int64   `json:"totalBudget"`
	TotalRealization     int64   `json:"totalRealization"`
	TotalRemaining       int64   `json:"totalRemaining"`
	TotalAbsorption      float64 `json:"totalAbsorption"`
	TotalTransactions    int64   `json:"totalTransactions"`
	TotalSubmissionValue int64   `json:"totalSubmissionValue"`
	TotalProcurementType int     `json:"totalProcurementType"`
}

type ReportResponse struct {
	Year    int         `json:"year"`
	Rows    []ReportRow `json:"rows"`
	Summary Summary     `json:"summary"`
}

type SaveBudgetRequest struct {
	Year              int       `json:"year"`
	ProcurementTypeID uuid.UUID `json:"procurementTypeId"`
	Amount            int64     `json:"amount"`
}
