package user

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/kemnaker/perjadin-backend/internal/db"
	"github.com/kemnaker/perjadin-backend/internal/models"
)

type Repository interface {
	CreateUser(user *models.User) error
	GetUserByEmail(email string) (*models.User, error)
	GetUserByID(id uuid.UUID) (*models.User, error)
	UpdateUser(user *models.User) error
	GetUsers() ([]models.User, error)
	DeleteUser(id uuid.UUID) error
}

type repository struct {
	q *db.Queries
	d *sql.DB
}

func NewRepository(sqlDB *sql.DB) Repository {
	return &repository{
		q: db.New(sqlDB),
		d: sqlDB,
	}
}

func toNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

func fromNullString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

func fromNullTime(nt sql.NullTime) time.Time {
	if nt.Valid {
		return nt.Time
	}
	return time.Time{}
}

func mapDBUser(dbu db.User) models.User {
	return models.User{
		Base: models.Base{
			ID:        dbu.ID,
			CreatedAt: fromNullTime(dbu.CreatedAt),
			UpdatedAt: fromNullTime(dbu.UpdatedAt),
		},
		Email:        dbu.Email,
		Password:     dbu.Password,
		Name:         dbu.Name,
		Role:         dbu.Role,
		NIP:          fromNullString(dbu.Nip),
		NomorHP:      fromNullString(dbu.NomorHp),
		Pangkat:      fromNullString(dbu.Pangkat),
		Golongan:     fromNullString(dbu.Golongan),
		Jabatan:      fromNullString(dbu.Jabatan),
		TingkatBiaya: fromNullString(dbu.TingkatBiaya),
		SessionID:    fromNullString(dbu.SessionID),
		DemoPassword: fromNullString(dbu.DemoPassword),
	}
}

func (r *repository) CreateUser(user *models.User) error {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	dbu, err := r.q.CreateUser(context.Background(), db.CreateUserParams{
		ID:           user.ID,
		Email:        user.Email,
		Password:     user.Password,
		Name:         user.Name,
		Role:         user.Role,
		Nip:          toNullString(user.NIP),
		NomorHp:      toNullString(user.NomorHP),
		Pangkat:      toNullString(user.Pangkat),
		Golongan:     toNullString(user.Golongan),
		Jabatan:      toNullString(user.Jabatan),
		TingkatBiaya: toNullString(user.TingkatBiaya),
		SessionID:    toNullString(user.SessionID),
		DemoPassword: toNullString(user.DemoPassword),
	})
	if err != nil {
		return err
	}
	*user = mapDBUser(dbu)
	return nil
}

func (r *repository) GetUserByEmail(email string) (*models.User, error) {
	dbu, err := r.q.GetUserByEmail(context.Background(), email)
	if err != nil {
		return nil, err
	}
	u := mapDBUser(dbu)
	return &u, nil
}

func (r *repository) GetUserByID(id uuid.UUID) (*models.User, error) {
	dbu, err := r.q.GetUserByID(context.Background(), id)
	if err != nil {
		return nil, err
	}
	u := mapDBUser(dbu)
	return &u, nil
}

func (r *repository) UpdateUser(user *models.User) error {
	dbu, err := r.q.UpdateUser(context.Background(), db.UpdateUserParams{
		ID:           user.ID,
		Email:        user.Email,
		Password:     user.Password,
		Name:         user.Name,
		Role:         user.Role,
		Nip:          toNullString(user.NIP),
		NomorHp:      toNullString(user.NomorHP),
		Pangkat:      toNullString(user.Pangkat),
		Golongan:     toNullString(user.Golongan),
		Jabatan:      toNullString(user.Jabatan),
		TingkatBiaya: toNullString(user.TingkatBiaya),
		SessionID:    toNullString(user.SessionID),
		DemoPassword: toNullString(user.DemoPassword),
	})
	if err != nil {
		return err
	}
	*user = mapDBUser(dbu)
	return nil
}

func (r *repository) GetUsers() ([]models.User, error) {
	dbus, err := r.q.GetUsers(context.Background())
	if err != nil {
		return nil, err
	}
	users := make([]models.User, len(dbus))
	for i, dbu := range dbus {
		users[i] = mapDBUser(dbu)
	}
	return users, nil
}

func (r *repository) DeleteUser(id uuid.UUID) error {
	return r.q.DeleteUser(context.Background(), id)
}
