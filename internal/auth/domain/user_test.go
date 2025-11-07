package domain

import (
	"testing"

	"github.com/google/uuid"
)

func TestStripHash(t *testing.T) {
	tests := []struct {
		name     string
		withHash *UserWithHash
		wantID   uuid.UUID
		wantEmail string
	}{
		{
			name: "valid user with hash",
			withHash: &UserWithHash{
				ID:           uuid.New(),
				Email:        "test@example.com",
				PasswordHash: "$argon2id$v=19$m=32768,t=3,p=1$salt$hash",
			},
			wantEmail: "test@example.com",
		},
		{
			name: "user with empty hash",
			withHash: &UserWithHash{
				ID:           uuid.New(),
				Email:        "another@example.com",
				PasswordHash: "",
			},
			wantEmail: "another@example.com",
		},
		{
			name: "user with special email characters",
			withHash: &UserWithHash{
				ID:           uuid.New(),
				Email:        "user+tag@sub.example.com",
				PasswordHash: "some-hash",
			},
			wantEmail: "user+tag@sub.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StripHash(tt.withHash)

			if got == nil {
				t.Fatal("StripHash() returned nil")
			}

			if got.ID != tt.withHash.ID {
				t.Errorf("StripHash() ID = %v, want %v", got.ID, tt.withHash.ID)
			}

			if got.Email != tt.wantEmail {
				t.Errorf("StripHash() Email = %v, want %v", got.Email, tt.wantEmail)
			}

			// Verify that the returned User struct doesn't have password hash
			// (by checking it's the right type)
			if _, ok := interface{}(got).(*User); !ok {
				t.Error("StripHash() should return *User type")
			}
		})
	}
}

func TestStripHash_PreservesID(t *testing.T) {
	// Test that UUID is preserved exactly
	originalID := uuid.New()
	withHash := &UserWithHash{
		ID:           originalID,
		Email:        "test@example.com",
		PasswordHash: "hash",
	}

	result := StripHash(withHash)

	if result.ID != originalID {
		t.Errorf("StripHash() didn't preserve UUID: got %v, want %v", result.ID, originalID)
	}
}

func TestStripHash_PreservesEmail(t *testing.T) {
	// Test that email is preserved exactly
	testEmail := "exact.email+tag@example.com"
	withHash := &UserWithHash{
		ID:           uuid.New(),
		Email:        testEmail,
		PasswordHash: "hash",
	}

	result := StripHash(withHash)

	if result.Email != testEmail {
		t.Errorf("StripHash() didn't preserve email: got %v, want %v", result.Email, testEmail)
	}
}

func TestUserStructs(t *testing.T) {
	// Test that User and UserWithHash are compatible
	id := uuid.New()
	email := "test@example.com"
	hash := "password-hash"

	withHash := UserWithHash{
		ID:           id,
		Email:        email,
		PasswordHash: hash,
	}

	// Verify UserWithHash has all expected fields
	if withHash.ID != id {
		t.Errorf("UserWithHash.ID = %v, want %v", withHash.ID, id)
	}
	if withHash.Email != email {
		t.Errorf("UserWithHash.Email = %v, want %v", withHash.Email, email)
	}
	if withHash.PasswordHash != hash {
		t.Errorf("UserWithHash.PasswordHash = %v, want %v", withHash.PasswordHash, hash)
	}

	// Convert to User
	user := StripHash(&withHash)

	// Verify User has correct fields
	if user.ID != id {
		t.Errorf("User.ID = %v, want %v", user.ID, id)
	}
	if user.Email != email {
		t.Errorf("User.Email = %v, want %v", user.Email, email)
	}
}
