/*
Copyright 2021 Adevinta
*/

package endpoint

import (
	"context"
	"regexp"

	"github.com/adevinta/errors"
	"github.com/go-kit/kit/endpoint"
	kitlog "github.com/go-kit/kit/log"
	"github.com/adevinta/vulcan-api/pkg/api"
)

const (
	// Regular expression matching date format 'yyyy-mm-dd'.
	dateFmtRegEx = `^\d{4}\-(0[1-9]|1[012])\-(0[1-9]|[12][0-9]|3[01])$`
)

type StatsRequest struct {
	TeamID  string `json:"team_id" urlvar:"team_id"`
	MinDate string `urlquery:"minDate"`
	MaxDate string `urlquery:"maxDate"`
	AtDate  string `urlquery:"atDate"`
}

func makeStatsMTTREndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		r, ok := request.(*StatsRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}

		if !isValidMTTRRequest(r) {
			return nil, errors.Validation("Invalid query params")
		}

		team, err := s.FindTeam(ctx, r.TeamID)
		if err != nil {
			return nil, err
		}
		if team.Tag == "" {
			return nil, errors.Validation("no tag defined for the team")
		}

		params := buildStatsParams(team.Tag, r)

		response, err = s.StatsMTTR(ctx, params)
		if err != nil {
			return nil, err
		}
		return Ok{response}, nil
	}
}

func makeStatsOpenEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		r, ok := request.(*StatsRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}

		if !isValidStatsRequest(r) {
			return nil, errors.Validation("Invalid query params")
		}

		team, err := s.FindTeam(ctx, r.TeamID)
		if err != nil {
			return nil, err
		}
		if team.Tag == "" {
			return nil, errors.Validation("no tag defined for the team")
		}

		params := buildStatsParams(team.Tag, r)

		response, err = s.StatsOpen(ctx, params)
		if err != nil {
			return nil, err
		}
		return Ok{response}, nil
	}
}

func makeStatsFixedEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		r, ok := request.(*StatsRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}

		if !isValidStatsRequest(r) {
			return nil, errors.Validation("Invalid query params")
		}

		team, err := s.FindTeam(ctx, r.TeamID)
		if err != nil {
			return nil, err
		}
		if team.Tag == "" {
			return nil, errors.Validation("no tag defined for the team")
		}

		params := buildStatsParams(team.Tag, r)

		response, err = s.StatsFixed(ctx, params)
		if err != nil {
			return nil, err
		}
		return Ok{response}, nil
	}
}

func makeGlobalStatsMTTREndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		r, ok := request.(*StatsRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}

		if !isValidMTTRRequest(r) {
			return nil, errors.Validation("Invalid query params")
		}

		// Build stats param with void tag
		// so we get global metrics instead
		// of specific team metrics.
		params := buildStatsParams("", r)

		response, err = s.StatsMTTR(ctx, params)
		if err != nil {
			return nil, err
		}
		return Ok{response}, nil
	}
}

type StatsCoverageRequest struct {
	TeamID string `json:"team_id" urlvar:"team_id"`
}

func makeStatsCoverageEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		r, ok := request.(*StatsCoverageRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}

		coverage, err := s.StatsCoverage(ctx, r.TeamID)
		if err != nil {
			return nil, err
		}

		return Ok{coverage}, nil
	}
}

func isValidStatsRequest(r *StatsRequest) bool {
	return (r.MinDate == "" || isValidDate(r.MinDate)) &&
		(r.MaxDate == "" || isValidDate(r.MaxDate)) &&
		(r.AtDate == "" || isValidDate(r.AtDate))
}

func isValidMTTRRequest(r *StatsRequest) bool {
	return (r.MinDate == "" || isValidDate(r.MinDate)) &&
		(r.MaxDate == "" || isValidDate(r.MaxDate)) &&
		r.AtDate == ""
}

func buildStatsParams(tag string, r *StatsRequest) api.StatsParams {
	return api.StatsParams{
		Tag:     tag,
		MinDate: r.MinDate,
		MaxDate: r.MaxDate,
		AtDate:  r.AtDate,
	}
}

func isValidDate(date string) bool {
	return regexp.MustCompile(dateFmtRegEx).MatchString(date)
}
