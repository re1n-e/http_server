package auth

import (
	"fmt"

	"github.com/alexedwards/argon2id"
)

func HashPassword(password string) (string, error) {
    params := argon2id.DefaultParams
    hash, err := argon2id.CreateHash(password, params)
    if err != nil {
        return "", fmt.Errorf("failed to hash password: %w", err)
    }
    return hash, nil
}

func CheckPasswordHash(password, hash string) (bool, error) {
	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false, fmt.Errorf("failed to verify password: %w", err)
	}
	return match, nil
}
