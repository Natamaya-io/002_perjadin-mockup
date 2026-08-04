package gupls

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

var ErrValidation = errors.New("validasi LS GUP gagal")

type Service interface {
	Get(ctx context.Context, year int) (*Response, error)
	Save(ctx context.Context, request SaveRequest) (*Response, error)
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

func (s *service) Get(ctx context.Context, year int) (*Response, error) {
	if err := validateYear(year); err != nil {
		return nil, err
	}

	rows, err := s.repository.GetRows(ctx, year)
	if err != nil {
		return nil, err
	}

	summary := Summary{TotalMonths: len(rows)}
	for index := range rows {
		rows[index].Remaining = rows[index].Amount - rows[index].Realization
		summary.TotalAmount += rows[index].Amount
		summary.TotalRealization += rows[index].Realization
		if rows[index].Amount > 0 {
			summary.FilledMonths++
		}
	}
	summary.TotalRemaining = summary.TotalAmount - summary.TotalRealization

	return &Response{
		Year:    year,
		Rows:    rows,
		Summary: summary,
	}, nil
}

func (s *service) Save(
	ctx context.Context,
	request SaveRequest,
) (*Response, error) {
	if err := validateYear(request.Year); err != nil {
		return nil, err
	}
	if request.FundingSourceID == uuid.Nil {
		return nil, fmt.Errorf("%w: bulan LS wajib dipilih", ErrValidation)
	}
	if request.Amount < 0 {
		return nil, fmt.Errorf("%w: nilai LS tidak boleh negatif", ErrValidation)
	}

	if err := s.repository.SaveAmount(
		ctx,
		request.Year,
		request.FundingSourceID,
		request.Amount,
	); err != nil {
		return nil, err
	}

	return s.Get(ctx, request.Year)
}
