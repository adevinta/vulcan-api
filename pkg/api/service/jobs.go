/*
Copyright 2021 Adevinta
*/

package service

import (
	"context"

	"github.com/adevinta/errors"
	"github.com/adevinta/vulcan-api/pkg/api"
)

func (s vulcanitoService) FindJob(ctx context.Context, id string) (*api.Job, error) {
	if id == "" {
		return nil, errors.Validation(`ID is empty`)
	}
	return s.db.FindJob(id)
}
