package handlers

import (
	"encoding/json"
	"log"
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

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	log.Printf("Register request received: %s %s", r.Method, r.URL.Path)

	if r.Method != http.MethodPost {
		log.Printf("Method not allowed: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var credentials models.UserCredentials
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		log.Printf("Failed to decode request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("Registration attempt for login: %s", credentials.Login)

	if credentials.Login == "" || credentials.Password == "" {
		log.Printf("Empty login or password")
		http.Error(w, "Login and password are required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	log.Printf("Checking if user exists: %s", credentials.Login)
	existingUser, err := h.storage.GetUserByLogin(ctx, credentials.Login)
	if err != nil {
		log.Printf("Database error when checking existing user: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if existingUser != nil {
		log.Printf("User already exists: %s", credentials.Login)
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}

	log.Printf("Hashing password for user: %s", credentials.Login)
	hashedPassword, err := h.authService.HashPassword(credentials.Password)
	if err != nil {
		log.Printf("Failed to hash password: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Printf("Creating user: %s", credentials.Login)
	user := &models.User{
		Login:    credentials.Login,
		Password: hashedPassword,
		Created:  time.Now(),
	}

	if err := h.storage.CreateUser(ctx, user); err != nil {
		log.Printf("Failed to create user in database: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Printf("User created successfully with ID: %d", user.ID)

	log.Printf("Generating JWT token for user: %s", credentials.Login)
	token, err := h.authService.GenerateToken(user.ID, user.Login)
	if err != nil {
		log.Printf("Failed to generate JWT token: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // В продакшене должно быть true
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400, // 24 часа
	})

	log.Printf("Registration successful for user: %s", credentials.Login)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "User registered successfully",
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	log.Printf("Login request received: %s %s", r.Method, r.URL.Path)

	if r.Method != http.MethodPost {
		log.Printf("Method not allowed: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var credentials models.UserCredentials
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		log.Printf("Failed to decode request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("Login attempt for user: %s", credentials.Login)

	if credentials.Login == "" || credentials.Password == "" {
		log.Printf("Empty login or password")
		http.Error(w, "Login and password are required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	log.Printf("Getting user by login: %s", credentials.Login)
	user, err := h.storage.GetUserByLogin(ctx, credentials.Login)
	if err != nil {
		log.Printf("Database error when getting user: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if user == nil {
		log.Printf("User not found: %s", credentials.Login)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	log.Printf("Checking password for user: %s", credentials.Login)
	if err := h.authService.CheckPassword(user.Password, credentials.Password); err != nil {
		log.Printf("Invalid password for user: %s", credentials.Login)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	log.Printf("Generating JWT token for user: %s", credentials.Login)
	token, err := h.authService.GenerateToken(user.ID, user.Login)
	if err != nil {
		log.Printf("Failed to generate JWT token: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // В продакшене должно быть true
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400, // 24 часа
	})

	log.Printf("Login successful for user: %s", credentials.Login)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "User logged in successfully",
	})
}
