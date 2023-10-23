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

type SendDigestReportRequest struct {
	TeamID    string `json:"team_id" urlvar:"team_id"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

func makeSendDigestReportEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		sendDigestReportRequest, ok := request.(*SendDigestReportRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}

		if sendDigestReportRequest.StartDate != "" && sendDigestReportRequest.EndDate == "" {
			return nil, errors.Validation("Invalid date range. Provide a valid value for end_date")
		}
		if sendDigestReportRequest.StartDate == "" && sendDigestReportRequest.EndDate != "" {
			return nil, errors.Validation("Invalid date range. Provide a valid value for start_date")
		}

		err = s.SendDigestReport(ctx, sendDigestReportRequest.TeamID, sendDigestReportRequest.StartDate, sendDigestReportRequest.EndDate)
		if err != nil {
			return nil, err
		}

		return Created{}, nil
	}
}
