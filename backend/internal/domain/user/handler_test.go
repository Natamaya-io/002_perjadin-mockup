package user

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/kemnaker/perjadin-backend/internal/models"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockService is a mock implementation of user.Service
type MockService struct {
	mock.Mock
}

func (m *MockService) GetUsers() ([]models.User, error) {
	args := m.Called()
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *MockService) GetUserByID(id uuid.UUID) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) != nil {
		return args.Get(0).(*models.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockService) CreateUser(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockService) UpdateUser(user *models.User, newPassword string) error {
	args := m.Called(user, newPassword)
	return args.Error(0)
}

func (m *MockService) DeleteUser(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func TestHandler_GetUsers(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockService)
	handler := NewHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("role", "super_admin")
	c.Set("user_id", uuid.New().String())

	users := []models.User{
		{Email: "user1@example.com", Role: "staf"},
		{Email: "user2@example.com", Role: "kasubag"},
	}

	mockSvc.On("GetUsers").Return(users, nil)

	err := handler.GetUsers(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	
	var res []models.User
	json.Unmarshal(rec.Body.Bytes(), &res)
	assert.Len(t, res, 2)
}

func TestHandler_CreateUser(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockService)
	handler := NewHandler(mockSvc)

	reqBody := CreateUserRequest{
		Email:   "test@example.com",
		Name:    "Test User",
		NomorHP: "081234567890",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(bodyBytes))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockSvc.On("CreateUser", mock.AnythingOfType("*models.User")).Return(nil)

	err := handler.CreateUser(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestHandler_DeleteUser(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockService)
	handler := NewHandler(mockSvc)

	userID := uuid.New()
	req := httptest.NewRequest(http.MethodDelete, "/users/"+userID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(userID.String())

	mockSvc.On("DeleteUser", userID).Return(nil)

	err := handler.DeleteUser(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestHandler_UpdateUser(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockService)
	handler := NewHandler(mockSvc)

	userID := uuid.New()
	reqBody := UpdateUserRequest{
		Email:   "updated@example.com",
		Name:    "Updated User",
		NomorHP: "0899999999",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/users/"+userID.String(), bytes.NewReader(bodyBytes))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(userID.String())

	existingUser := &models.User{Base: models.Base{ID: userID}}
	mockSvc.On("GetUserByID", userID).Return(existingUser, nil)
	mockSvc.On("UpdateUser", mock.AnythingOfType("*models.User"), "").Return(nil)

	err := handler.UpdateUser(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestHandler_UpdateUser_InvalidUUID(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockService)
	handler := NewHandler(mockSvc)

	req := httptest.NewRequest(http.MethodPut, "/users/invalid-uuid", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("invalid-uuid")

	err := handler.UpdateUser(c)

	assert.Error(t, err)
	httpErr, _ := err.(*echo.HTTPError)
	assert.Equal(t, http.StatusBadRequest, httpErr.Code)
}

func TestHandler_GetUsers_WithFilters(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockService)
	handler := NewHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/users?search=agus&role=staf", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("role", "super_admin")
	c.Set("user_id", uuid.New().String())

	users := []models.User{
		{Base: models.Base{ID: uuid.New()}, Email: "agus@example.com", Name: "Agus", Role: "staf"},
		{Base: models.Base{ID: uuid.New()}, Email: "budi@example.com", Name: "Budi", Role: "staf"},
		{Base: models.Base{ID: uuid.New()}, Email: "agus2@example.com", Name: "Agus Kasubag", Role: "kasubag"},
	}

	mockSvc.On("GetUsers").Return(users, nil)

	err := handler.GetUsers(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	
	var res []models.User
	json.Unmarshal(rec.Body.Bytes(), &res)
	// Should only match "Agus" and role "staf"
	assert.Len(t, res, 1)
	assert.Equal(t, "agus@example.com", res[0].Email)
}
