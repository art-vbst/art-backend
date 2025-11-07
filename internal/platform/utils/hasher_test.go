package utils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetHash(t *testing.T) {
	password := "testPassword123"
	
	hash, err := GetHash(password)
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	
	// Verify hash format
	assert.True(t, strings.HasPrefix(hash, "$argon2id$"))
}

func TestGetHashString_WithCustomParams(t *testing.T) {
	password := "testPassword123"
	params := Argon2Params{
		Time:    2,
		Memory:  16 * 1024,
		Threads: 2,
		KeyLen:  32,
	}
	
	hash, err := GetHashString(password, params)
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.True(t, strings.HasPrefix(hash, "$argon2id$"))
}

func TestVerifyHash_ValidPassword(t *testing.T) {
	password := "testPassword123"
	
	hash, err := GetHash(password)
	require.NoError(t, err)
	
	valid, err := VerifyHash(password, hash)
	require.NoError(t, err)
	assert.True(t, valid)
}

func TestVerifyHash_InvalidPassword(t *testing.T) {
	password := "testPassword123"
	wrongPassword := "wrongPassword456"
	
	hash, err := GetHash(password)
	require.NoError(t, err)
	
	valid, err := VerifyHash(wrongPassword, hash)
	require.NoError(t, err)
	assert.False(t, valid)
}

func TestVerifyHash_InvalidHashFormat(t *testing.T) {
	password := "testPassword123"
	invalidHash := "invalid-hash-format"
	
	valid, err := VerifyHash(password, invalidHash)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidHash, err)
	assert.False(t, valid)
}

func TestVerifyHash_TooFewParts(t *testing.T) {
	password := "testPassword123"
	invalidHash := "$argon2id$v=19"
	
	valid, err := VerifyHash(password, invalidHash)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidHash, err)
	assert.False(t, valid)
}

func TestGetHash_DifferentHashesForSamePassword(t *testing.T) {
	password := "testPassword123"
	
	hash1, err := GetHash(password)
	require.NoError(t, err)
	
	hash2, err := GetHash(password)
	require.NoError(t, err)
	
	// Hashes should be different due to different salts
	assert.NotEqual(t, hash1, hash2)
	
	// But both should verify correctly
	valid1, err := VerifyHash(password, hash1)
	require.NoError(t, err)
	assert.True(t, valid1)
	
	valid2, err := VerifyHash(password, hash2)
	require.NoError(t, err)
	assert.True(t, valid2)
}

func TestVerifyHash_EmptyPassword(t *testing.T) {
	password := ""
	
	hash, err := GetHash(password)
	require.NoError(t, err)
	
	valid, err := VerifyHash(password, hash)
	require.NoError(t, err)
	assert.True(t, valid)
	
	// Different password should not match
	valid, err = VerifyHash("notEmpty", hash)
	require.NoError(t, err)
	assert.False(t, valid)
}

func TestVerifyHash_SpecialCharacters(t *testing.T) {
	password := "!@#$%^&*()_+-=[]{}|;:',.<>?/`~"
	
	hash, err := GetHash(password)
	require.NoError(t, err)
	
	valid, err := VerifyHash(password, hash)
	require.NoError(t, err)
	assert.True(t, valid)
}

func TestVerifyHash_Unicode(t *testing.T) {
	password := "пароль密码🔐"
	
	hash, err := GetHash(password)
	require.NoError(t, err)
	
	valid, err := VerifyHash(password, hash)
	require.NoError(t, err)
	assert.True(t, valid)
}

func TestGetHashString_VerifyParams(t *testing.T) {
	password := "testPassword123"
	params := Argon2Params{
		Time:    5,
		Memory:  64 * 1024,
		Threads: 4,
		KeyLen:  32,
	}
	
	hash, err := GetHashString(password, params)
	require.NoError(t, err)
	
	// Verify the hash contains the correct parameters
	assert.Contains(t, hash, "m=65536")  // 64 * 1024
	assert.Contains(t, hash, "t=5")
	assert.Contains(t, hash, "p=4")
	
	// Verify the password
	valid, err := VerifyHash(password, hash)
	require.NoError(t, err)
	assert.True(t, valid)
}
