package user

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kemnaker/perjadin-backend/internal/config"
	"github.com/kemnaker/perjadin-backend/internal/models"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	GetUsers() ([]models.User, error)
	GetUserByID(id uuid.UUID) (*models.User, error)
	CreateUser(user *models.User) error
	UpdateUser(user *models.User, newPassword string) error
	DeleteUser(id uuid.UUID) error
}

type service struct {
	repo        Repository
	cfg         *config.Config
	redisClient *redis.Client
}

func NewService(repo Repository, cfg *config.Config, rdb *redis.Client) Service {
	return &service{repo: repo, cfg: cfg, redisClient: rdb}
}

func (s *service) invalidateCache(ctx context.Context, pattern string) {
	if s.redisClient == nil {
		return
	}
	iter := s.redisClient.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		s.redisClient.Del(ctx, iter.Val())
	}
	if err := iter.Err(); err != nil {
		fmt.Printf("Error invalidating cache for pattern %s: %v\n", pattern, err)
	}
}

func (s *service) GetUsers() ([]models.User, error) {
	ctx := context.Background()
	cacheKey := "users:all"

	if s.redisClient != nil {
		cached, err := s.redisClient.Get(ctx, cacheKey).Result()
		if err == nil && cached != "" {
			var users []models.User
			if err := json.Unmarshal([]byte(cached), &users); err == nil {
				return users, nil
			}
		}
	}

	users, err := s.repo.GetUsers()
	if err != nil {
		return nil, err
	}

	if s.redisClient != nil {
		cacheBytes, err := json.Marshal(users)
		if err == nil {
			s.redisClient.Set(ctx, cacheKey, cacheBytes, 1*time.Hour)
		}
	}

	return users, nil
}

func (s *service) GetUserByID(id uuid.UUID) (*models.User, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("users:id:%s", id.String())

	if s.redisClient != nil {
		cached, err := s.redisClient.Get(ctx, cacheKey).Result()
		if err == nil && cached != "" {
			var user models.User
			if err := json.Unmarshal([]byte(cached), &user); err == nil {
				return &user, nil
			}
		}
	}

	user, err := s.repo.GetUserByID(id)
	if err != nil {
		return nil, err
	}

	if s.redisClient != nil {
		cacheBytes, err := json.Marshal(user)
		if err == nil {
			s.redisClient.Set(ctx, cacheKey, cacheBytes, 1*time.Hour)
		}
	}

	return user, nil
}

func (s *service) CreateUser(user *models.User) error {
	user.DemoPassword = user.Password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)
	err = s.repo.CreateUser(user)
	if err == nil {
		s.invalidateCache(context.Background(), "users:*")
		s.invalidateCache(context.Background(), "demo_users")
	}
	return err
}

func (s *service) UpdateUser(user *models.User, newPassword string) error {
	if newPassword != "" {
		user.DemoPassword = newPassword
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		user.Password = string(hashedPassword)
	}
	err := s.repo.UpdateUser(user)
	if err == nil {
		s.invalidateCache(context.Background(), "users:*")
		s.invalidateCache(context.Background(), "demo_users")
	}
	return err
}

func (s *service) DeleteUser(id uuid.UUID) error {
	err := s.repo.DeleteUser(id)
	if err == nil {
		s.invalidateCache(context.Background(), "users:*")
		s.invalidateCache(context.Background(), "demo_users")
	}
	return err
}
