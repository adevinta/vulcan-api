/*
Copyright 2021 Adevinta
*/

package service

import (
	"context"

	"github.com/adevinta/errors"
	"github.com/adevinta/vulcan-api/pkg/api"
	"gopkg.in/go-playground/validator.v9"
)

func (s vulcanitoService) FindJob(ctx context.Context, id string) (*api.Job, error) {
	if id == "" {
		return nil, errors.Validation(`ID is empty`)
	}
	return s.db.FindJob(id)
}

func (s vulcanitoService) UpdateJob(ctx context.Context, job api.Job) (*api.Job, error) {
	err := validator.New().Struct(job)
	if err != nil {
		return nil, errors.Validation(err)
	}

	return s.db.UpdateJob(job)
}
