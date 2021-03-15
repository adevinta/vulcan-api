/*
Copyright 2021 Adevinta
*/

package endpoint

import (
	"context"
	"strings"
	"time"

	"github.com/go-kit/kit/endpoint"
	kitlog "github.com/go-kit/kit/log"

	"github.com/adevinta/errors"
	"github.com/adevinta/vulcan-api/pkg/api"
)

// ListProgramScansRequest holds the information passed to the
// ListProgramScans endpoint.
type ListProgramScansRequest struct {
	TeamID    string `urlvar:"team_id"`
	ProgramID string `urlvar:"program_id"`
}

type ScanRequest struct {
	ID            string     `json:"id" urlvar:"scan_id"`
	TeamID        string     `json:"team_id" urlvar:"team_id"`
	ProgramID     string     `json:"program_id"`
	ScheduledTime *time.Time `json:"scheduled_time"`
	EndTime       *time.Time `json:"end_time"`
}

func makeCreateScanEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		scanRequest, ok := request.(*ScanRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}

		user, err := api.UserFromContext(ctx)
		if err != nil {
			return nil, errors.Default(err)
		}
		if scanRequest.ScheduledTime == nil {
			now := time.Now()
			scanRequest.ScheduledTime = &now
		}
		scan := api.Scan{
			ProgramID:     scanRequest.ProgramID,
			ScheduledTime: scanRequest.ScheduledTime,
			RequestedBy:   strings.ToLower(user.Email),
			EndTime:       nil,
		}
		createdScan, err := s.CreateScan(ctx, scan, scanRequest.TeamID)
		if err != nil {
			return nil, err
		}
		return Created{createdScan.ToResponse()}, nil
	}
}

func makeFindScanEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		scanRequest, ok := request.(*ScanRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}

		found, err := s.FindScan(ctx, scanRequest.ID, scanRequest.TeamID)
		if err != nil {
			return nil, err
		}
		return Ok{found.ToResponse()}, nil
	}
}

func makeAbortScanEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		scanRequest, ok := request.(*ScanRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}

		scan, err := s.AbortScan(ctx, scanRequest.ID, scanRequest.TeamID)
		if err != nil {
			return nil, err
		}
		return Ok{scan.ToResponse()}, nil
	}
}
