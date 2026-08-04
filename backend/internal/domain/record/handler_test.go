package record

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kemnaker/perjadin-backend/internal/config"
	"github.com/kemnaker/perjadin-backend/internal/models"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockService is a mock implementation of record.Service
type MockService struct {
	mock.Mock
}

func (m *MockService) GenerateSpdNumber(ctx context.Context) (string, error) { return "", nil }
func (m *MockService) CreateRecord(ctx context.Context, record *models.TravelRecord) error {
	args := m.Called(ctx, record)
	return args.Error(0)
}
func (m *MockService) CreateRecordsBulk(ctx context.Context, records []*models.TravelRecord) error {
	return nil
}
func (m *MockService) CreateRecordDirect(ctx context.Context, record *models.TravelRecord) error {
	return nil
}
func (m *MockService) InvalidateAllCache(ctx context.Context) {}

func (m *MockService) GetRecords(ctx context.Context, filters map[string]interface{}) ([]models.TravelRecord, error) {
	args := m.Called(ctx, filters)
	if args.Get(0) != nil {
		return args.Get(0).([]models.TravelRecord), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockService) GetRecordByID(ctx context.Context, id uuid.UUID) (*models.TravelRecord, error) {
	args := m.Called(ctx, id)
	if args.Get(0) != nil {
		return args.Get(0).(*models.TravelRecord), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockService) UpdateRecord(ctx context.Context, record *models.TravelRecord) error {
	args := m.Called(ctx, record)
	return args.Error(0)
}

func (m *MockService) DeleteRecord(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockService) DeleteRecordsBySpd(ctx context.Context, spd string) error {
	args := m.Called(ctx, spd)
	return args.Error(0)
}

func (m *MockService) GetDashboardSummary(ctx context.Context, role string, userIDStr string) (*models.DashboardSummary, error) {
	args := m.Called(ctx, role, userIDStr)
	if args.Get(0) != nil {
		return args.Get(0).(*models.DashboardSummary), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockService) GetPaginatedRecords(ctx context.Context, params models.PaginatedParams) (*models.PaginatedResponse, error) {
	return nil, nil
}

func TestHandler_GetRecords(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockService)
	handler := NewHandler(mockSvc, nil, &config.Config{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/records?status=Draft", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("role", "admin")
	c.Set("user_id", "admin-1")

	expectedRecords := []models.TravelRecord{{SPDNumber: "ID-SPJ-001"}}
	mockSvc.On("GetRecords", mock.Anything, mock.Anything).Return(expectedRecords, nil)

	err := handler.GetRecords(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestHandler_CreateRecord(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockService)
	handler := NewHandler(mockSvc, nil, &config.Config{}, nil)

	record := models.TravelRecord{
		EmployeeID: uuid.New(),
		Locations: []models.TravelLocation{
			{StartDate: time.Now(), EndDate: time.Now().AddDate(0, 0, 1)},
		},
	}
	bodyBytes, _ := json.Marshal(record)

	req := httptest.NewRequest(http.MethodPost, "/records", bytes.NewReader(bodyBytes))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockSvc.On("CreateRecord", mock.Anything, mock.AnythingOfType("*models.TravelRecord")).Return(nil)

	err := handler.CreateRecord(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestHandler_DeleteRecord(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockService)
	handler := NewHandler(mockSvc, nil, &config.Config{}, nil)

	recordID := uuid.New()
	req := httptest.NewRequest(http.MethodDelete, "/records/"+recordID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(recordID.String())
	c.Set("role", "super_admin")
	c.Set("user_id", uuid.New().String())

	mockSvc.On("GetRecordByID", mock.Anything, recordID).Return(&models.TravelRecord{}, nil)
	mockSvc.On("DeleteRecord", mock.Anything, recordID).Return(nil)

	err := handler.DeleteRecord(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestHandler_DeleteRecordsBySpd(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockService)
	handler := NewHandler(mockSvc, nil, &config.Config{}, nil)

	req := httptest.NewRequest(http.MethodDelete, "/records/spd/ID-SPJ-001", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("spd")
	c.SetParamValues("ID-SPJ-001")
	c.Set("role", "super_admin")

	mockSvc.On("DeleteRecordsBySpd", mock.Anything, "ID-SPJ-001").Return(nil)

	err := handler.DeleteRecordsBySpd(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestHandler_GetDashboardSummary(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockService)
	handler := NewHandler(mockSvc, nil, &config.Config{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/dashboard/summary", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("role", "admin")
	c.Set("user_id", "user-1")

	expectedSummary := &models.DashboardSummary{TotalTrips: 5}
	mockSvc.On("GetDashboardSummary", mock.Anything, "admin", "user-1").Return(expectedSummary, nil)

	err := handler.GetDashboardSummary(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestHandler_UpdateRecord(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockService)
	handler := NewHandler(mockSvc, nil, &config.Config{}, nil)

	recordID := uuid.New()
	body := `{"spd": "ID-SPJ-001"}`
	req := httptest.NewRequest(http.MethodPut, "/records/"+recordID.String(), strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(recordID.String())
	c.Set("role", "super_admin")
	c.Set("user_id", "user-1")

	mockRecord := &models.TravelRecord{Base: models.Base{ID: recordID}, SPDNumber: "ID-SPJ-001"}
	mockSvc.On("GetRecordByID", mock.Anything, recordID).Return(mockRecord, nil)
	mockSvc.On("UpdateRecord", mock.Anything, mock.AnythingOfType("*models.TravelRecord")).Return(nil)

	err := handler.UpdateRecord(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestHandler_GetRecordByID(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockService)
	handler := NewHandler(mockSvc, nil, &config.Config{}, nil)

	recordID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/records/"+recordID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(recordID.String())
	c.Set("role", "super_admin")
	c.Set("user_id", "user-1")

	mockRecord := &models.TravelRecord{Base: models.Base{ID: recordID}, SPDNumber: "ID-SPJ-001"}
	mockSvc.On("GetRecordByID", mock.Anything, recordID).Return(mockRecord, nil)

	err := handler.GetRecordByID(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}


