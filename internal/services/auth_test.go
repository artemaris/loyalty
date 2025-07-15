package services

import (
	"testing"
)

func TestAuthService(t *testing.T) {
	authService := NewAuthService("test-secret")

	password := "testpassword"
	hashedPassword, err := authService.HashPassword(password)
	if err != nil {
		t.Errorf("Failed to hash password: %v", err)
	}

	if hashedPassword == password {
		t.Error("Password was not hashed")
	}

	err = authService.CheckPassword(hashedPassword, password)
	if err != nil {
		t.Errorf("Failed to check correct password: %v", err)
	}

	err = authService.CheckPassword(hashedPassword, "wrongpassword")
	if err == nil {
		t.Error("Should fail for wrong password")
	}

	userID := int64(123)
	login := "testuser"
	token, err := authService.GenerateToken(userID, login)
	if err != nil {
		t.Errorf("Failed to generate token: %v", err)
	}

	if token == "" {
		t.Error("Generated token is empty")
	}

	claims, err := authService.ValidateToken(token)
	if err != nil {
		t.Errorf("Failed to validate token: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("Expected user ID %d, got %d", userID, claims.UserID)
	}

	if claims.Login != login {
		t.Errorf("Expected login %s, got %s", login, claims.Login)
	}

	_, err = authService.ValidateToken("invalid-token")
	if err == nil {
		t.Error("Should fail for invalid token")
	}
}

func TestGenerateRandomString(t *testing.T) {
	length := 32
	randomString, err := GenerateRandomString(length)
	if err != nil {
		t.Errorf("Failed to generate random string: %v", err)
	}

	if len(randomString) != length*2 {
		t.Errorf("Expected length %d, got %d", length*2, len(randomString))
	}

	randomString2, err := GenerateRandomString(length)
	if err != nil {
		t.Errorf("Failed to generate second random string: %v", err)
	}

	if randomString == randomString2 {
		t.Error("Generated strings should be different")
	}
}
