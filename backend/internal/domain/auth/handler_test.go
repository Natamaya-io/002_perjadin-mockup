package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/kemnaker/perjadin-backend/internal/models"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockService is a mock implementation of auth.Service
type MockService struct {
	mock.Mock
}

func (m *MockService) Register(email, password, name, role string) (*models.User, error) {
	args := m.Called(email, password, name, role)
	if args.Get(0) != nil {
		return args.Get(0).(*models.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockService) Login(email, password string) (string, *models.User, bool, error) {
	args := m.Called(email, password)
	if args.Get(1) != nil {
		return args.String(0), args.Get(1).(*models.User), args.Bool(2), args.Error(3)
	}
	return args.String(0), nil, args.Bool(2), args.Error(3)
}

func (m *MockService) GetDemoUsers() ([]map[string]string, error) {
	args := m.Called()
	if args.Get(0) != nil {
		return args.Get(0).([]map[string]string), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockService) ChangePassword(userIDStr, newPassword string) error {
	args := m.Called(userIDStr, newPassword)
	return args.Error(0)
}

func (m *MockService) GetUserByID(id uuid.UUID) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) != nil {
		return args.Get(0).(*models.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func TestHandler_Login(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockService)
	handler := NewHandler(mockSvc)

	// Valid payload
	reqBody := LoginRequest{Email: "test@example.com", Password: "password123"}
	bodyBytes, _ := json.Marshal(reqBody)
	
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(bodyBytes))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	fakeUser := &models.User{Email: "test@example.com"}
	mockSvc.On("Login", "test@example.com", "password123").Return("mock-token", fakeUser, false, nil)

	err := handler.Login(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestHandler_Login_InvalidPayload(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockService)
	handler := NewHandler(mockSvc)

	// Invalid payload (not JSON)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader([]byte(`invalid json`)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.Login(c)

	assert.Error(t, err)
	httpErr, _ := err.(*echo.HTTPError)
	assert.Equal(t, http.StatusBadRequest, httpErr.Code)
}

func TestHandler_Login_Unauthorized(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockService)
	handler := NewHandler(mockSvc)

	reqBody := LoginRequest{Email: "wrong@example.com", Password: "password"}
	bodyBytes, _ := json.Marshal(reqBody)
	
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(bodyBytes))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockSvc.On("Login", "wrong@example.com", "password").Return("", nil, false, errors.New("invalid credentials"))

	err := handler.Login(c)

	assert.Error(t, err)
	httpErr, _ := err.(*echo.HTTPError)
	assert.Equal(t, http.StatusUnauthorized, httpErr.Code)
}

func TestHandler_Register(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockService)
	handler := NewHandler(mockSvc)

	reqBody := RegisterRequest{Email: "test@example.com", Password: "password123", Name: "Test User", Role: "admin"}
	bodyBytes, _ := json.Marshal(reqBody)
	
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(bodyBytes))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	fakeUser := &models.User{Email: "test@example.com", Role: "staf"}
	// It should downgrade role admin -> staf
	mockSvc.On("Register", "test@example.com", "password123", "Test User", "staf").Return(fakeUser, nil)

	err := handler.Register(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestHandler_Register_InvalidPayload(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockService)
	handler := NewHandler(mockSvc)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader([]byte(`invalid json`)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.Register(c)
	assert.Error(t, err)
	httpErr, _ := err.(*echo.HTTPError)
	assert.Equal(t, http.StatusBadRequest, httpErr.Code)
}

func TestHandler_GetDemoUsers(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockService)
	handler := NewHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/demo-users", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	demoUsers := []map[string]string{{"email": "demo@example.com"}}
	mockSvc.On("GetDemoUsers").Return(demoUsers, nil)

	err := handler.GetDemoUsers(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestHandler_ChangePassword(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockService)
	handler := NewHandler(mockSvc)

	reqBody := ChangePasswordRequest{NewPassword: "newpassword123"}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/change-password", bytes.NewReader(bodyBytes))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	// Simulate middleware setting user_id
	userID := uuid.New().String()
	c.Set("user_id", userID)

	mockSvc.On("ChangePassword", userID, "newpassword123").Return(nil)

	err := handler.ChangePassword(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestHandler_GetMe(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockService)
	handler := NewHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	userID := uuid.New()
	c.Set("user_id", userID.String())

	fakeUser := &models.User{Base: models.Base{ID: userID}, Email: "test@example.com"}
	mockSvc.On("GetUserByID", userID).Return(fakeUser, nil)

	err := handler.GetMe(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}
