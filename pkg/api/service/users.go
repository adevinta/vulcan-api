/*
Copyright 2021 Adevinta
*/

package service

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"
	"time"

	"gopkg.in/go-playground/validator.v9"

	"github.com/adevinta/errors"
	"github.com/adevinta/vulcan-api/pkg/api"
)

func (s vulcanitoService) ListUsers(ctx context.Context) ([]*api.User, error) {
	return s.db.ListUsers()
}

func (s vulcanitoService) CreateUser(ctx context.Context, user api.User) (*api.User, error) {
	err := validator.New().Struct(user)
	if err != nil {
		return nil, errors.Validation(err)
	}
	return s.db.CreateUser(user)
}

func (s vulcanitoService) UpdateUser(ctx context.Context, user api.User) (*api.User, error) {
	err := validator.New().Struct(user)
	if err != nil {
		return nil, errors.Validation(err)
	}

	return s.db.UpdateUser(user)
}

func (s vulcanitoService) FindUser(ctx context.Context, id string) (*api.User, error) {
	if id == "" {
		return nil, errors.Validation(`ID is empty`)
	}
	if strings.Contains(id, "@") {
		return s.db.FindUserByEmail(id)
	}
	return s.db.FindUserByID(id)
}

func (s vulcanitoService) DeleteUser(ctx context.Context, id string) error {
	if id == "" {
		return errors.Validation(`ID is empty`)
	}
	return s.db.DeleteUserByID(id)
}

// GenerateAPIToken creates an API token for a given userID
func (s vulcanitoService) GenerateAPIToken(ctx context.Context, userID string) (*api.Token, error) {
	res := api.Token{}

	// Finds the user in the database
	user, err := s.db.FindUserByID(userID)
	if err != nil {
		_ = s.logger.Log("User not found in db", userID)
		return nil, err
	}

	// Retrieve current authenticated user from context
	currentUser, err := api.UserFromContext(ctx)
	if err != nil {
		_ = s.logger.Log(err.Error())
		return nil, errors.Default(err)
	}

	// Return an error if the user is not and admin and emailAuthenticatedUser
	// is not the same as the target user
	user.Email = strings.ToLower(user.Email)
	if currentUser.Admin != nil && !*currentUser.Admin {
		if strings.ToLower(currentUser.Email) != user.Email {
			return nil, errors.Forbidden("Invalid permissions")
		}
	}

	// Get the current time
	tokenGenTime := time.Now()

	// Generates a new JWT token
	token, err := s.jwtConfig.GenerateToken(map[string]interface{}{
		"iat":  tokenGenTime.Unix(),
		"sub":  user.Email,
		"type": "API",
	})
	if err != nil {
		_ = s.logger.Log("error", err)
		return nil, errors.Create(err)
	}

	// Store the token in the database
	user.APIToken = fmt.Sprintf("%x", sha256.Sum256([]byte(token)))
	respUser, err := s.db.UpdateUser(*user)
	if err != nil {
		_ = s.logger.Log("error", err)
		return nil, err
	}

	res.Email = respUser.Email
	res.Token = token
	res.Hash = user.APIToken
	res.CreationTime = tokenGenTime
	return &res, nil
}
