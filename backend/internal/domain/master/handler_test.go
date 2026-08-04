package master

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kemnaker/perjadin-backend/internal/db"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockService is a mock implementation of master.Service
type MockService struct {
	mock.Mock
}

func (m *MockService) GetProvinces(ctx context.Context) ([]db.Province, error) {
	args := m.Called(ctx)
	if args.Get(0) != nil {
		return args.Get(0).([]db.Province), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockService) GetSBMRates(ctx context.Context) ([]db.SbmRate, error) {
	args := m.Called(ctx)
	if args.Get(0) != nil {
		return args.Get(0).([]db.SbmRate), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockService) GetSettings(ctx context.Context) (map[string]string, error) {
	args := m.Called(ctx)
	if args.Get(0) != nil {
		return args.Get(0).(map[string]string), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockService) UpdateSettings(ctx context.Context, settings map[string]string) error {
	args := m.Called(ctx, settings)
	return args.Error(0)
}

func TestHandler_GetProvinces(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockService)
	handler := NewHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/master/provinces", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	provinces := []db.Province{{Name: "Jakarta"}}
	mockSvc.On("GetProvinces", mock.Anything).Return(provinces, nil)

	err := handler.GetProvinces(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	
	var res []db.Province
	json.Unmarshal(rec.Body.Bytes(), &res)
	assert.Len(t, res, 1)
	assert.Equal(t, "Jakarta", res[0].Name)
}

func TestHandler_GetSBMRates(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockService)
	handler := NewHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/master/sbm-rates", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	rates := []db.SbmRate{{OutsideCityRate: sql.NullFloat64{Float64: 500000, Valid: true}}}
	mockSvc.On("GetSBMRates", mock.Anything).Return(rates, nil)

	err := handler.GetSBMRates(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestHandler_GetSettings(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockService)
	handler := NewHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/master/settings", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	settings := map[string]string{"key": "value"}
	mockSvc.On("GetSettings", mock.Anything).Return(settings, nil)

	err := handler.GetSettings(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestHandler_UpdateSettings(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockService)
	handler := NewHandler(mockSvc)

	settings := map[string]string{"key": "value"}
	bodyBytes, _ := json.Marshal(settings)

	req := httptest.NewRequest(http.MethodPost, "/master/settings", bytes.NewReader(bodyBytes))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockSvc.On("UpdateSettings", mock.Anything, settings).Return(nil)

	err := handler.UpdateSettings(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestHandler_UpdateSettings_InvalidPayload(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockService)
	handler := NewHandler(mockSvc)

	req := httptest.NewRequest(http.MethodPost, "/master/settings", bytes.NewReader([]byte(`invalid`)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.UpdateSettings(c)

	assert.Error(t, err)
	httpErr, _ := err.(*echo.HTTPError)
	assert.Equal(t, http.StatusBadRequest, httpErr.Code)
}