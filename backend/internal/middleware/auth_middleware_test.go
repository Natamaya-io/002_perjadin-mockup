package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/kemnaker/perjadin-backend/internal/config"
	"github.com/kemnaker/perjadin-backend/internal/models"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository is a mock implementation of user.Repository for the middleware
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(u *models.User) error { return nil }
func (m *MockUserRepository) GetUserByEmail(email string) (*models.User, error) { return nil, nil }
func (m *MockUserRepository) UpdateUser(u *models.User) error { return nil }
func (m *MockUserRepository) GetUsers() ([]models.User, error) { return nil, nil }
func (m *MockUserRepository) DeleteUser(id uuid.UUID) error { return nil }

func (m *MockUserRepository) GetUserByID(id uuid.UUID) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) != nil {
		return args.Get(0).(*models.User), args.Error(1)
	}
	return nil, args.Error(1)
}

// Generate valid test token
func generateTestToken(secret, userID, sessionID string, role string, expired bool) string {
	claims := jwt.MapClaims{
		"user_id":    userID,
		"session_id": sessionID,
		"role":       role,
		"email":      "test@example.com",
	}

	if expired {
		claims["exp"] = time.Now().Add(-1 * time.Hour).Unix()
	} else {
		claims["exp"] = time.Now().Add(1 * time.Hour).Unix()
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secret))
	return tokenString
}

func TestJWTMiddleware_MissingHeader(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockRepo := new(MockUserRepository)
	cfg := &config.Config{JWT: config.JWTConfig{Secret: "secret"}}

	h := JWTMiddleware(cfg, mockRepo)(func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	})

	err := h(c)
	assert.Error(t, err)
	httpErr, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, httpErr.Code)
	assert.Equal(t, "Missing Authorization Header", httpErr.Message)
}

func TestJWTMiddleware_ValidToken(t *testing.T) {
	e := echo.New()
	userID := uuid.New()
	sessionID := uuid.New().String()
	secret := "mysecret"

	token := generateTestToken(secret, userID.String(), sessionID, "admin", false)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockRepo := new(MockUserRepository)
	cfg := &config.Config{JWT: config.JWTConfig{Secret: secret}}

	mockUser := &models.User{
		Base:      models.Base{ID: userID},
		SessionID: sessionID,
	}
	mockRepo.On("GetUserByID", userID).Return(mockUser, nil)

	h := JWTMiddleware(cfg, mockRepo)(func(c echo.Context) error {
		// Verify context is set
		assert.Equal(t, userID.String(), c.Get("user_id"))
		assert.Equal(t, "admin", c.Get("role"))
		return c.String(http.StatusOK, "success")
	})

	err := h(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockRepo.AssertExpectations(t)
}

func TestJWTMiddleware_InvalidSession(t *testing.T) {
	e := echo.New()
	userID := uuid.New()
	sessionID := uuid.New().String()
	secret := "mysecret"

	token := generateTestToken(secret, userID.String(), sessionID, "admin", false)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockRepo := new(MockUserRepository)
	cfg := &config.Config{JWT: config.JWTConfig{Secret: secret}}

	mockUser := &models.User{
		Base:      models.Base{ID: userID},
		SessionID: "different_session_id_from_db", // Indicates login from elsewhere
	}
	mockRepo.On("GetUserByID", userID).Return(mockUser, nil)

	h := JWTMiddleware(cfg, mockRepo)(func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	})

	err := h(c)
	assert.Error(t, err)
	httpErr, _ := err.(*echo.HTTPError)
	assert.Equal(t, http.StatusUnauthorized, httpErr.Code)
	assert.Contains(t, httpErr.Message.(string), "Sesi Anda telah berakhir")
}

func TestRoleMiddleware_Allowed(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	// Pre-set role as if JWT middleware passed
	c.Set("role", "super_admin")

	h := RoleMiddleware("super_admin", "admin")(func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	})

	err := h(c)
	assert.NoError(t, err)
}

func TestRoleMiddleware_Forbidden(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	c.Set("role", "staf")

	h := RoleMiddleware("super_admin", "admin")(func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	})

	err := h(c)
	assert.Error(t, err)
	httpErr, _ := err.(*echo.HTTPError)
	assert.Equal(t, http.StatusForbidden, httpErr.Code)
}
