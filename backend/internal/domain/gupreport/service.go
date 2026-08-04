package gupreport

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

var ErrValidation = errors.New("validasi laporan GUP gagal")

type Service interface {
	GetReport(ctx context.Context, year int) (*ReportResponse, error)
	SaveBudget(ctx context.Context, request SaveBudgetRequest) (*ReportResponse, error)
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

func (s *service) GetReport(ctx context.Context, year int) (*ReportResponse, error) {
	if err := validateYear(year); err != nil {
		return nil, err
	}

	rows, err := s.repository.GetRows(ctx, year)
	if err != nil {
		return nil, err
	}

	summary := Summary{TotalProcurementType: len(rows)}
	for index := range rows {
		rows[index].Remaining = rows[index].Budget - rows[index].Realization
		if rows[index].Budget > 0 {
			rows[index].Absorption = float64(rows[index].Realization) / float64(rows[index].Budget) * 100
		}

		summary.TotalBudget += rows[index].Budget
		summary.TotalRealization += rows[index].Realization
		summary.TotalTransactions += rows[index].TransactionCount
		summary.TotalSubmissionValue += rows[index].SubmissionValue
	}

	summary.TotalRemaining = summary.TotalBudget - summary.TotalRealization
	if summary.TotalBudget > 0 {
		summary.TotalAbsorption = float64(summary.TotalRealization) / float64(summary.TotalBudget) * 100
	}

	return &ReportResponse{
		Year:    year,
		Rows:    rows,
		Summary: summary,
	}, nil
}

func (s *service) SaveBudget(
	ctx context.Context,
	request SaveBudgetRequest,
) (*ReportResponse, error) {
	if err := validateYear(request.Year); err != nil {
		return nil, err
	}
	if request.ProcurementTypeID == uuid.Nil {
		return nil, fmt.Errorf("%w: jenis pengadaan wajib dipilih", ErrValidation)
	}
	if request.Amount < 0 {
		return nil, fmt.Errorf("%w: anggaran tidak boleh negatif", ErrValidation)
	}

	if err := s.repository.SaveBudget(
		ctx,
		request.Year,
		request.ProcurementTypeID,
		request.Amount,
	); err != nil {
		return nil, err
	}

	return s.GetReport(ctx, request.Year)
}
