/*
Copyright 2021 Adevinta
*/

package jwt

import (
	"errors"

	libjwt "github.com/dgrijalva/jwt-go"
)

var (
	// ErrEmptyKey when key is empty.
	ErrEmptyKey = errors.New("empty sign key")
	// ErrTokenInvalid when token has expired or has invalid data.
	ErrTokenInvalid = errors.New("invalid token")
)

// Config defines the config for JWT tokens.
type Config struct {
	SigningKey    string
	SigningMethod libjwt.SigningMethod
	KeyFunc       libjwt.Keyfunc
}

// NewJWTConfig creates an instance of a default jwt config.
func NewJWTConfig(key string) Config {
	return Config{
		SigningKey:    key,
		SigningMethod: libjwt.SigningMethodHS256,
		KeyFunc: func(token *libjwt.Token) (interface{}, error) {
			return []byte(key), nil
		},
	}
}

// GenerateToken generates a JWT token for given config and claims.
func (c Config) GenerateToken(claims map[string]interface{}) (string, error) {
	if c.SigningKey == "" {
		return "", ErrEmptyKey
	}

	token := libjwt.NewWithClaims(c.SigningMethod, libjwt.MapClaims(claims))
	return token.SignedString([]byte(c.SigningKey))
}

// ValidateToken returns whether given token is valid or not
// for configuration.
func (c Config) ValidateToken(tokenString string) error {
	token, err := libjwt.Parse(tokenString, c.KeyFunc)
	if err == nil && token.Valid {
		return nil
	}
	return ErrTokenInvalid
}
