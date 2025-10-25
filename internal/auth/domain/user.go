package domain

import "github.com/google/uuid"

type UserWithHash struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
}

type User struct {
	ID    uuid.UUID
	Email string
}

func StripHash(withHash *UserWithHash) *User {
	return &User{
		ID:    withHash.ID,
		Email: withHash.Email,
	}
}
