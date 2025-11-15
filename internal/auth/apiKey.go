package auth

import (
	"errors"
	"net/http"
	"strings"
)

func GetApiKey(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("authorization header missing")
	}

	const prefix = "ApiKey "
	if !strings.HasPrefix(authHeader, prefix) {
		return "", errors.New("invalid authorization header format")
	}

	token := strings.TrimSpace(strings.TrimPrefix(authHeader, prefix))
	if token == "" {
		return "", errors.New("token missing from authorization header")
	}

	return token, nil
}
