package domain_test

import (
	"testing"

	"github.com/AndB0ndar/doc-archive/internal/domain"
)

func TestNewUser(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		password    string
		wantErr     bool
		errContains string
	}{
		{
			name:     "valid user",
			email:    "test@example.com",
			password: "password123",
			wantErr:  false,
		},
		{
			name:        "empty email",
			email:       "",
			password:    "password123",
			wantErr:     true,
			errContains: "email required",
		},
		{
			name:     "empty password",
			email:    "test@example.com",
			password: "",
			wantErr:  false, // bcrypt accepts empty password
		},
		{
			name:     "email with spaces",
			email:    "test user@example.com",
			password: "password123",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := domain.NewUser(tt.email, tt.password)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error but got nil")
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.errContains)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if user == nil {
				t.Fatal("user is nil")
			}

			if user.Email != tt.email {
				t.Errorf("email = %q, want %q", user.Email, tt.email)
			}

			if user.ID == "" {
				t.Error("user ID should not be empty")
			}

			if user.PasswordHash == "" {
				t.Error("password hash should not be empty")
			}
		})
	}
}

func TestUser_ValidatePassword(t *testing.T) {
	// Create a test user with a known password
	user, err := domain.NewUser("test@example.com", "correctpassword")
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	tests := []struct {
		name     string
		password string
		want     bool
	}{
		{
			name:     "correct password",
			password: "correctpassword",
			want:     true,
		},
		{
			name:     "incorrect password",
			password: "wrongpassword",
			want:     false,
		},
		{
			name:     "empty password",
			password: "",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := user.ValidatePassword(tt.password)
			if got != tt.want {
				t.Errorf("ValidatePassword(%q) = %v, want %v", tt.password, got, tt.want)
			}
		})
	}
}

func TestUser_PasswordHashing(t *testing.T) {
	// Test that passwords are properly hashed
	password := "SuperSecret123!"

	user1, err := domain.NewUser("user1@example.com", password)
	if err != nil {
		t.Fatalf("failed to create user1: %v", err)
	}

	user2, err := domain.NewUser("user2@example.com", password)
	if err != nil {
		t.Fatalf("failed to create user2: %v", err)
	}

	// Same password should result in different hashes (due to different salt)
	if user1.PasswordHash == user2.PasswordHash {
		t.Error("same password should have different hashes")
	}

	// Both should validate with the correct password
	if !user1.ValidatePassword(password) {
		t.Error("user1 should validate with correct password")
	}
	if !user2.ValidatePassword(password) {
		t.Error("user2 should validate with correct password")
	}

	// Neither should validate with wrong password
	if user1.ValidatePassword("wrong") {
		t.Error("user1 should not validate with wrong password")
	}
	if user2.ValidatePassword("wrong") {
		t.Error("user2 should not validate with wrong password")
	}
}

func TestUser_UniqueIDs(t *testing.T) {
	// Test that new users get unique IDs
	user1, err := domain.NewUser("user1@example.com", "pass1")
	if err != nil {
		t.Fatalf("failed to create user1: %v", err)
	}

	user2, err := domain.NewUser("user2@example.com", "pass2")
	if err != nil {
		t.Fatalf("failed to create user2: %v", err)
	}

	if user1.ID == user2.ID {
		t.Error("different users should have different IDs")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && contains(s[1:], substr))
}
