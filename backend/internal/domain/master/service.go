package master

import (
	"context"
	"encoding/json"
	"time"

	"github.com/kemnaker/perjadin-backend/internal/db"
	"github.com/redis/go-redis/v9"
)

type Service interface {
	GetProvinces(ctx context.Context) ([]db.Province, error)
	GetSBMRates(ctx context.Context) ([]db.SbmRate, error)
	GetSettings(ctx context.Context) (map[string]string, error)
	UpdateSettings(ctx context.Context, settings map[string]string) error
}

type service struct {
	repo  Repository
	rdb   *redis.Client
}

func NewService(repo Repository, rdb *redis.Client) Service {
	return &service{repo: repo, rdb: rdb}
}

func (s *service) GetProvinces(ctx context.Context) ([]db.Province, error) {
	cacheKey := "master:provinces:v2"
	if s.rdb != nil {
		if val, err := s.rdb.Get(ctx, cacheKey).Result(); err == nil {
			var provinces []db.Province
			if err := json.Unmarshal([]byte(val), &provinces); err == nil {
				return provinces, nil
			}
		}
	}

	provinces, err := s.repo.GetProvinces(ctx)
	if err != nil {
		return nil, err
	}

	if s.rdb != nil {
		if data, err := json.Marshal(provinces); err == nil {
			s.rdb.Set(ctx, cacheKey, data, 24*time.Hour)
		}
	}

	return provinces, nil
}

func (s *service) GetSBMRates(ctx context.Context) ([]db.SbmRate, error) {
	cacheKey := "master:sbm_rates:v2"
	if s.rdb != nil {
		if val, err := s.rdb.Get(ctx, cacheKey).Result(); err == nil {
			var rates []db.SbmRate
			if err := json.Unmarshal([]byte(val), &rates); err == nil {
				return rates, nil
			}
		}
	}

	rates, err := s.repo.GetSBMRates(ctx)
	if err != nil {
		return nil, err
	}

	if s.rdb != nil {
		if data, err := json.Marshal(rates); err == nil {
			s.rdb.Set(ctx, cacheKey, data, 24*time.Hour)
		}
	}

	return rates, nil
}

func (s *service) GetSettings(ctx context.Context) (map[string]string, error) {
	cacheKey := "master:settings"
	if s.rdb != nil {
		if val, err := s.rdb.Get(ctx, cacheKey).Result(); err == nil {
			var settings map[string]string
			if err := json.Unmarshal([]byte(val), &settings); err == nil {
				return settings, nil
			}
		}
	}

	dbs, err := s.repo.GetSettings(ctx)
	if err != nil {
		return nil, err
	}

	settings := make(map[string]string)
	for _, setting := range dbs {
		settings[setting.Key] = setting.Value
	}

	if s.rdb != nil {
		if data, err := json.Marshal(settings); err == nil {
			s.rdb.Set(ctx, cacheKey, data, 24*time.Hour)
		}
	}

	return settings, nil
}

func (s *service) UpdateSettings(ctx context.Context, settings map[string]string) error {
	for k, v := range settings {
		if _, err := s.repo.UpdateSetting(ctx, k, v); err != nil {
			return err
		}
	}

	if s.rdb != nil {
		s.rdb.Del(ctx, "master:settings")
	}

	return nil
}
