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

type ReportRequest struct {
	ID     string `json:"id" urlvar:"scan_id"`
	TeamID string `json:"team_id" urlvar:"team_id"`
	ScanID string `json:"scan_id" urlvar:"scan_id"`
}

func makeCreateReportEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		reportRequest, ok := request.(*ReportRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}

		team, err := s.FindTeam(ctx, reportRequest.TeamID)
		if err != nil {
			return nil, err
		}

		err = s.GenerateReport(ctx, reportRequest.TeamID, team.Name, reportRequest.ScanID, false)
		if err != nil {
			_ = logger.Log("ErrGenerateReport", err)
		}

		return Accepted{}, nil
	}
}

func makeFindReportEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		reportRequest, ok := request.(*ReportRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}

		found, err := s.FindReport(ctx, reportRequest.ScanID)
		if err != nil {
			return nil, err
		}
		return Ok{found.ToResponse()}, nil
	}
}

func makeFindReportEmailEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		reportRequest, ok := request.(*ReportRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}

		found, err := s.FindReport(ctx, reportRequest.ScanID)
		if err != nil {
			return nil, err
		}

		return Ok{found.ToEmailResponse()}, nil
	}
}

func makeSendReportEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		reportRequest, ok := request.(*ReportRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}

		err = s.SendReport(ctx, reportRequest.ScanID, reportRequest.TeamID)
		if err != nil {
			return nil, err
		}

		return Ok{}, nil
	}
}

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
