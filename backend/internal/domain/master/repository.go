package master

import (
	"context"
	"database/sql"

	"github.com/kemnaker/perjadin-backend/internal/db"
)

type Repository interface {
	GetProvinces(ctx context.Context) ([]db.Province, error)
	GetSBMRates(ctx context.Context) ([]db.SbmRate, error)
	GetSettings(ctx context.Context) ([]db.Setting, error)
	UpdateSetting(ctx context.Context, key, value string) (db.Setting, error)
}

type repository struct {
	q *db.Queries
}

func NewRepository(sqlDB *sql.DB) Repository {
	return &repository{
		q: db.New(sqlDB),
	}
}

func (r *repository) GetProvinces(ctx context.Context) ([]db.Province, error) {
	return r.q.GetProvinces(ctx)
}

func (r *repository) GetSBMRates(ctx context.Context) ([]db.SbmRate, error) {
	return r.q.GetSBMRates(ctx)
}

func (r *repository) GetSettings(ctx context.Context) ([]db.Setting, error) {
	return r.q.GetSettings(ctx)
}

func (r *repository) UpdateSetting(ctx context.Context, key, value string) (db.Setting, error) {
	return r.q.UpdateSetting(ctx, db.UpdateSettingParams{
		Key:   key,
		Value: value,
	})
}
