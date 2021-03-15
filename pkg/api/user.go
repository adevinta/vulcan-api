/*
Copyright 2021 Adevinta
*/

package api

import (
	"context"
	"errors"
	"strings"
	"time"
)

// User ...
type User struct {
	ID        string `gorm:"primary_key:true"`
	Firstname string
	Lastname  string
	Email     string `validate:"required"`
	APIToken  string `gorm:"Column:api_token"`
	Active    *bool  `gorm:"default:true"`
	Admin     *bool  `gorm:"default:false"`
	Observer  *bool  `gorm:"default:false"`
	LastLogin *time.Time
	// A user can belong to multiple teams
	UserTeams []*UserTeam

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (u User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:        u.ID,
		Firstname: u.Firstname,
		Lastname:  u.Lastname,
		Email:     strings.ToLower(u.Email),
		Admin:     u.Admin,
		Observer:  u.Observer,
		Active:    u.Active,
		LastLogin: u.LastLogin,
	}
}

type UserResponse struct {
	ID        string     `json:"id"`
	Firstname string     `json:"firstname"`
	Lastname  string     `json:"lastname"`
	Email     string     `json:"email"`
	Admin     *bool      `json:"admin"`
	Observer  *bool      `json:"observer"`
	Active    *bool      `json:"active"`
	LastLogin *time.Time `json:"last_login"`
}

// Token represents a personal API token
type Token struct {
	Token        string    `json:"token"`
	Email        string    `json:"email"`
	Hash         string    `json:"hash"`
	CreationTime time.Time `json:"creation_time"`
}

// UserStore contains methods to manage teams in data store
type UserStore interface {
	Create(User) (User, error)
	Update(User) (User, error)
	FindByID(string) (User, error)
	FindByEmail(string) (User, error)
}

func ContextWithUser(ctx context.Context, u User) context.Context {
	return context.WithValue(ctx, "email", u)
}

func UserFromContext(ctx context.Context) (User, error) {
	u, ok := ctx.Value("email").(User)
	if !ok {
		return User{}, errors.New("type assertion failed when retrieving User from context")
	}

	return u, nil
}
