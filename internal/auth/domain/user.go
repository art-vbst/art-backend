package domain

import "github.com/google/uuid"

type UserWithHash struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
}

type User struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
}

func StripHash(withHash *UserWithHash) *User {
	return &User{
		ID:    withHash.ID,
		Email: withHash.Email,
	}
}
