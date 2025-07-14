package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/artemaris/loyalty/internal/models"
	"github.com/artemaris/loyalty/internal/services"
	"github.com/artemaris/loyalty/internal/storage"
)

type AuthHandler struct {
	storage     storage.Storage
	authService *services.AuthService
}

func NewAuthHandler(storage storage.Storage, authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		storage:     storage,
		authService: authService,
	}
}

// Register обрабатывает регистрацию пользователя
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var credentials models.UserCredentials
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Проверяем, что логин и пароль не пустые
	if credentials.Login == "" || credentials.Password == "" {
		http.Error(w, "Login and password are required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Проверяем, существует ли пользователь с таким логином
	existingUser, err := h.storage.GetUserByLogin(ctx, credentials.Login)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if existingUser != nil {
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}

	// Хешируем пароль
	hashedPassword, err := h.authService.HashPassword(credentials.Password)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Создаем пользователя
	user := &models.User{
		Login:    credentials.Login,
		Password: hashedPassword,
		Created:  time.Now(),
	}

	if err := h.storage.CreateUser(ctx, user); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Генерируем JWT токен
	token, err := h.authService.GenerateToken(user.ID, user.Login)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Устанавливаем токен в cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // В продакшене должно быть true
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400, // 24 часа
	})

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "User registered successfully",
	})
}

// Login обрабатывает вход пользователя
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var credentials models.UserCredentials
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Проверяем, что логин и пароль не пустые
	if credentials.Login == "" || credentials.Password == "" {
		http.Error(w, "Login and password are required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Получаем пользователя по логину
	user, err := h.storage.GetUserByLogin(ctx, credentials.Login)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Проверяем пароль
	if err := h.authService.CheckPassword(user.Password, credentials.Password); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Генерируем JWT токен
	token, err := h.authService.GenerateToken(user.ID, user.Login)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Устанавливаем токен в cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // В продакшене должно быть true
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400, // 24 часа
	})

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "User logged in successfully",
	})
}
