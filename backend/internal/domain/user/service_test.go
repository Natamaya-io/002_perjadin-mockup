package user

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/kemnaker/perjadin-backend/internal/config"
	"github.com/kemnaker/perjadin-backend/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// MockRepository is a mock implementation of user.Repository
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreateUser(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockRepository) GetUserByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	if args.Get(0) != nil {
		return args.Get(0).(*models.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockRepository) GetUserByID(id uuid.UUID) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) != nil {
		return args.Get(0).(*models.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockRepository) UpdateUser(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockRepository) GetUsers() ([]models.User, error) {
	args := m.Called()
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *MockRepository) DeleteUser(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func TestGetUsers(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo, &config.Config{}, nil)

	users := []models.User{
		{Email: "user1@example.com", Name: "User 1"},
		{Email: "user2@example.com", Name: "User 2"},
	}

	mockRepo.On("GetUsers").Return(users, nil)

	result, err := service.GetUsers()

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "user1@example.com", result[0].Email)

	mockRepo.AssertExpectations(t)
}

func TestGetUserByID(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo, &config.Config{}, nil)

	userID := uuid.New()
	user := &models.User{
		Base:  models.Base{ID: userID},
		Email: "test@example.com",
	}

	mockRepo.On("GetUserByID", userID).Return(user, nil)

	result, err := service.GetUserByID(userID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, userID, result.ID)

	mockRepo.AssertExpectations(t)
}

func TestGetUserByID_NotFound(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo, &config.Config{}, nil)

	userID := uuid.New()

	mockRepo.On("GetUserByID", userID).Return(nil, errors.New("user not found"))

	result, err := service.GetUserByID(userID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "user not found", err.Error())

	mockRepo.AssertExpectations(t)
}

func TestCreateUser(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo, &config.Config{}, nil)

	user := &models.User{
		Email:    "new@example.com",
		Password: "plainpassword",
	}

	mockRepo.On("CreateUser", mock.AnythingOfType("*models.User")).Return(nil)

	err := service.CreateUser(user)

	assert.NoError(t, err)
	assert.Equal(t, "plainpassword", user.DemoPassword)
	assert.NotEqual(t, "plainpassword", user.Password) // Should be hashed
	
	errCompare := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte("plainpassword"))
	assert.NoError(t, errCompare)

	mockRepo.AssertExpectations(t)
}

func TestUpdateUser_WithNewPassword(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo, &config.Config{}, nil)

	user := &models.User{
		Email:    "test@example.com",
		Password: "oldhashedpassword",
	}

	mockRepo.On("UpdateUser", mock.AnythingOfType("*models.User")).Return(nil)

	err := service.UpdateUser(user, "newpassword")

	assert.NoError(t, err)
	assert.Equal(t, "newpassword", user.DemoPassword)
	
	errCompare := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte("newpassword"))
	assert.NoError(t, errCompare)

	mockRepo.AssertExpectations(t)
}

func TestUpdateUser_WithoutNewPassword(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo, &config.Config{}, nil)

	user := &models.User{
		Email:    "test@example.com",
		Password: "oldhashedpassword",
	}

	mockRepo.On("UpdateUser", mock.AnythingOfType("*models.User")).Return(nil)

	err := service.UpdateUser(user, "")

	assert.NoError(t, err)
	assert.Equal(t, "oldhashedpassword", user.Password) // Should remain unchanged

	mockRepo.AssertExpectations(t)
}

func TestDeleteUser(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo, &config.Config{}, nil)

	userID := uuid.New()

	mockRepo.On("DeleteUser", userID).Return(nil)

	err := service.DeleteUser(userID)

	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
}
