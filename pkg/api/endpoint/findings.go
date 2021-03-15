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

type FindingsRequest struct {
	ID         string  `json:"id" urlvar:"finding_id"`
	TeamID     string  `json:"team_id" urlvar:"team_id"`
	Status     string  `urlquery:"status"`
	MinScore   float64 `urlquery:"minScore"`
	MaxScore   float64 `urlquery:"maxScore"`
	AtDate     string  `urlquery:"atDate"`
	MinDate    string  `urlquery:"minDate"`
	MaxDate    string  `urlquery:"maxDate"`
	SortBy     string  `urlquery:"sortBy"`
	Page       int     `urlquery:"page"`
	Size       int     `urlquery:"size"`
	Identifier string  `urlquery:"identifier"`
}

type FindingsByIssueRequest struct {
	TeamID   string  `json:"team_id" urlvar:"team_id"`
	Status   string  `urlquery:"status"`
	MinScore float64 `urlquery:"minScore"`
	MaxScore float64 `urlquery:"maxScore"`
	AtDate   string  `urlquery:"atDate"`
	MinDate  string  `urlquery:"minDate"`
	MaxDate  string  `urlquery:"maxDate"`
	SortBy   string  `urlquery:"sortBy"`
	Page     int     `urlquery:"page"`
	Size     int     `urlquery:"size"`
	IssueID  string  `json:"issue_id" urlvar:"issue_id"`
}

type FindingsByTargetRequest struct {
	TeamID   string  `json:"team_id" urlvar:"team_id"`
	Status   string  `urlquery:"status"`
	MinScore float64 `urlquery:"minScore"`
	MaxScore float64 `urlquery:"maxScore"`
	AtDate   string  `urlquery:"atDate"`
	MinDate  string  `urlquery:"minDate"`
	MaxDate  string  `urlquery:"maxDate"`
	SortBy   string  `urlquery:"sortBy"`
	Page     int     `urlquery:"page"`
	Size     int     `urlquery:"size"`
	TargetID string  `json:"target_id" urlvar:"target_id"`
}

func makeListFindingsEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		r, ok := request.(*FindingsRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}

		if !isValidListFindingsRequest(r) {
			return nil, errors.Validation("Invalid date format")
		}

		team, err := s.FindTeam(ctx, r.TeamID)
		if err != nil {
			return nil, err
		}
		if team.Tag == "" {
			return nil, errors.Validation("no tag defined for the team")
		}

		params := buildFindingsParams(team.Tag, r)
		pagination := api.Pagination{Page: r.Page, Size: r.Size}

		response, err = s.ListFindings(ctx, params, pagination)
		if err != nil {
			return nil, err
		}
		return Ok{response}, nil
	}
}

func makeListFindingsIssuesEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		r, ok := request.(*FindingsRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}

		if !isValidFindingsSummaryRequest(r) {
			return nil, errors.Validation("Invalid request parameters")
		}

		team, err := s.FindTeam(ctx, r.TeamID)
		if err != nil {
			return nil, err
		}
		if team.Tag == "" {
			return nil, errors.Validation("no tag defined for the team")
		}

		params := buildFindingsParams(team.Tag, r)
		pagination := api.Pagination{Page: r.Page, Size: r.Size}

		response, err = s.ListFindingsIssues(ctx, params, pagination)
		if err != nil {
			return nil, err
		}
		return Ok{response}, nil
	}
}

func makeListFindingsByIssueEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		r, ok := request.(*FindingsByIssueRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}

		team, err := s.FindTeam(ctx, r.TeamID)
		if err != nil {
			return nil, err
		}
		if team.Tag == "" {
			return nil, errors.Validation("no tag defined for the team")
		}

		params := buildFindingsByIssueParams(team.Tag, r)
		pagination := api.Pagination{Page: r.Page, Size: r.Size}

		response, err = s.ListFindingsByIssue(ctx, params, pagination)
		if err != nil {
			return nil, err
		}
		return Ok{response}, nil
	}
}

func makeListFindingsTargetsEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		r, ok := request.(*FindingsRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}

		if !isValidFindingsSummaryRequest(r) {
			return nil, errors.Validation("Invalid request parameters")
		}

		team, err := s.FindTeam(ctx, r.TeamID)
		if err != nil {
			return nil, err
		}
		if team.Tag == "" {
			return nil, errors.Validation("no tag defined for the team")
		}

		params := buildFindingsParams(team.Tag, r)
		pagination := api.Pagination{Page: r.Page, Size: r.Size}

		response, err = s.ListFindingsTargets(ctx, params, pagination)
		if err != nil {
			return nil, err
		}
		return Ok{response}, nil
	}
}

func makeListFindingsByTargetEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		r, ok := request.(*FindingsByTargetRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}

		team, err := s.FindTeam(ctx, r.TeamID)
		if err != nil {
			return nil, err
		}
		if team.Tag == "" {
			return nil, errors.Validation("no tag defined for the team")
		}

		params := buildFindingsByTargetParams(team.Tag, r)
		pagination := api.Pagination{Page: r.Page, Size: r.Size}

		response, err = s.ListFindingsByTarget(ctx, params, pagination)
		if err != nil {
			return nil, err
		}
		return Ok{response}, nil
	}
}

func makeFindFindingEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		r, ok := request.(*FindingsRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}
		team, err := s.FindTeam(ctx, r.TeamID)
		if err != nil {
			return nil, err
		}
		if team.Tag == "" {
			return nil, errors.Validation("no tag defined for the team")
		}

		finding, err := s.FindFinding(ctx, r.ID)
		if err != nil {
			return nil, err
		}

		if authorizedFindFindingRequest(finding.Finding.Target.Tags, team.Tag) {
			return Ok{finding}, nil
		}

		return Forbidden{nil}, nil
	}
}

func isValidListFindingsRequest(r *FindingsRequest) bool {
	return (r.MinDate == "" || isValidDate(r.MinDate)) &&
		(r.MaxDate == "" || isValidDate(r.MaxDate)) &&
		(r.AtDate == "" || isValidDate(r.AtDate))
}

func isValidFindingsSummaryRequest(r *FindingsRequest) bool {
	return r.MinScore == 0 && r.MaxScore == 0 &&
		(r.MinDate == "" || isValidDate(r.MinDate)) &&
		(r.MaxDate == "" || isValidDate(r.MaxDate)) &&
		(r.AtDate == "" || isValidDate(r.AtDate))
}

func buildFindingsParams(tag string, r *FindingsRequest) api.FindingsParams {
	return api.FindingsParams{
		Tag:             tag,
		Status:          r.Status,
		MinScore:        r.MinScore,
		MaxScore:        r.MaxScore,
		AtDate:          r.AtDate,
		MinDate:         r.MinDate,
		MaxDate:         r.MaxDate,
		SortBy:          r.SortBy,
		Identifier:      r.Identifier,
		IdentifierMatch: true,
	}
}

func buildFindingsByIssueParams(tag string, r *FindingsByIssueRequest) api.FindingsParams {
	return api.FindingsParams{
		Tag:      tag,
		Status:   r.Status,
		MinScore: r.MinScore,
		MaxScore: r.MaxScore,
		AtDate:   r.AtDate,
		MinDate:  r.MinDate,
		MaxDate:  r.MaxDate,
		SortBy:   r.SortBy,
		IssueID:  r.IssueID,
	}
}

func buildFindingsByTargetParams(tag string, r *FindingsByTargetRequest) api.FindingsParams {
	return api.FindingsParams{
		Tag:      tag,
		Status:   r.Status,
		MinScore: r.MinScore,
		MaxScore: r.MaxScore,
		AtDate:   r.AtDate,
		MinDate:  r.MinDate,
		MaxDate:  r.MaxDate,
		SortBy:   r.SortBy,
		TargetID: r.TargetID,
	}
}

func authorizedFindFindingRequest(allowedTags []string, currentTag string) bool {
	for _, v := range allowedTags {
		if v == currentTag {
			return true
		}
	}
	return false
}
