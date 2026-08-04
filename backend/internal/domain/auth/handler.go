package auth

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	svc Service
}

func NewHandler(s Service) *Handler {
	return &Handler{svc: s}
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type RegisterRequest struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required,min=6"`
	Name     string `json:"name" validate:"required"`
	Role     string `json:"role"`
}

func (h *Handler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}

	token, user, requirePasswordChange, err := h.svc.Login(req.Email, req.Password)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"token": token,
		"user":  user,
		"require_password_change": requirePasswordChange,
	})
}

func (h *Handler) Register(c echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}

	// SECURITY: Prevent arbitrary privilege escalation during registration
	// Force all public registrations to default 'staf' role unless configured otherwise
	if req.Role == "super_admin" || req.Role == "admin" {
		req.Role = "staf"
	} else if req.Role == "" {
		req.Role = "staf"
	}

	user, err := h.svc.Register(req.Email, req.Password, req.Name, req.Role)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to register user")
	}

	return c.JSON(http.StatusCreated, user)
}

func (h *Handler) GetDemoUsers(c echo.Context) error {
	demoUsers, err := h.svc.GetDemoUsers()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, demoUsers)
}

type ChangePasswordRequest struct {
	NewPassword string `json:"new_password" validate:"required"`
}

func (h *Handler) ChangePassword(c echo.Context) error {
	var req ChangePasswordRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}

	userIdVal := c.Get("user_id")
	userId, ok := userIdVal.(string)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid user ID")
	}

	err := h.svc.ChangePassword(userId, req.NewPassword)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to change password: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Password updated successfully"})
}

func (h *Handler) GetMe(c echo.Context) error {
	userIDVal := c.Get("user_id")
	userIDStr, ok := userIDVal.(string)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid user ID")
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid user ID format")
	}
	
	user, err := h.svc.GetUserByID(userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "User not found")
	}
	
	return c.JSON(http.StatusOK, user)
}
