package utils

import (
	"strings"
	"testing"
)

func TestGetHash(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "simple password",
			input: "password123",
		},
		{
			name:  "empty string",
			input: "",
		},
		{
			name:  "long password",
			input: strings.Repeat("a", 100),
		},
		{
			name:  "special characters",
			input: "p@ssw0rd!#$%^&*()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := GetHash(tt.input)
			if err != nil {
				t.Fatalf("GetHash() error = %v", err)
			}

			if hash == "" {
				t.Error("GetHash() returned empty hash")
			}

			// Verify hash format
			if !strings.HasPrefix(hash, "$argon2id$") {
				t.Errorf("GetHash() hash doesn't have correct prefix, got = %v", hash)
			}

			// Verify hash has correct number of parts
			parts := strings.Split(hash, "$")
			if len(parts) != 6 {
				t.Errorf("GetHash() hash has wrong number of parts, got = %d, want = 6", len(parts))
			}
		})
	}
}

func TestGetHashString(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		params Argon2Params
	}{
		{
			name:  "with default params",
			input: "password123",
			params: Argon2Params{
				Time:    3,
				Memory:  32 * 1024,
				Threads: 1,
				KeyLen:  32,
			},
		},
		{
			name:  "with custom params",
			input: "password123",
			params: Argon2Params{
				Time:    1,
				Memory:  64 * 1024,
				Threads: 2,
				KeyLen:  16,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := GetHashString(tt.input, tt.params)
			if err != nil {
				t.Fatalf("GetHashString() error = %v", err)
			}

			if hash == "" {
				t.Error("GetHashString() returned empty hash")
			}
		})
	}
}

func TestVerifyHash(t *testing.T) {
	password := "mySecretPassword123"
	hash, err := GetHash(password)
	if err != nil {
		t.Fatalf("Failed to create hash for testing: %v", err)
	}

	tests := []struct {
		name     string
		val      string
		encoded  string
		want     bool
		wantErr  bool
		errorMsg string
	}{
		{
			name:    "correct password",
			val:     password,
			encoded: hash,
			want:    true,
			wantErr: false,
		},
		{
			name:    "incorrect password",
			val:     "wrongPassword",
			encoded: hash,
			want:    false,
			wantErr: false,
		},
		{
			name:     "invalid hash format - too few parts",
			val:      password,
			encoded:  "$argon2id$v=19$m=32768",
			want:     false,
			wantErr:  true,
			errorMsg: "invalid hash format",
		},
		{
			name:     "empty hash",
			val:      password,
			encoded:  "",
			want:     false,
			wantErr:  true,
			errorMsg: "invalid hash format",
		},
		{
			name:    "empty password with valid hash",
			val:     "",
			encoded: hash,
			want:    false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := VerifyHash(tt.val, tt.encoded)
			if (err != nil) != tt.wantErr {
				t.Errorf("VerifyHash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.errorMsg != "" {
				if !strings.Contains(err.Error(), tt.errorMsg) && err != ErrInvalidHash {
					t.Errorf("VerifyHash() error = %v, want error containing %v", err, tt.errorMsg)
				}
			}
			if got != tt.want {
				t.Errorf("VerifyHash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVerifyHash_InvalidBase64(t *testing.T) {
	// Test with invalid base64 in salt
	invalidHash := "$argon2id$v=19$m=32768,t=3,p=1$!!!invalid!!!$validhash"
	got, err := VerifyHash("password", invalidHash)
	if err == nil {
		t.Error("VerifyHash() expected error for invalid base64 salt, got nil")
	}
	if got {
		t.Error("VerifyHash() = true, want false for invalid hash")
	}
}

func TestHashUniqueness(t *testing.T) {
	password := "testPassword"

	hash1, err := GetHash(password)
	if err != nil {
		t.Fatalf("GetHash() error = %v", err)
	}

	hash2, err := GetHash(password)
	if err != nil {
		t.Fatalf("GetHash() error = %v", err)
	}

	// Hashes should be different due to different salts
	if hash1 == hash2 {
		t.Error("GetHash() produced same hash for same password (salts should be different)")
	}

	// But both should verify correctly
	verified1, err := VerifyHash(password, hash1)
	if err != nil {
		t.Fatalf("VerifyHash() error = %v", err)
	}
	if !verified1 {
		t.Error("VerifyHash() failed to verify hash1")
	}

	verified2, err := VerifyHash(password, hash2)
	if err != nil {
		t.Fatalf("VerifyHash() error = %v", err)
	}
	if !verified2 {
		t.Error("VerifyHash() failed to verify hash2")
	}
}
