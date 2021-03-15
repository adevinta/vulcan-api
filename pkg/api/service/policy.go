/*
Copyright 2021 Adevinta
*/

package service

import (
	"context"

	"gopkg.in/go-playground/validator.v9"

	"github.com/adevinta/errors"
	"github.com/adevinta/vulcan-api/pkg/api"
)

func (s vulcanitoService) CreatePolicy(ctx context.Context, policy api.Policy) (*api.Policy, error) {
	validationErr := validator.New().Struct(policy)
	if validationErr != nil {
		return nil, errors.Validation(validationErr)
	}
	return s.db.CreatePolicy(policy)
}

func (s vulcanitoService) ListPolicies(ctx context.Context, teamID string) ([]*api.Policy, error) {
	return s.db.ListPolicies(teamID)
}

func (s vulcanitoService) FindPolicy(ctx context.Context, policyID string) (*api.Policy, error) {
	return s.db.FindPolicy(policyID)
}

func (s vulcanitoService) UpdatePolicy(ctx context.Context, policy api.Policy) (*api.Policy, error) {
	return s.db.UpdatePolicy(policy)
}

func (s vulcanitoService) DeletePolicy(ctx context.Context, policy api.Policy) error {
	return s.db.DeletePolicy(policy)
}

func (s vulcanitoService) ListChecktypeSetting(ctx context.Context, policyID string) ([]*api.ChecktypeSetting, error) {
	return s.db.ListChecktypeSetting(policyID)
}

func (s vulcanitoService) CreateChecktypeSetting(ctx context.Context, setting api.ChecktypeSetting) (*api.ChecktypeSetting, error) {
	validationErr := setting.Validate()
	if validationErr != nil {
		return nil, errors.Validation(validationErr)
	}
	return s.db.CreateChecktypeSetting(setting)
}

func (s vulcanitoService) FindChecktypeSetting(ctx context.Context, policyID, checktypeSettingID string) (*api.ChecktypeSetting, error) {
	return s.db.FindChecktypeSetting(checktypeSettingID)
}

func (s vulcanitoService) UpdateChecktypeSetting(ctx context.Context, checktypeSetting api.ChecktypeSetting) (*api.ChecktypeSetting, error) {
	validationErr := checktypeSetting.Validate()
	if validationErr != nil {
		return nil, errors.Validation(validationErr)
	}
	return s.db.UpdateChecktypeSetting(checktypeSetting)
}

func (s vulcanitoService) DeleteChecktypeSetting(ctx context.Context, checktypeSettingID string) error {
	return s.db.DeleteChecktypeSetting(checktypeSettingID)
}
