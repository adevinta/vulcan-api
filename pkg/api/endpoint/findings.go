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
	ID          string  `json:"id" urlvar:"finding_id"`
	TeamID      string  `json:"team_id" urlvar:"team_id"`
	Status      string  `urlquery:"status"`
	MinScore    float64 `urlquery:"minScore"`
	MaxScore    float64 `urlquery:"maxScore"`
	AtDate      string  `urlquery:"atDate"`
	MinDate     string  `urlquery:"minDate"`
	MaxDate     string  `urlquery:"maxDate"`
	SortBy      string  `urlquery:"sortBy"`
	Page        int     `urlquery:"page"`
	Size        int     `urlquery:"size"`
	Identifier  string  `urlquery:"identifier"`
	IssueID     string  `urlquery:"issueID"`
	TargetID    string  `urlquery:"targetID"`
	Identifiers string  `urlquery:"identifiers"`
	Labels      string  `urlquery:"labels"`
}

type FindingsByIssueRequest struct {
	TeamID      string  `json:"team_id" urlvar:"team_id"`
	Status      string  `urlquery:"status"`
	MinScore    float64 `urlquery:"minScore"`
	MaxScore    float64 `urlquery:"maxScore"`
	AtDate      string  `urlquery:"atDate"`
	MinDate     string  `urlquery:"minDate"`
	MaxDate     string  `urlquery:"maxDate"`
	SortBy      string  `urlquery:"sortBy"`
	Page        int     `urlquery:"page"`
	Size        int     `urlquery:"size"`
	IssueID     string  `json:"issue_id" urlvar:"issue_id"`
	Identifiers string  `urlquery:"identifiers"`
	Labels      string  `urlquery:"labels"`
}

type FindingsByTargetRequest struct {
	TeamID      string  `json:"team_id" urlvar:"team_id"`
	Status      string  `urlquery:"status"`
	MinScore    float64 `urlquery:"minScore"`
	MaxScore    float64 `urlquery:"maxScore"`
	AtDate      string  `urlquery:"atDate"`
	MinDate     string  `urlquery:"minDate"`
	MaxDate     string  `urlquery:"maxDate"`
	SortBy      string  `urlquery:"sortBy"`
	Page        int     `urlquery:"page"`
	Size        int     `urlquery:"size"`
	TargetID    string  `json:"target_id" urlvar:"target_id"`
	Identifiers string  `urlquery:"identifiers"`
	Labels      string  `urlquery:"labels"`
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

		params := buildFindingsParams(r)
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

		params := buildFindingsParams(r)
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

		params := buildFindingsByIssueParams(r)
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

		params := buildFindingsParams(r)
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

		params := buildFindingsByTargetParams(r)
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

		finding, err := s.FindFinding(ctx, r.ID)
		if err != nil {
			return nil, err
		}

		findingTicket, err := s.GetFindingTicket(ctx, finding.Finding.ID, r.TeamID)
		if err != nil {
			return nil, err
		}

		finding.Finding.TicketURL = findingTicket.Ticket.URLTracker

		if authorizedFindFindingRequest(finding.Finding.Target.Teams, r.TeamID) {
			return Ok{finding.Finding}, nil
		}

		return Forbidden{nil}, nil
	}
}

type FindingOverwriteRequest struct {
	FindingID string `json:"finding_id" urlvar:"finding_id"`
	TeamID    string `json:"team_id" urlvar:"team_id"`
	Status    string `json:"status"`
	Notes     string `json:"notes"`
}

func makeCreateFindingOverwriteEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		r, ok := request.(*FindingOverwriteRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}

		finding, err := s.FindFinding(ctx, r.FindingID)
		if err != nil {
			return nil, err
		}

		user, err := api.UserFromContext(ctx)
		if err != nil {
			return nil, errors.Default(err)
		}

		findingOverwrite := api.FindingOverwrite{
			UserID:         user.ID,
			FindingID:      r.FindingID,
			StatusPrevious: finding.Finding.Status,
			Status:         r.Status,
			Notes:          r.Notes,
			TeamID:         r.TeamID,
		}

		if authorizedFindFindingRequest(finding.Finding.Target.Teams, r.TeamID) {
			err := s.CreateFindingOverwrite(ctx, findingOverwrite)
			if err != nil {
				return nil, err
			}

			return Ok{}, nil
		}

		return Forbidden{}, nil
	}
}

func makeListFindingOverwritesEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		r, ok := request.(*FindingsRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}

		findingOverwrites, err := s.ListFindingOverwrites(ctx, r.ID)
		if err != nil {
			return nil, err
		}

		output := []api.FindingOverwriteResponse{}
		for _, fr := range findingOverwrites {
			output = append(output, fr.ToResponse())
		}

		return Ok{output}, nil
	}
}

func makeListFindingsLabelsEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		r, ok := request.(*FindingsRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}

		if !isValidListFindingsRequest(r) {
			return nil, errors.Validation("Invalid date format")
		}

		params := buildFindingsParams(r)

		response, err = s.ListFindingsLabels(ctx, params)
		if err != nil {
			return nil, err
		}
		return Ok{response}, nil
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

func buildFindingsParams(r *FindingsRequest) api.FindingsParams {
	return api.FindingsParams{
		Team:            r.TeamID,
		Status:          r.Status,
		MinScore:        r.MinScore,
		MaxScore:        r.MaxScore,
		AtDate:          r.AtDate,
		MinDate:         r.MinDate,
		MaxDate:         r.MaxDate,
		SortBy:          r.SortBy,
		Identifier:      r.Identifier,
		IdentifierMatch: true,
		IssueID:         r.IssueID,
		TargetID:        r.TargetID,
		Identifiers:     r.Identifiers,
		Labels:          r.Labels,
	}
}

func buildFindingsByIssueParams(r *FindingsByIssueRequest) api.FindingsParams {
	return api.FindingsParams{
		Team:        r.TeamID,
		Status:      r.Status,
		MinScore:    r.MinScore,
		MaxScore:    r.MaxScore,
		AtDate:      r.AtDate,
		MinDate:     r.MinDate,
		MaxDate:     r.MaxDate,
		SortBy:      r.SortBy,
		IssueID:     r.IssueID,
		Identifiers: r.Identifiers,
		Labels:      r.Labels,
	}
}

func buildFindingsByTargetParams(r *FindingsByTargetRequest) api.FindingsParams {
	return api.FindingsParams{
		Team:        r.TeamID,
		Status:      r.Status,
		MinScore:    r.MinScore,
		MaxScore:    r.MaxScore,
		AtDate:      r.AtDate,
		MinDate:     r.MinDate,
		MaxDate:     r.MaxDate,
		SortBy:      r.SortBy,
		TargetID:    r.TargetID,
		Identifiers: r.Identifiers,
		Labels:      r.Labels,
	}
}

func authorizedFindFindingRequest(allowedTeams []string, currentTeam string) bool {
	for _, v := range allowedTeams {
		if v == currentTeam {
			return true
		}
	}
	return false
}
