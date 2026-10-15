package auth

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestMakeAndValidateJWT(t *testing.T) {
	userID := uuid.New()
	secret := "test-secret"

	token, err := MakeJWT(userID, secret, time.Hour)
	if err != nil {
		t.Fatalf("MakeJWT() error = %v", err)
	}
	if token == "" {
		t.Fatal("MakeJWT() returned empty token")
	}

	got, err := ValidateJWT(token, secret)
	if err != nil {
		t.Fatalf("ValidateJWT() error = %v", err)
	}
	if got != userID {
		t.Errorf("ValidateJWT() = %v, want %v", got, userID)
	}
}

func TestValidateJWT_Expired(t *testing.T) {
	secret := "test-secret"

	token, err := MakeJWT(uuid.New(), secret, -time.Minute)
	if err != nil {
		t.Fatalf("MakeJWT() error = %v", err)
	}

	got, err := ValidateJWT(token, secret)
	if err == nil {
		t.Fatal("ValidateJWT() expected error for expired token, got nil")
	}
	if !errors.Is(err, jwt.ErrTokenExpired) {
		t.Errorf("ValidateJWT() error = %v, want %v", err, jwt.ErrTokenExpired)
	}
	if got != uuid.Nil {
		t.Errorf("ValidateJWT() = %v, want uuid.Nil", got)
	}
}

func TestValidateJWT_WrongSecret(t *testing.T) {
	token, err := MakeJWT(uuid.New(), "correct-secret", time.Hour)
	if err != nil {
		t.Fatalf("MakeJWT() error = %v", err)
	}

	got, err := ValidateJWT(token, "wrong-secret")
	if err == nil {
		t.Fatal("ValidateJWT() expected error for wrong secret, got nil")
	}
	if !errors.Is(err, jwt.ErrTokenSignatureInvalid) {
		t.Errorf("ValidateJWT() error = %v, want %v", err, jwt.ErrTokenSignatureInvalid)
	}
	if got != uuid.Nil {
		t.Errorf("ValidateJWT() = %v, want uuid.Nil", got)
	}
}

func TestValidateJWT_Malformed(t *testing.T) {
	if _, err := ValidateJWT("not-a-jwt", "secret"); err == nil {
		t.Fatal("ValidateJWT() expected error for malformed token, got nil")
	}
}
