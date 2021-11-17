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

type RecipientsData struct {
	TeamID string   `json:"team_id" urlvar:"team_id"`
	Emails []string `json:"emails"`
}

func makeListRecipientsEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(*RecipientsData)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}

		recipients, err := s.ListRecipients(ctx, req.TeamID)
		if err != nil {
			return nil, err
		}

		response := []*api.RecipientResponse{}
		for _, recipient := range recipients {
			response = append(response, recipient.ToResponse())
		}

		return Ok{response}, nil
	}
}

func makeUpdateRecipientsEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		requestBody, ok := request.(*RecipientsData)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}

		if err := s.UpdateRecipients(ctx, requestBody.TeamID, requestBody.Emails); err != nil {
			return nil, err
		}

		recipients, err := s.ListRecipients(ctx, requestBody.TeamID)
		if err != nil {
			return nil, err
		}

		response := []*api.RecipientResponse{}
		for _, recipient := range recipients {
			response = append(response, recipient.ToResponse())
		}

		return Ok{response}, nil
	}
}
