/*
Copyright 2021 Adevinta
*/

package endpoint

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	kitlog "github.com/go-kit/kit/log"

	"github.com/adevinta/errors"
	"github.com/adevinta/vulcan-api/pkg/api"
)

type PolicyRequest struct {
	ID     string `json:"id" urlvar:"policy_id"`
	TeamID string `json:"team_id" urlvar:"team_id"`
	Name   string `json:"name"`
	Global *bool  `json:"global"`
}

func makeListPoliciesEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		policyRequest, ok := request.(*PolicyRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}
		policies, err := s.ListPolicies(ctx, policyRequest.TeamID)
		if err != nil {
			return nil, err
		}
		response := []api.PolicyResponse{}
		for _, policy := range policies {
			response = append(response, *policy.ToResponse())
		}
		return Ok{response}, nil
	}
}

func makeCreatePolicyEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		policyRequest, ok := request.(*PolicyRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}
		policy := api.Policy{
			Name:   policyRequest.Name,
			TeamID: policyRequest.TeamID,
		}
		createdPolicy, err := s.CreatePolicy(ctx, policy)
		if err != nil {
			return nil, err
		}
		return Ok{createdPolicy.ToResponse()}, nil
	}
}

func makeFindPolicyEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		policyRequest, ok := request.(*PolicyRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}

		found, err := s.FindPolicy(ctx, policyRequest.ID)
		if err != nil {
			return nil, err
		}

		return Ok{found.ToResponse()}, nil
	}
}

func makeUpdatePolicyEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		policyRequest, ok := request.(*PolicyRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}

		policy := api.Policy{
			ID:     policyRequest.ID,
			Name:   policyRequest.Name,
			TeamID: policyRequest.TeamID,
		}
		updated, err := s.UpdatePolicy(ctx, policy)
		if err != nil {
			return nil, err
		}
		return Ok{updated.ToResponse()}, nil
	}
}

func makeDeletePolicyEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		policyRequest, ok := request.(*PolicyRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}
		policy := api.Policy{
			ID:     policyRequest.ID,
			TeamID: policyRequest.TeamID,
		}
		err = s.DeletePolicy(ctx, policy)
		if err != nil {
			return nil, err
		}
		return NoContent{nil}, nil
	}
}

type ChecktypeSettingRequest struct {
	ID            string  `json:"id" urlvar:"setting_id"`
	TeamID        string  `json:"team_id" urlvar:"team_id"`
	PolicyID      string  `json:"policy_id" urlvar:"policy_id"`
	CheckTypeName string  `json:"checktype_name"`
	Options       *string `json:"options"`
}

func makeListChecktypeSettingEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		checktypeSettingRequest, ok := request.(*ChecktypeSettingRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}
		checktypeSettings, err := s.ListChecktypeSetting(ctx, checktypeSettingRequest.PolicyID)
		if err != nil {
			return nil, err
		}
		checktypeSettingsResponse := []api.ChecktypeSettingResponse{}
		for _, checktypeSetting := range checktypeSettings {
			checktypeSettingsResponse = append(checktypeSettingsResponse, *checktypeSetting.ToResponse())
		}

		return Ok{checktypeSettingsResponse}, nil
	}
}

func makeCreateChecktypeSettingEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		checktypeSettingRequest, ok := request.(*ChecktypeSettingRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}
		checktypeSetting := api.ChecktypeSetting{
			CheckTypeName: checktypeSettingRequest.CheckTypeName,
			PolicyID:      checktypeSettingRequest.PolicyID,
			Options:       checktypeSettingRequest.Options,
		}
		createdChecktypeSetting, err := s.CreateChecktypeSetting(ctx, checktypeSetting)
		if err != nil {
			return nil, err
		}
		return Created{createdChecktypeSetting.ToResponse()}, nil
	}
}

func makeFindChecktypeSettingEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		requestBody, ok := request.(*ChecktypeSettingRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}
		checktypeSetting := api.ChecktypeSetting{
			ID:       requestBody.ID,
			PolicyID: requestBody.PolicyID,
		}
		checktypeSettingr, err := s.FindChecktypeSetting(ctx, checktypeSetting.PolicyID, checktypeSetting.ID)
		if err != nil {
			return nil, errors.NotFound("cannot find setting")
		}
		return Ok{checktypeSettingr.ToResponse()}, nil
	}
}

func makeUpdateChecktypeSettingEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		requestBody, ok := request.(*ChecktypeSettingRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}
		checktypeSetting := api.ChecktypeSetting{
			ID:            requestBody.ID,
			CheckTypeName: requestBody.CheckTypeName,
			Options:       requestBody.Options,
			PolicyID:      requestBody.PolicyID,
		}
		checktypeSettingr, err := s.UpdateChecktypeSetting(ctx, checktypeSetting)
		if err != nil {
			return nil, errors.Update(err)
		}
		return Ok{checktypeSettingr.ToResponse()}, nil
	}
}

func makeDeleteChecktypeSettingEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		requestBody, ok := request.(*ChecktypeSettingRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}

		err := s.DeleteChecktypeSetting(ctx, requestBody.ID)
		if err != nil {
			return nil, errors.Delete(err)
		}
		return NoContent{nil}, nil
	}
}
