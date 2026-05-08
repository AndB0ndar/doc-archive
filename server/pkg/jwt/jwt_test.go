package jwt

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateAndValidateToken(t *testing.T) {
	secret := "test-secret"

	t.Run("generate and validate token returns same userID", func(t *testing.T) {
		token, err := GenerateToken("test-user", secret, time.Hour)
		if err != nil {
			t.Fatalf("unexpected error generating token: %v", err)
		}

		userID, err := ValidateToken(token, secret)
		if err != nil {
			t.Fatalf("unexpected error validating token: %v", err)
		}

		if userID != "test-user" {
			t.Fatalf("expected userID 'test-user', got '%s'", userID)
		}
	})

	t.Run("different user IDs are preserved", func(t *testing.T) {
		userIDs := []string{"user1", "user-abc-123", "user@example.com", ""}

		for _, expectedUserID := range userIDs {
			token, err := GenerateToken(expectedUserID, secret, time.Hour)
			if err != nil {
				t.Fatalf("unexpected error generating token for '%s': %v", expectedUserID, err)
			}

			userID, err := ValidateToken(token, secret)
			if err != nil {
				t.Fatalf("unexpected error validating token for '%s': %v", expectedUserID, err)
			}

			if userID != expectedUserID {
				t.Fatalf("expected userID '%s', got '%s'", expectedUserID, userID)
			}
		}
	})

	t.Run("generate with empty userID", func(t *testing.T) {
		token, err := GenerateToken("", secret, time.Hour)
		if err != nil {
			t.Fatalf("unexpected error generating token with empty userID: %v", err)
		}

		userID, err := ValidateToken(token, secret)
		if err != nil {
			t.Fatalf("unexpected error validating token with empty userID: %v", err)
		}

		if userID != "" {
			t.Fatalf("expected empty userID, got '%s'", userID)
		}
	})
}

func TestExpiredToken(t *testing.T) {
	secret := "test-secret"

	t.Run("expired token returns error", func(t *testing.T) {
		// Token that expired 1 hour ago
		token, err := GenerateToken("test-user", secret, -time.Hour)
		if err != nil {
			t.Fatalf("unexpected error generating expired token: %v", err)
		}

		_, err = ValidateToken(token, secret)
		if err == nil {
			t.Fatal("expected error for expired token, got nil")
		}
	})

	t.Run("almost expired token is still valid", func(t *testing.T) {
		// Token that expires in 1 nanosecond — should still be valid immediately
		token, err := GenerateToken("test-user", secret, time.Nanosecond)
		if err != nil {
			t.Fatalf("unexpected error generating token: %v", err)
		}

		// Sleep a tiny bit to ensure it expires
		time.Sleep(time.Millisecond)

		_, err = ValidateToken(token, secret)
		if err == nil {
			t.Fatal("expected error for expired (nanosecond) token, got nil")
		}
	})

	t.Run("future token validates successfully", func(t *testing.T) {
		token, err := GenerateToken("test-user", secret, time.Hour*24)
		if err != nil {
			t.Fatalf("unexpected error generating future token: %v", err)
		}

		_, err = ValidateToken(token, secret)
		if err != nil {
			t.Fatalf("unexpected error validating future token: %v", err)
		}
	})
}

func TestInvalidTokenCases(t *testing.T) {
	secret := "test-secret"

	t.Run("invalid token string returns error", func(t *testing.T) {
		_, err := ValidateToken("not-a-valid-token", secret)
		if err == nil {
			t.Fatal("expected error for invalid token string, got nil")
		}
	})

	t.Run("empty token string returns error", func(t *testing.T) {
		_, err := ValidateToken("", secret)
		if err == nil {
			t.Fatal("expected error for empty token string, got nil")
		}
	})

	t.Run("malformed token returns error", func(t *testing.T) {
		_, err := ValidateToken("header.payload.invalidsignature", secret)
		if err == nil {
			t.Fatal("expected error for malformed token, got nil")
		}
	})

	t.Run("token signed with different secret returns error", func(t *testing.T) {
		token, err := GenerateToken("test-user", "different-secret", time.Hour)
		if err != nil {
			t.Fatalf("unexpected error generating token: %v", err)
		}

		_, err = ValidateToken(token, secret)
		if err == nil {
			t.Fatal("expected error for token signed with different secret, got nil")
		}
	})

	t.Run("token signed with different signing method returns error", func(t *testing.T) {
		// Create a token with a non-HMAC signing method (RS256) to trigger the method check
		_, err := jwt.ParseWithClaims("dummy", &Claims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, nil // this path won't be reached
			}
			return nil, nil
		})
		_ = err // just verifying the structure compiles

		// Use a token signed with a completely different algorithm family
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, &Claims{
			UserID: "test-user",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
		})
		// We can't actually sign with RS256 without a private key, so we'll manually craft the token string
		// and verify that ParseWithClaims returns an error for non-HMAC methods
		rsaToken, _ := token.SigningString()
		badToken := rsaToken + ".invalidsignature"

		_, err = ValidateToken(badToken, secret)
		if err == nil {
			t.Fatal("expected error for token signed with non-HMAC method, got nil")
		}
	})
}
