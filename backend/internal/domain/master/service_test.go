package master

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/kemnaker/perjadin-backend/internal/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRepository is a mock implementation of master.Repository
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) GetProvinces(ctx context.Context) ([]db.Province, error) {
	args := m.Called(ctx)
	if args.Get(0) != nil {
		return args.Get(0).([]db.Province), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockRepository) GetSBMRates(ctx context.Context) ([]db.SbmRate, error) {
	args := m.Called(ctx)
	if args.Get(0) != nil {
		return args.Get(0).([]db.SbmRate), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockRepository) GetSettings(ctx context.Context) ([]db.Setting, error) {
	args := m.Called(ctx)
	if args.Get(0) != nil {
		return args.Get(0).([]db.Setting), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockRepository) UpdateSetting(ctx context.Context, key, value string) (db.Setting, error) {
	args := m.Called(ctx, key, value)
	return db.Setting{}, args.Error(1)
}

func TestService_GetProvinces(t *testing.T) {
	mockRepo := new(MockRepository)
	svc := NewService(mockRepo, nil)
	ctx := context.Background()

	expectedProvinces := []db.Province{
		{ID: uuid.New(), Name: "DKI Jakarta"},
		{ID: uuid.New(), Name: "Jawa Barat"},
	}

	mockRepo.On("GetProvinces", ctx).Return(expectedProvinces, nil)

	res, err := svc.GetProvinces(ctx)

	assert.NoError(t, err)
	assert.Len(t, res, 2)
	assert.Equal(t, "DKI Jakarta", res[0].Name)
	mockRepo.AssertExpectations(t)
}

func TestService_GetSBMRates(t *testing.T) {
	mockRepo := new(MockRepository)
	svc := NewService(mockRepo, nil)
	ctx := context.Background()

	expectedRates := []db.SbmRate{
		{ProvinceID: uuid.New(), OutsideCityRate: sql.NullFloat64{Float64: 530000, Valid: true}},
	}

	mockRepo.On("GetSBMRates", ctx).Return(expectedRates, nil)

	res, err := svc.GetSBMRates(ctx)

	assert.NoError(t, err)
	assert.Len(t, res, 1)
	assert.Equal(t, float64(530000), res[0].OutsideCityRate.Float64)
	mockRepo.AssertExpectations(t)
}

func TestService_GetSettings(t *testing.T) {
	mockRepo := new(MockRepository)
	svc := NewService(mockRepo, nil)
	ctx := context.Background()

	dbSettings := []db.Setting{
		{Key: "nama_ppk", Value: "Mr. X"},
		{Key: "nip_ppk", Value: "12345"},
	}

	mockRepo.On("GetSettings", ctx).Return(dbSettings, nil)

	res, err := svc.GetSettings(ctx)

	assert.NoError(t, err)
	assert.Len(t, res, 2)
	assert.Equal(t, "Mr. X", res["nama_ppk"])
	assert.Equal(t, "12345", res["nip_ppk"])
	mockRepo.AssertExpectations(t)
}

func TestService_UpdateSettings(t *testing.T) {
	mockRepo := new(MockRepository)
	svc := NewService(mockRepo, nil)
	ctx := context.Background()

	newSettings := map[string]string{
		"nama_ppk": "New PPK",
		"nip_ppk":  "67890",
	}

	mockRepo.On("UpdateSetting", ctx, "nama_ppk", "New PPK").Return(db.Setting{}, nil)
	mockRepo.On("UpdateSetting", ctx, "nip_ppk", "67890").Return(db.Setting{}, nil)

	err := svc.UpdateSettings(ctx, newSettings)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestService_UpdateSettings_Error(t *testing.T) {
	mockRepo := new(MockRepository)
	svc := NewService(mockRepo, nil)
	ctx := context.Background()

	newSettings := map[string]string{
		"nama_ppk": "New PPK",
	}

	mockRepo.On("UpdateSetting", ctx, "nama_ppk", "New PPK").Return(db.Setting{}, errors.New("db error"))

	err := svc.UpdateSettings(ctx, newSettings)

	assert.Error(t, err)
	assert.Equal(t, "db error", err.Error())
	mockRepo.AssertExpectations(t)
}
