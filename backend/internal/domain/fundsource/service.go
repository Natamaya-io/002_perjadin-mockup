package fundsource

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

var ErrValidation = errors.New("validasi sumber dana gagal")

type Service interface {
	GetDashboard(ctx context.Context, year int) (*DashboardResponse, error)
	GetSourceOptions(ctx context.Context, year int, budgetType string, spdNumber string) ([]SourceOption, error)
	CreateSource(ctx context.Context, input SaveSourceRequest) (*Source, error)
	UpdateSource(ctx context.Context, id uuid.UUID, input SaveSourceRequest) (*Source, error)
	DeleteSource(ctx context.Context, id uuid.UUID) error
	CreateAllocation(ctx context.Context, input SaveAllocationRequest) (*Allocation, error)
	DeleteAllocation(ctx context.Context, id uuid.UUID) error
}

type service struct {
	repository Repository
}

func NewService(repository Repository) Service {
	return &service{repository: repository}
}

func validateYear(year int) error {
	if year < 2000 || year > 2100 {
		return fmt.Errorf("%w: tahun tidak valid", ErrValidation)
	}
	return nil
}

func normalizeSource(input SaveSourceRequest) (SaveSourceRequest, error) {
	input.Code = strings.ToUpper(strings.TrimSpace(input.Code))
	input.Name = strings.TrimSpace(input.Name)
	input.Notes = strings.TrimSpace(input.Notes)

	if err := validateYear(input.Year); err != nil {
		return input, err
	}
	if input.Code == "" {
		return input, fmt.Errorf("%w: kode sumber dana wajib diisi", ErrValidation)
	}
	if input.Name == "" {
		return input, fmt.Errorf("%w: nama sumber dana wajib diisi", ErrValidation)
	}
	if input.Amount < 0 {
		return input, fmt.Errorf("%w: nilai sumber dana tidak boleh negatif", ErrValidation)
	}
	return input, nil
}

func (s *service) GetDashboard(ctx context.Context, year int) (*DashboardResponse, error) {
	if err := validateYear(year); err != nil {
		return nil, err
	}

	sources, err := s.repository.ListSources(ctx, year)
	if err != nil {
		return nil, err
	}
	allocations, err := s.repository.ListAllocations(ctx, year)
	if err != nil {
		return nil, err
	}

	summary := Summary{}
	for _, source := range sources {
		summary.TotalFunds += source.Amount
		summary.TotalGUP += source.AllocatedGUP
		summary.TotalLS += source.AllocatedLS
		summary.TotalUsedGUP += source.UsedGUP
		summary.TotalUsedLS += source.UsedLS
		summary.TotalUsed += source.UsedTotal
		summary.TotalAllocated += source.AllocatedTotal
		summary.TotalAvailable += source.Available
		summary.TotalRemaining += source.RemainingTotal
	}

	return &DashboardResponse{
		Year:        year,
		Sources:     sources,
		Allocations: allocations,
		Summary:     summary,
	}, nil
}

func (s *service) GetSourceOptions(ctx context.Context, year int, budgetType string, spdNumber string) ([]SourceOption, error) {
	if err := validateYear(year); err != nil {
		return nil, err
	}

	budgetType = strings.ToUpper(strings.TrimSpace(budgetType))
	spdNumber = strings.TrimSpace(spdNumber)
	if budgetType != "GUP" && budgetType != "LS" {
		return nil, fmt.Errorf("%w: jenis anggaran hanya boleh GUP atau LS", ErrValidation)
	}

	return s.repository.ListSourceOptions(ctx, year, budgetType, spdNumber)
}

func (s *service) CreateSource(ctx context.Context, input SaveSourceRequest) (*Source, error) {
	normalized, err := normalizeSource(input)
	if err != nil {
		return nil, err
	}
	return s.repository.CreateSource(ctx, normalized)
}

func (s *service) UpdateSource(ctx context.Context, id uuid.UUID, input SaveSourceRequest) (*Source, error) {
	normalized, err := normalizeSource(input)
	if err != nil {
		return nil, err
	}

	current, err := s.repository.GetSource(ctx, id)
	if err != nil {
		return nil, err
	}
	if normalized.Amount < current.AllocatedTotal {
		return nil, fmt.Errorf(
			"%w: nominal tidak boleh lebih kecil dari total alokasi %d",
			ErrValidation,
			current.AllocatedTotal,
		)
	}
	if !normalized.IsActive && current.UsedTotal > 0 {
		return nil, fmt.Errorf(
			"%w: sumber dana masih mempunyai realisasi perjalanan dinas",
			ErrValidation,
		)
	}

	return s.repository.UpdateSource(ctx, id, normalized)
}

func (s *service) DeleteSource(ctx context.Context, id uuid.UUID) error {
	return s.repository.DeleteSource(ctx, id)
}

func (s *service) CreateAllocation(ctx context.Context, input SaveAllocationRequest) (*Allocation, error) {
	input.AllocationType = strings.ToUpper(strings.TrimSpace(input.AllocationType))
	input.Notes = strings.TrimSpace(input.Notes)

	if input.FinanceSourceID == uuid.Nil {
		return nil, fmt.Errorf("%w: sumber dana wajib dipilih", ErrValidation)
	}
	if input.AllocationType != "GUP" && input.AllocationType != "LS" {
		return nil, fmt.Errorf("%w: jenis alokasi hanya boleh GUP atau LS", ErrValidation)
	}
	if input.Amount <= 0 {
		return nil, fmt.Errorf("%w: nominal alokasi harus lebih dari 0", ErrValidation)
	}

	return s.repository.CreateAllocation(ctx, input)
}

func (s *service) DeleteAllocation(ctx context.Context, id uuid.UUID) error {
	return s.repository.DeleteAllocation(ctx, id)
}
