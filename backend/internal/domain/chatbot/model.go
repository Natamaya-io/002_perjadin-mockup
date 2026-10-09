package chatbot

import "github.com/google/uuid"

// PeriodFilter represents an analytics period. Month 0 means the whole year.
type PeriodFilter struct {
	Year  int `json:"year"`
	Month int `json:"month"`
}

type PeriodInfo struct {
	Year  int    `json:"year"`
	Month int    `json:"month"`
	Label string `json:"label"`
}

type GUPSummary struct {
	TotalLS               int64   `json:"totalLs"`
	TotalRealization      int64   `json:"totalRealization"`
	Balance               int64   `json:"balance"`
	RealizationPercentage float64 `json:"realizationPercentage"`
	TransactionCount      int64   `json:"transactionCount"`
	TotalTax              int64   `json:"totalTax"`
}

type GUPExpenseRank struct {
	Rank                int    `json:"rank"`
	BusinessID          string `json:"businessId"`
	PaymentDescription  string `json:"paymentDescription"`
	ProcurementTypeName string `json:"procurementTypeName"`
	AccountCode         string `json:"accountCode"`
	MAK                 string `json:"mak"`
	FundingSourceLabel  string `json:"fundingSourceLabel"`
	Recipient           string `json:"recipient"`
	ReceiptDate         string `json:"receiptDate"`
	PaidAmount          int64  `json:"paidAmount"`
}

type GUPAccountRank struct {
	Rank             int     `json:"rank"`
	AccountCode      string  `json:"accountCode"`
	MAK              string  `json:"mak"`
	PaidAmount       int64   `json:"paidAmount"`
	TransactionCount int64   `json:"transactionCount"`
	Percentage       float64 `json:"percentage"`
}

type GUPMonthlyRow struct {
	MonthNumber      int    `json:"monthNumber"`
	MonthName        string `json:"monthName"`
	GUPLabel         string `json:"gupLabel"`
	Income           int64  `json:"income"`
	Realization      int64  `json:"realization"`
	Balance          int64  `json:"balance"`
	TransactionCount int64  `json:"transactionCount"`
	Status           string `json:"status"`
}

type GUPAnalysis struct {
	Period         PeriodInfo       `json:"period"`
	Summary        GUPSummary       `json:"summary"`
	TopExpenses    []GUPExpenseRank `json:"topExpenses"`
	LowestExpenses []GUPExpenseRank `json:"lowestExpenses"`
	AccountRanking []GUPAccountRank `json:"accountRanking"`
	LargestAccount *GUPAccountRank  `json:"largestAccount,omitempty"`
	Monthly        []GUPMonthlyRow  `json:"monthly"`
}

type TripStatusCount struct {
	Assigned   int64 `json:"assigned"`
	InProgress int64 `json:"inProgress"`
	Completed  int64 `json:"completed"`
	Rejected   int64 `json:"rejected"`
}

type OutsideCityUserRow struct {
	UserID        uuid.UUID       `json:"userId"`
	Name          string          `json:"name"`
	Email         string          `json:"email"`
	NIP           string          `json:"nip"`
	Position      string          `json:"position"`
	TotalTrips    int64           `json:"totalTrips"`
	ExecutedTrips int64           `json:"executedTrips"`
	ExecutedCost  int64           `json:"executedCost"`
	Statuses      TripStatusCount `json:"statuses"`
	Rank          int             `json:"rank"`
}

type OutsideCitySummary struct {
	ProtocolUsers int64           `json:"protocolUsers"`
	TotalTrips    int64           `json:"totalTrips"`
	ExecutedTrips int64           `json:"executedTrips"`
	ExecutedCost  int64           `json:"executedCost"`
	Statuses      TripStatusCount `json:"statuses"`
}

type OutsideCityAnalysis struct {
	Period  PeriodInfo           `json:"period"`
	Summary OutsideCitySummary   `json:"summary"`
	Users   []OutsideCityUserRow `json:"users"`
}

type DalkotStatusCount struct {
	Draft   int64 `json:"draft"`
	Pending int64 `json:"pending"`
	Success int64 `json:"success"`
}

type InsideCityUserRow struct {
	UserID             uuid.UUID         `json:"userId"`
	Name               string            `json:"name"`
	Email              string            `json:"email"`
	NIP                string            `json:"nip"`
	Position           string            `json:"position"`
	TotalActivities    int64             `json:"totalActivities"`
	ExecutedActivities int64             `json:"executedActivities"`
	TotalSPJ           int64             `json:"totalSpj"`
	TotalRIIL          int64             `json:"totalRiil"`
	Balance            int64             `json:"balance"`
	Statuses           DalkotStatusCount `json:"statuses"`
	Rank               int               `json:"rank"`
}

type InsideCitySummary struct {
	ProtocolUsers      int64             `json:"protocolUsers"`
	TotalActivities    int64             `json:"totalActivities"`
	ExecutedActivities int64             `json:"executedActivities"`
	TotalSPJ           int64             `json:"totalSpj"`
	TotalRIIL          int64             `json:"totalRiil"`
	Balance            int64             `json:"balance"`
	Statuses           DalkotStatusCount `json:"statuses"`
}

type InsideCityAnalysis struct {
	Period  PeriodInfo          `json:"period"`
	Summary InsideCitySummary   `json:"summary"`
	Users   []InsideCityUserRow `json:"users"`
}

// ChatMessageRequest is the single entry point used by the conversational UI.
// Financial data is never loaded until the user sends a command.
type ChatMessageRequest struct {
	Message string `json:"message"`
}

// ChatMessageResponse tells the frontend which result renderer to use.
// Data contains one of GUPAnalysis, OutsideCityAnalysis, InsideCityAnalysis, or nil for menu/help replies.
type ChatMessageResponse struct {
	Intent      string      `json:"intent"`
	Message     string      `json:"message"`
	Data        interface{} `json:"data,omitempty"`
	Suggestions []string    `json:"suggestions,omitempty"`
}
