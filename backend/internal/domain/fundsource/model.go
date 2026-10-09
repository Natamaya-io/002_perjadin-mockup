package fundsource

import (
	"time"

	"github.com/google/uuid"
)

type Source struct {
	ID     uuid.UUID `json:"id"`
	Year   int       `json:"year"`
	Code   string    `json:"code"`
	Name   string    `json:"name"`
	Amount int64     `json:"amount"`

	AllocatedGUP   int64 `json:"allocatedGup"`
	AllocatedLS    int64 `json:"allocatedLs"`
	AllocatedTotal int64 `json:"allocatedTotal"`

	UsedGUP   int64 `json:"usedGup"`
	UsedLS    int64 `json:"usedLs"`
	UsedTotal int64 `json:"usedTotal"`

	RemainingGUP   int64 `json:"remainingGup"`
	RemainingLS    int64 `json:"remainingLs"`
	RemainingTotal int64 `json:"remainingTotal"`

	// Available adalah dana yang belum dialokasikan ke GUP/LS.
	Available int64 `json:"available"`

	Notes     string    `json:"notes"`
	IsActive  bool      `json:"isActive"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Allocation struct {
	ID              uuid.UUID `json:"id"`
	FinanceSourceID uuid.UUID `json:"financeSourceId"`
	SourceCode      string    `json:"sourceCode"`
	SourceName      string    `json:"sourceName"`
	AllocationType  string    `json:"allocationType"`
	Amount          int64     `json:"amount"`
	Notes           string    `json:"notes"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type SourceOption struct {
	ID         uuid.UUID `json:"id"`
	Year       int       `json:"year"`
	Code       string    `json:"code"`
	Name       string    `json:"name"`
	BudgetType string    `json:"budgetType"`
	Allocated  int64     `json:"allocated"`
	UsedOther  int64     `json:"usedOther"`
	Available  int64     `json:"available"`
}

type Summary struct {
	TotalFunds int64 `json:"totalFunds"`

	TotalGUP int64 `json:"totalGup"`
	TotalLS  int64 `json:"totalLs"`

	TotalUsedGUP int64 `json:"totalUsedGup"`
	TotalUsedLS  int64 `json:"totalUsedLs"`
	TotalUsed    int64 `json:"totalUsed"`

	TotalAllocated int64 `json:"totalAllocated"`
	TotalAvailable int64 `json:"totalAvailable"`
	TotalRemaining int64 `json:"totalRemaining"`
}

type DashboardResponse struct {
	Year        int          `json:"year"`
	Sources     []Source     `json:"sources"`
	Allocations []Allocation `json:"allocations"`
	Summary     Summary      `json:"summary"`
}

type SaveSourceRequest struct {
	Year     int    `json:"year"`
	Code     string `json:"code"`
	Name     string `json:"name"`
	Amount   int64  `json:"amount"`
	Notes    string `json:"notes"`
	IsActive bool   `json:"isActive"`
}

type SaveAllocationRequest struct {
	FinanceSourceID uuid.UUID `json:"financeSourceId"`
	AllocationType  string    `json:"allocationType"`
	Amount          int64     `json:"amount"`
	Notes           string    `json:"notes"`
}
