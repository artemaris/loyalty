package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/artemaris/loyalty/internal/models"
)

func TestPostgresStorage(t *testing.T) {
	databaseURI := "postgres://test:test@localhost:5432/loyalty_test?sslmode=disable"

	storage, err := New(databaseURI)
	if err != nil {
		t.Skipf("Skipping test: failed to connect to test database: %v", err)
	}
	defer storage.Close()

	ctx := context.Background()

	user := &models.User{
		Login:    "testuser",
		Password: "hashedpassword",
		Created:  time.Now(),
	}

	err = storage.CreateUser(ctx, user)
	if err != nil {
		t.Errorf("Failed to create user: %v", err)
	}

	foundUser, err := storage.GetUserByLogin(ctx, "testuser")
	if err != nil {
		t.Errorf("Failed to get user by login: %v", err)
	}
	if foundUser == nil {
		t.Error("User not found")
	}

	foundUserByID, err := storage.GetUserByID(ctx, user.ID)
	if err != nil {
		t.Errorf("Failed to get user by ID: %v", err)
	}
	if foundUserByID == nil {
		t.Error("User not found by ID")
	}
}
