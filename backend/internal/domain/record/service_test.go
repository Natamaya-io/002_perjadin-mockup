package record

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kemnaker/perjadin-backend/internal/config"
	"github.com/kemnaker/perjadin-backend/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRepository is a mock implementation of record.Repository
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) NextSpdNumber(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRepository) CreateTravelRecord(ctx context.Context, record *models.TravelRecord) error {
	args := m.Called(ctx, record)
	return args.Error(0)
}

func (m *MockRepository) CreateTravelRecordsBulk(ctx context.Context, records []*models.TravelRecord) error {
	args := m.Called(ctx, records)
	return args.Error(0)
}

func (m *MockRepository) GetOverlappingRecords(ctx context.Context, employeeID uuid.UUID, startDate, endDate time.Time) ([]models.TravelRecord, error) {
	args := m.Called(ctx, employeeID, startDate, endDate)
	if args.Get(0) != nil {
		return args.Get(0).([]models.TravelRecord), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockRepository) GetTravelRecords(ctx context.Context, filters map[string]interface{}) ([]models.TravelRecord, error) {
	args := m.Called(ctx, filters)
	if args.Get(0) != nil {
		return args.Get(0).([]models.TravelRecord), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockRepository) GetTravelRecordByID(ctx context.Context, id uuid.UUID) (*models.TravelRecord, error) {
	args := m.Called(ctx, id)
	if args.Get(0) != nil {
		return args.Get(0).(*models.TravelRecord), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) != nil {
		return args.Get(0).(*models.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockRepository) UpdateTravelRecord(ctx context.Context, record *models.TravelRecord) error {
	args := m.Called(ctx, record)
	return args.Error(0)
}

func (m *MockRepository) SyncReportBySpd(ctx context.Context, spd string, sourceRecord *models.TravelRecord) error {
	args := m.Called(ctx, spd, sourceRecord)
	return args.Error(0)
}

func (m *MockRepository) DeleteTravelRecord(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) DeleteTravelRecordsBySpd(ctx context.Context, spd string) error {
	args := m.Called(ctx, spd)
	return args.Error(0)
}

func (m *MockRepository) SyncSpdSequence(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockRepository) GetDashboardSummary(ctx context.Context, role, userIDStr string) (*models.DashboardSummary, error) {
	args := m.Called(ctx, role, userIDStr)
	if args.Get(0) != nil {
		return args.Get(0).(*models.DashboardSummary), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockRepository) GetPaginatedRecords(ctx context.Context, params models.PaginatedParams) (*models.PaginatedResponse, error) {
	args := m.Called(ctx, params)
	if args.Get(0) != nil {
		return args.Get(0).(*models.PaginatedResponse), args.Error(1)
	}
	return nil, args.Error(1)
}

func TestService_GenerateSpdNumber(t *testing.T) {
	mockRepo := new(MockRepository)
	svc := NewService(mockRepo, &config.Config{}, nil)
	ctx := context.Background()

	mockRepo.On("NextSpdNumber", mock.Anything).Return(int64(42), nil)

	spd, err := svc.GenerateSpdNumber(ctx)

	assert.NoError(t, err)
	assert.Equal(t, "ID-SPJ-042", spd)
	mockRepo.AssertExpectations(t)
}

func TestService_GenerateSpdNumber_Error(t *testing.T) {
	mockRepo := new(MockRepository)
	svc := NewService(mockRepo, &config.Config{}, nil)
	ctx := context.Background()

	mockRepo.On("NextSpdNumber", mock.Anything).Return(int64(0), errors.New("db error"))

	spd, err := svc.GenerateSpdNumber(ctx)

	assert.Error(t, err)
	assert.Empty(t, spd)
	assert.Contains(t, err.Error(), "GenerateSpdNumber: nextval failed")
	mockRepo.AssertExpectations(t)
}

func TestService_CreateRecord_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	svc := NewService(mockRepo, &config.Config{}, nil)
	ctx := context.Background()
	employeeID := uuid.New()

	record := &models.TravelRecord{
		EmployeeID: employeeID,
		Locations: []models.TravelLocation{
			{StartDate: time.Now(), EndDate: time.Now().AddDate(0, 0, 1)},
		},
	}

	mockRepo.On("GetOverlappingRecords", ctx, employeeID, mock.Anything, mock.Anything).Return([]models.TravelRecord{}, nil)
	mockRepo.On("NextSpdNumber", mock.Anything).Return(int64(1), nil)
	mockRepo.On("CreateTravelRecord", ctx, record).Return(nil)
	mockRepo.On("GetUserByID", mock.Anything, mock.Anything).Return(&models.User{NomorHP: "08123456789"}, nil).Maybe()

	err := svc.CreateRecord(ctx, record)

	assert.NoError(t, err)
	assert.Equal(t, "ID-SPJ-001", record.SPDNumber)
}

func TestService_CreateRecord_NoLocation(t *testing.T) {
	mockRepo := new(MockRepository)
	svc := NewService(mockRepo, &config.Config{}, nil)
	ctx := context.Background()

	record := &models.TravelRecord{
		EmployeeID: uuid.New(),
		Locations:  []models.TravelLocation{},
	}

	err := svc.CreateRecord(ctx, record)

	assert.Error(t, err)
	assert.Equal(t, "at least one location is required", err.Error())
	mockRepo.AssertNotCalled(t, "CreateTravelRecord")
}

func TestService_CreateRecord_InvalidDates(t *testing.T) {
	mockRepo := new(MockRepository)
	svc := NewService(mockRepo, &config.Config{}, nil)
	ctx := context.Background()

	record := &models.TravelRecord{
		EmployeeID: uuid.New(),
		Locations: []models.TravelLocation{
			{Location: "Jakarta", StartDate: time.Now(), EndDate: time.Now().AddDate(0, 0, -1)}, // End before start
		},
	}

	err := svc.CreateRecord(ctx, record)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "end date cannot be before start date")
	mockRepo.AssertNotCalled(t, "CreateTravelRecord")
}

func TestService_CreateRecord_Overlapping(t *testing.T) {
	mockRepo := new(MockRepository)
	svc := NewService(mockRepo, &config.Config{}, nil)
	ctx := context.Background()
	employeeID := uuid.New()

	record := &models.TravelRecord{
		EmployeeID: employeeID,
		Locations: []models.TravelLocation{
			{StartDate: time.Now(), EndDate: time.Now().AddDate(0, 0, 1)},
		},
	}

	overlapping := []models.TravelRecord{{Base: models.Base{ID: uuid.New()}}}
	mockRepo.On("GetOverlappingRecords", ctx, employeeID, mock.Anything, mock.Anything).Return(overlapping, nil)

	err := svc.CreateRecord(ctx, record)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "employee is already assigned")
	mockRepo.AssertNotCalled(t, "CreateTravelRecord")
}

func TestService_DeleteRecord(t *testing.T) {
	mockRepo := new(MockRepository)
	svc := NewService(mockRepo, &config.Config{}, nil)
	ctx := context.Background()

	recordID := uuid.New()
	mockRepo.On("DeleteTravelRecord", ctx, recordID).Return(nil)
	mockRepo.On("SyncSpdSequence", mock.Anything).Return(nil).Maybe()

	err := svc.DeleteRecord(ctx, recordID)

	assert.NoError(t, err)
}

func TestService_GetRecords(t *testing.T) {
	mockRepo := new(MockRepository)
	svc := NewService(mockRepo, &config.Config{}, nil)
	ctx := context.Background()

	mockRecords := []models.TravelRecord{{Base: models.Base{ID: uuid.New()}}}
	filters := map[string]interface{}{"status": "Draft"}

	mockRepo.On("GetTravelRecords", ctx, filters).Return(mockRecords, nil)

	records, err := svc.GetRecords(ctx, filters)

	assert.NoError(t, err)
	assert.Len(t, records, 1)
	mockRepo.AssertExpectations(t)
}

func TestService_GetRecordByID(t *testing.T) {
	mockRepo := new(MockRepository)
	svc := NewService(mockRepo, &config.Config{}, nil)
	ctx := context.Background()

	recordID := uuid.New()
	mockRecord := &models.TravelRecord{Base: models.Base{ID: recordID}}

	mockRepo.On("GetTravelRecordByID", ctx, recordID).Return(mockRecord, nil)

	record, err := svc.GetRecordByID(ctx, recordID)

	assert.NoError(t, err)
	assert.NotNil(t, record)
	assert.Equal(t, recordID, record.ID)
	mockRepo.AssertExpectations(t)
}

func TestService_UpdateRecord(t *testing.T) {
	mockRepo := new(MockRepository)
	svc := NewService(mockRepo, &config.Config{}, nil)
	ctx := context.Background()

	recordID := uuid.New()
	record := &models.TravelRecord{
		Base: models.Base{ID: recordID},
		Locations: []models.TravelLocation{
			{StartDate: time.Now(), EndDate: time.Now().AddDate(0, 0, 1)},
		},
		SPDNumber: "ID-SPJ-001",
		Report: &models.TravelReport{Text: "Test"},
	}

	mockRepo.On("UpdateTravelRecord", ctx, record).Return(nil)
	mockRepo.On("SyncReportBySpd", mock.Anything, "ID-SPJ-001", record).Return(nil).Maybe()
	mockRepo.On("GetTravelRecords", ctx, map[string]interface{}{"spd": "ID-SPJ-001"}).Return([]models.TravelRecord{}, nil)
	mockRepo.On("SyncSpdSequence", mock.Anything).Return(nil).Maybe()

	err := svc.UpdateRecord(ctx, record)

	assert.NoError(t, err)
	// We wait a bit for goroutines to execute, although Maybe() handles if it's not called yet
	time.Sleep(10 * time.Millisecond)
	mockRepo.AssertExpectations(t)
}

func TestService_InvalidateRecordCaches(t *testing.T) {
	mockRepo := new(MockRepository)
	svc := NewService(mockRepo, &config.Config{}, nil)
	ctx := context.Background()

	// Just testing it doesn't panic when redisClient is nil
	assert.NotPanics(t, func() {
		svc.InvalidateAllCache(ctx)
	})
}

func TestService_GetPaginatedRecords(t *testing.T) {
	mockRepo := new(MockRepository)
	svc := NewService(mockRepo, &config.Config{}, nil)
	ctx := context.Background()

	params := models.PaginatedParams{Limit: 10}
	mockResponse := &models.PaginatedResponse{
		TotalItems: 1,
		Data:       []models.TravelRecord{{Base: models.Base{ID: uuid.New()}}},
	}

	mockRepo.On("GetPaginatedRecords", ctx, params).Return(mockResponse, nil)

	res, err := svc.GetPaginatedRecords(ctx, params)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, int64(1), res.TotalItems)
	mockRepo.AssertExpectations(t)
}

func TestService_GetDashboardSummary(t *testing.T) {
	mockRepo := new(MockRepository)
	svc := NewService(mockRepo, &config.Config{}, nil)
	ctx := context.Background()

	mockSummary := &models.DashboardSummary{TotalTrips: 5}

	mockRepo.On("GetDashboardSummary", ctx, "admin", "user123").Return(mockSummary, nil)

	res, err := svc.GetDashboardSummary(ctx, "admin", "user123")

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, int64(5), res.TotalTrips)
	mockRepo.AssertExpectations(t)
}

func TestService_DeleteRecordsBySpd(t *testing.T) {
	mockRepo := new(MockRepository)
	svc := NewService(mockRepo, &config.Config{}, nil)
	ctx := context.Background()

	mockRepo.On("DeleteTravelRecordsBySpd", ctx, "ID-SPJ-001").Return(nil)
	mockRepo.On("SyncSpdSequence", mock.Anything).Return(nil).Maybe()

	err := svc.DeleteRecordsBySpd(ctx, "ID-SPJ-001")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}


