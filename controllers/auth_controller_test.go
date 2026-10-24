package controllers

import (
	"go_market_otomasyon/middleware"
	"testing"
	"time"
)

func TestHashPassword(t *testing.T) {
	// "1234" için beklenen SHA-256 hash değeri
	expected := "03ac674216f3e15c761ee1a5e255f067953623c8b388b4459e13f978d7c846f4"
	actual := hashPassword("1234")

	if actual != expected {
		t.Errorf("hashPassword('1234') = %s; beklenen %s", actual, expected)
	}
}

func TestIsValidSuperKey(t *testing.T) {
	cases := []struct {
		key      string
		expected bool
	}{
		{"super123", true},
		{"staj2024", true},
		{"wrongkey", false},
		{"", false},
		{"123456", false},
	}

	for _, c := range cases {
		result := isValidSuperKey(c.key)
		if result != c.expected {
			t.Errorf("isValidSuperKey(%q) = %v; beklenen %v", c.key, result, c.expected)
		}
	}
}

func TestIsSuperRole(t *testing.T) {
	if !middleware.IsSuperRole("superuser") {
		t.Errorf("IsSuperRole('superuser') should be true")
	}
	if !middleware.IsSuperRole("super") {
		t.Errorf("IsSuperRole('super') should be true")
	}
	if middleware.IsSuperRole("normal") {
		t.Errorf("IsSuperRole('normal') should be false")
	}
	if middleware.IsSuperRole("guest") {
		t.Errorf("IsSuperRole('guest') should be false")
	}
}

func TestJWTCreationAndClaims(t *testing.T) {
	username := "testuser"
	role := "superuser"
	tokenStr, err := createJWT(username, role, time.Hour)
	if err != nil {
		t.Fatalf("createJWT failed: %v", err)
	}

	if tokenStr == "" {
		t.Fatalf("tokenString is empty")
	}
}
