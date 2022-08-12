/*
Copyright 2021 Adevinta
*/

package endpoint

import (
	"context"
	"regexp"
	"strings"

	"github.com/adevinta/errors"
	"github.com/adevinta/vulcan-api/pkg/api"
	"github.com/go-kit/kit/endpoint"
	kitlog "github.com/go-kit/kit/log"
)

const (
	// Regular expression matching date format 'yyyy-mm-dd'.
	dateFmtRegEx = `^\d{4}\-(0[1-9]|1[012])\-(0[1-9]|[12][0-9]|3[01])$`
)

type StatsRequest struct {
	TeamID      string  `json:"team_id" urlvar:"team_id"`
	Teams       string  `urlquery:"teams"`
	MinDate     string  `urlquery:"minDate"`
	MaxDate     string  `urlquery:"maxDate"`
	AtDate      string  `urlquery:"atDate"`
	MinScore    float64 `urlquery:"minScore"`
	MaxScore    float64 `urlquery:"maxScore"`
	Identifiers string  `urlquery:"identifiers"`
	Labels      string  `urlquery:"labels"`
}

type GlobalStatsRequest struct {
	Tags string `urlquery:"tags"`
	StatsRequest
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

		params := buildStatsParams(r)

		response, err = s.StatsMTTR(ctx, params)
		if err != nil {
			return nil, err
		}
		return Ok{response}, nil
	}
}

func makeStatsExposureEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		r, ok := request.(*StatsRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}

		if !isValidExposureRequest(r) {
			return nil, errors.Validation("Invalid query params")
		}

		params := buildStatsParams(r)

		response, err = s.StatsExposure(ctx, params)
		if err != nil {
			return nil, err
		}
		return Ok{response}, nil
	}
}

func makeStatsCurrentExposureEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		r, ok := request.(*StatsRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}

		params := buildStatsParams(r)

		response, err = s.StatsCurrentExposure(ctx, params)
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

		params := buildStatsParams(r)

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

		params := buildStatsParams(r)

		response, err = s.StatsFixed(ctx, params)
		if err != nil {
			return nil, err
		}
		return Ok{response}, nil
	}
}

func makeGlobalStatsMTTREndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		r, ok := request.(*GlobalStatsRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}

		if !isValidMTTRRequest(&r.StatsRequest) {
			return nil, errors.Validation("Invalid query params")
		}

		var teamsFilter string
		// Only admin and observer users can set Tags query parameter.
		if r.Tags != "" {
			authorized, err := isAuthorizedTagsParam(ctx)
			if err != nil {
				return nil, err
			}
			if !authorized {
				return nil, errors.Forbidden("User is not allowed to set Tags parameter")
			}
			// The findings are stored by team so we must translate the tags
			// to team ids.
			teams, err := tagsToTeams(ctx, s, r.Tags)
			if err != nil {
				return nil, err
			}
			if len(teams) == 0 {
				return nil, errors.NotFound("There are no teams with the specified tags")
			}
			teamsFilter = teams
		}

		params := buildGlobalStatsParams(teamsFilter, r)

		response, err = s.StatsMTTR(ctx, params)
		if err != nil {
			return nil, err
		}
		return Ok{response}, nil
	}
}

func makeGlobalStatsExposureEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		r, ok := request.(*GlobalStatsRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}

		if !isValidExposureRequest(&r.StatsRequest) {
			return nil, errors.Validation("Invalid query params")
		}

		var teamsFilter string
		// Only admin and observer users can set Tags query parameter.
		if r.Tags != "" {
			authorized, err := isAuthorizedTagsParam(ctx)
			if err != nil {
				return nil, err
			}
			if !authorized {
				return nil, errors.Forbidden("User is not allowed to set Tags parameter")
			}
			// The findings are stored by team so we must translate the tags
			// to team ids.
			teams, err := tagsToTeams(ctx, s, r.Tags)
			if err != nil {
				return nil, err
			}
			if len(teams) == 0 {
				return nil, errors.NotFound("There are no teams with the specified tags")
			}
			teamsFilter = teams
		}

		params := buildGlobalStatsParams(teamsFilter, r)

		response, err = s.StatsExposure(ctx, params)
		if err != nil {
			return nil, err
		}
		return Ok{response}, nil
	}
}

func makeGlobalStatsCurrentExposureEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		r, ok := request.(*GlobalStatsRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}

		var teamsFilter string
		// Only admin and observer users can set Tags query parameter.
		if r.Tags != "" {
			authorized, err := isAuthorizedTagsParam(ctx)
			if err != nil {
				return nil, err
			}
			if !authorized {
				return nil, errors.Forbidden("User is not allowed to set Tags parameter")
			}
			// The findings are stored by team so we must translate the tags
			// to team ids.
			teams, err := tagsToTeams(ctx, s, r.Tags)
			if err != nil {
				return nil, err
			}
			if len(teams) == 0 {
				return nil, errors.NotFound("There are no teams with the specified tags")
			}
			teamsFilter = teams
		}

		params := buildGlobalStatsParams(teamsFilter, r)

		response, err = s.StatsCurrentExposure(ctx, params)
		if err != nil {
			return nil, err
		}
		return Ok{response}, nil
	}
}

func makeGlobalStatsOpenEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		r, ok := request.(*GlobalStatsRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}

		var teamsFilter string
		// Only admin and observer users can set Tags query parameter.
		if r.Tags != "" {
			authorized, err := isAuthorizedTagsParam(ctx)
			if err != nil {
				return nil, err
			}
			if !authorized {
				return nil, errors.Forbidden("User is not allowed to set Tags parameter")
			}
			// The findings are stored by team so we must translate the tags
			// to team ids.
			teams, err := tagsToTeams(ctx, s, r.Tags)
			if err != nil {
				return nil, err
			}
			if len(teams) == 0 {
				return nil, errors.NotFound("There are no teams with the specified tags")
			}
			teamsFilter = teams
		}

		params := buildGlobalStatsParams(teamsFilter, r)

		response, err = s.StatsOpen(ctx, params)
		if err != nil {
			return nil, err
		}
		return Ok{response}, nil
	}
}

func makeGlobalStatsFixedEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		r, ok := request.(*GlobalStatsRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}

		var teamsFilter string
		// Only admin and observer users can set Tags query parameter.
		if r.Tags != "" {
			authorized, err := isAuthorizedTagsParam(ctx)
			if err != nil {
				return nil, err
			}
			if !authorized {
				return nil, errors.Forbidden("User is not allowed to set Tags parameter")
			}
			// The findings are stored by team so we must translate the tags
			// to team ids.
			teams, err := tagsToTeams(ctx, s, r.Tags)
			if err != nil {
				return nil, err
			}
			if len(teams) == 0 {
				return nil, errors.NotFound("There are no teams with the specified tags")
			}
			teamsFilter = teams
		}

		params := buildGlobalStatsParams(teamsFilter, r)

		response, err = s.StatsFixed(ctx, params)
		if err != nil {
			return nil, err
		}
		return Ok{response}, nil
	}
}

func makeGlobalStatsAssetsEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		r, ok := request.(*GlobalStatsRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}

		var teamsFilter string
		// Only admin and observer users can set Tags query parameter.
		if r.Tags != "" {
			authorized, err := isAuthorizedTagsParam(ctx)
			if err != nil {
				return nil, err
			}
			if !authorized {
				return nil, errors.Forbidden("User is not allowed to set Tags parameter")
			}
			// The findings are stored by team so we must translate the tags
			// to team ids.
			teams, err := tagsToTeams(ctx, s, r.Tags)
			if err != nil {
				return nil, err
			}
			if len(teams) == 0 {
				return nil, errors.NotFound("There are no teams with the specified tags")
			}
			teamsFilter = teams
		}

		params := buildGlobalStatsParams(teamsFilter, r)

		response, err = s.StatsAssets(ctx, params)
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

func isValidExposureRequest(r *StatsRequest) bool {
	return r.AtDate == "" || isValidDate(r.AtDate)
}

func buildStatsParams(r *StatsRequest) api.StatsParams {
	return api.StatsParams{
		Team:        r.TeamID,
		MinDate:     r.MinDate,
		MaxDate:     r.MaxDate,
		AtDate:      r.AtDate,
		MinScore:    r.MinScore,
		MaxScore:    r.MaxScore,
		Identifiers: r.Identifiers,
		Labels:      r.Labels,
	}
}

func buildGlobalStatsParams(teams string, r *GlobalStatsRequest) api.StatsParams {
	return api.StatsParams{
		Teams:       teams,
		MinDate:     r.MinDate,
		MaxDate:     r.MaxDate,
		AtDate:      r.AtDate,
		MinScore:    r.MinScore,
		MaxScore:    r.MaxScore,
		Identifiers: r.Identifiers,
		Labels:      r.Labels,
	}
}

func isValidDate(date string) bool {
	return regexp.MustCompile(dateFmtRegEx).MatchString(date)
}

// isAuthorizedTagsParam returns true if context user is authorized
// to set Tags query param. Only admin and observer roles are allowed.
func isAuthorizedTagsParam(ctx context.Context) (bool, error) {
	user, err := api.UserFromContext(ctx)
	if err != nil {
		return false, err
	}
	return (user.Admin != nil && *user.Admin) ||
		(user.Observer != nil && *user.Observer), nil
}

// tagsToTeams takes a command separated list of tags, looks for the teams that
// have any of those tags and returns a comma seperated list of those team
// id's.
func tagsToTeams(ctx context.Context, s api.VulcanitoService, tagsStr string) (string, error) {
	tags := strings.Split(tagsStr, ",")
	teams, err := s.FindTeamsByTags(ctx, tags)
	if err != nil {
		return "", err
	}
	ids := []string{}
	for _, t := range teams {
		ids = append(ids, t.ID)
	}
	return strings.Join(ids, ","), nil
}
