package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

type Argon2Params struct {
	Time    uint32
	Memory  uint32
	Threads uint8
	KeyLen  uint32
}

var defaultParams = Argon2Params{
	Time:    3,
	Memory:  32 * 1024,
	Threads: 1,
	KeyLen:  32,
}

const paramsFormat = "m=%d,t=%d,p=%d"
const hashFormat = "$argon2id$v=%d$" + paramsFormat + "$%s$%s"

func GetHash(val string) (string, error) {
	return GetHashString(val, defaultParams)
}

func GetHashString(val string, p Argon2Params) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(val), salt, p.Time, p.Memory, p.Threads, p.KeyLen)

	encoded := fmt.Sprintf(hashFormat,
		argon2.Version, p.Memory, p.Time, p.Threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)

	return encoded, nil
}

func VerifyHash(val, encoded string) (bool, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 {
		return false, errors.New("invalid hash format")
	}

	params := parts[3]
	saltB64 := parts[4]
	hashB64 := parts[5]

	var memory, time uint32
	var threads uint8

	_, err := fmt.Sscanf(params, paramsFormat, &memory, &time, &threads)
	if err != nil {
		return false, err
	}

	salt, err := base64.RawStdEncoding.DecodeString(saltB64)
	if err != nil {
		return false, err
	}

	hash, err := base64.RawStdEncoding.DecodeString(hashB64)
	if err != nil {
		return false, err
	}

	derived := argon2.IDKey([]byte(val), salt, time, memory, threads, uint32(len(hash)))

	if subtle.ConstantTimeCompare(hash, derived) == 1 {
		return true, nil
	}

	return false, nil
}
