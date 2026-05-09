package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"

	"github.com/jlast1o/job-exchange/internal/user"
)

type AuthHandler struct {
	userService *user.Service
}

func NewAuthHandler(userService *user.Service) *AuthHandler {
	return &AuthHandler{
		userService: userService,
	}
}

// Register обрабатывает POST /api/v1/auth/register.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req user.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Инвалидное тело запроса")
		return
	}

	if err := validateRegister(req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := h.userService.Register(r.Context(), req)

	if err != nil {
		switch {
		case errors.Is(err, user.ErrAlreadyExist):
			respondError(w, http.StatusConflict, "Пользователь уже существует")
		default:
			respondError(w, http.StatusInternalServerError, "Внутренняя ошибка")
		}
		return
	}

	respondJSON(w, http.StatusCreated, resp)
}

// Login обрабатывает POST /api/v1/auth/login.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req user.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Инвалидное тело запроса")
		return
	}

	if req.Email == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "Почта или пароль некорректны")
		return
	}

	resp, err := h.userService.Login(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrInvalidCredentials):
			respondError(w, http.StatusBadRequest, "Почта или пароль некорректны")
		case errors.Is(err, user.ErrUserBanned):
			respondError(w, http.StatusForbidden, "Пользователь заблокирован")
		default:
			respondError(w, http.StatusInternalServerError, "Ошибка сервера")
		}
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, ErrorResponse{Error: message})
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func validateRegister(req user.RegisterRequest) error {
	if req.Email == "" || req.Password == "" {
		return fmt.Errorf("email and password are required")
	}
	if len(req.Password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	if !emailRegex.MatchString(req.Email) {
		return fmt.Errorf("invalid email format")
	}
	if req.Role != "" && req.Role != "user" && req.Role != "executor" {
		return fmt.Errorf("invalid role")
	}
	return nil
}
