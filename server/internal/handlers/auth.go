package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/AndB0ndar/doc-archive/internal/domain"
	"github.com/AndB0ndar/doc-archive/internal/models"
)

type AuthHandler struct {
	authService domain.AuthService
	logger      *slog.Logger
}

func NewAuthHandler(
	authService domain.AuthService, logger *slog.Logger,
) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
	}
}

// Register регистрирует нового пользователя.
// @Summary      Регистрация пользователя
// @Description  Создаёт нового пользователя с email и паролем.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body models.RegisterRequest true "Данные для регистрации"
// @Success      201  {object}  models.AuthResponse
// @Failure      400  {object}  map[string]string
// @Failure      409  {object}  map[string]string
// @Router       /register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	logger := h.logger.With(
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
	)
	logger.Debug("Register request started")

	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn("Register: invalid request body", slog.Any("error", err))
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	logger = logger.With(slog.String("email", req.Email))
	logger.Debug("Register: request body decoded")

	if req.Email == "" || req.Password == "" {
		logger.Warn("Register: missing email or password")
		http.Error(w, "Email and password required", http.StatusBadRequest)
		return
	}
	logger.Debug("Register: validation passed")

	logger.Debug("Register: calling authService.Register")
	token, user, err := h.authService.Register(
		r.Context(), req.Email, req.Password,
	)
	if err != nil {
		logger.Error("Register: registration failed", slog.Any("error", err))
		http.Error(
			w, "Registration failed: "+err.Error(),
			http.StatusInternalServerError,
		)
		return
	}
	logger = logger.With(slog.String("user_id", user.ID))
	logger.Debug("Register: user created successfully")

	resp := models.AuthResponse{
		Token: token,
		User: models.UserResponse{
			ID:    user.ID,
			Email: user.Email,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	logger.Debug("Register: encoding response")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logger.Error(
			"Register: failed to encode response", slog.Any("error", err),
		)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	logger.Info(
		"Register: completed", slog.Duration("duration", time.Since(start)),
	)
}

// Login аутентифицирует пользователя и возвращает JWT-токен.
// @Summary      Вход в систему
// @Description  Аутентификация пользователя, получение JWT.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body models.LoginRequest true "Учётные данные"
// @Success      200  {object}  models.AuthResponse
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Router       /login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	logger := h.logger.With(
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
	)
	logger.Debug("Login request started")

	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn("Login: invalid request body", slog.Any("error", err))
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	logger = logger.With(slog.String("email", req.Email))
	logger.Debug("Login: request body decoded")

	logger.Debug("Login: calling authService.Login")
	token, user, err := h.authService.Login(
		r.Context(), req.Email, req.Password,
	)
	if err != nil {
		logger.Warn("Login: invalid credentials", slog.Any("error", err))
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}
	logger = logger.With(slog.String("user_id", user.ID))
	logger.Debug("Login: authentication successful")

	resp := models.AuthResponse{
		Token: token,
		User: models.UserResponse{
			ID:    user.ID,
			Email: user.Email,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	logger.Debug("Login: encoding response")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logger.Error(
			"Login: failed to encode response", slog.Any("error", err),
		)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	logger.Info(
		"Login: completed", slog.Duration("duration", time.Since(start)),
	)
}
