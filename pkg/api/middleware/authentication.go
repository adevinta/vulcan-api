/*
Copyright 2021 Adevinta
*/

package middleware

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"

	stdjwt "github.com/dgrijalva/jwt-go"
	jwtkit "github.com/go-kit/kit/auth/jwt"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"

	"github.com/adevinta/errors"
	"github.com/adevinta/vulcan-api/pkg/api"
)

// Helper which returns the private signkey for a Token object
// For now, we are going to use a single general sign key
var keyFunc = func(signKey string) func(token *stdjwt.Token) (interface{}, error) {
	return func(token *stdjwt.Token) (interface{}, error) {
		return []byte(signKey), nil
	}
}

// Authentication retrieves information about:
// 1 - The token used in the current request
// 2 - Information about the user associated with this token
//
// And also
// 3 - Validates if the token is active for the current user
//
// If the authentication succeeds, the email fied is added to the context.
func Authentication(logger log.Logger, signKey string, userRepo api.VulcanitoStore) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			tokenRaw, tokenClaims, err := getTokenInformation(ctx, logger, signKey)
			if err != nil {
				return nil, err
			}

			_ = logger.Log("context", "Authentication", "claims", tokenClaims)
			user, err := getTokenUser(logger, userRepo, tokenClaims)
			if err != nil {
				return nil, err
			}

			// Check if the current token is an API token
			tokenType, ok := tokenClaims["type"].(string)
			if ok && tokenType == "API" {
				// If it is, then validate it
				err = validateAPIToken(tokenRaw, tokenClaims, *user)
				if err != nil {
					return nil, err
				}
			}

			if user.Active != nil && !*user.Active {
				return nil, errors.Unauthorized(fmt.Errorf("user %v with ID %v is disabled", user.Email, user.ID))
			}

			// Store the user in the context
			ctx = api.ContextWithUser(ctx, *user)

			return next(ctx, request)
		}
	}
}

// getTokenInformation extracts the token from the context, parses it (also v
// alidating). and returns the raw token and a map containing the token claims.
func getTokenInformation(ctx context.Context, logger log.Logger, signKey string) (tokenRaw string, tokenClaims stdjwt.MapClaims, err error) {
	// Retrieve the JWT token from the context
	tokenRaw, ok := ctx.Value(jwtkit.JWTTokenContextKey).(string)
	if !ok {
		return "", nil, errors.Unauthorized("Cannot read token")
	}

	// Validate the JWT token against the private signing key
	token, err := stdjwt.Parse(tokenRaw, keyFunc(signKey))
	if err != nil {
		_ = logger.Log("error", err)
		return "", nil, errors.Unauthorized("cannot parse jwt token")
	}

	// Type assertion for token claims
	tokenClaims, ok = token.Claims.(stdjwt.MapClaims)
	if !ok {
		return "", nil, errors.Unauthorized("Cannot map claims")
	}

	return tokenRaw, tokenClaims, nil
}

// getTokenUser extracts the current user email from the tokens claims and
// then look for this user in the database. This function will returns the user
// stored in the database.
func getTokenUser(logger log.Logger, userRepo api.VulcanitoStore, tokenClaims stdjwt.MapClaims) (user *api.User, err error) {
	// Type assertion for email field
	email, ok := tokenClaims["sub"].(string)
	if !ok {
		return nil, errors.Unauthorized("Cannot get sub from token")
	}

	// Retrieve information about the current user
	user, err = userRepo.FindUserByEmail(email)
	if err != nil {
		_ = logger.Log("error", err)
		return nil, errors.Unauthorized("cannot find user")
	}

	return user, nil
}

// validateAPIToken receives an API token and checks if it's valid and active for the user.
// we are storing the hash256 of the token in the database, not the raw token.
// This function will not validate session tokens.
func validateAPIToken(tokenRaw string, tokenClaims stdjwt.MapClaims, user api.User) error {
	hash256 := sha256.Sum256([]byte(tokenRaw))
	tokenBytes, err := hex.DecodeString(user.APIToken)

	if err != nil {
		return errors.Unauthorized("cannot decode token")
	}

	if i := subtle.ConstantTimeCompare(tokenBytes, hash256[:]); i != 1 {
		return errors.Unauthorized("Not a valid API token")
	}

	return nil
}
