/*
Copyright 2021 Adevinta
*/

package api

import (
	vulndb "github.com/adevinta/vulnerability-db-api/pkg/model"
)

// FindingsParams represents the group of parameters
// that can be used to customize the call to retrieve
// the list of findings.
type FindingsParams struct {
	Tag             string
	Status          string
	MinScore        float64
	MaxScore        float64
	AtDate          string
	MinDate         string
	MaxDate         string
	SortBy          string
	IssueID         string
	TargetID        string
	Identifier      string
	IdentifierMatch bool
	Identifiers     string
	Labels          string
}

// FindingsList represents the response data returned
// from the vulnerability DB for a findings requests.
type FindingsList struct {
	Findings   []vulndb.FindingExpanded `json:"findings"`
	Pagination PaginationInfo           `json:"pagination"`
}

// FindingsIssuesList represents the response data returned
// from the vulnerability DB for the issues summary request.
type FindingsIssuesList struct {
	Issues     []vulndb.IssueSummary `json:"issues"`
	Pagination PaginationInfo        `json:"pagination"`
}

// FindingsTargetsList represents the response data returned
// from the vulnerability DB for the targets summary request.
type FindingsTargetsList struct {
	Targets    []vulndb.TargetSummary `json:"targets"`
	Pagination PaginationInfo         `json:"pagination"`
}

// Finding represents the response data returned from the vulnerability DB for
// the get finding request.
type Finding struct {
	Finding vulndb.FindingExpanded `json:"finding"`
}

// FindingsLabels represents the response data returned from the vulnerability DB
// for the list labels request.
type FindingsLabels struct {
	Labels []string `json:"labels"`
}

// UpdateFinding represents the payload submitted to update a finding.
type UpdateFinding struct {
	Status *string `json:"status"`
}

// CreateTarget specifies the payload for the vulnerability DB
// create target endpoint.
type CreateTarget struct {
	Identifier string   `json:"identifier"`
	Tags       []string `json:"tags"`
}

// TargetsParams represents the group of parameters
// that can be used to customize the call to retrieve
// the list of targets.
type TargetsParams struct {
	Tag             string
	Identifier      string
	IdentifierMatch bool
}

// Target represents the response data returned from the vulnerability DB
// for the create target request.
type Target struct {
	Target vulndb.Target `json:"target"`
}

// TargetsList represents the response data returned
// from the vulnerability DB for the targets list request.
type TargetsList struct {
	Targets    []vulndb.Target `json:"targets"`
	Pagination PaginationInfo  `json:"pagination"`
}

// StatsParams represents the group of parameters
// that can be used to customize the call to retrieve
// the statistics.
type StatsParams struct {
	Tag         string
	Tags        string
	MinDate     string
	MaxDate     string
	AtDate      string
	MinScore    float64
	MaxScore    float64
	Identifiers string
	Labels      string
}

// StatsMTTR represents the mean time to remediation stats by issue severity.
type StatsMTTR struct {
	MTTR vulndb.StatsMTTRSeverity `json:"mttr"`
}

// StatsExposure represents the exposure time stats by different averages.
type StatsExposure struct {
	Exposure vulndb.StatsExposure `json:"exposure"`
}

// StatsCurrentExposure represents the current exposure time stats by different averages.
type StatsCurrentExposure struct {
	Exposure vulndb.StatsExposure `json:"current_exposure"`
}

// StatsOpen represents the stats for open issues grouped by severity.
type StatsOpen struct {
	OpenIssues vulndb.StatsIssueSeverity `json:"open_issues"`
}

// StatsFixed represents the stats for fixed issues grouped by severity.
type StatsFixed struct {
	FixedIssues vulndb.StatsIssueSeverity `json:"fixed_issues"`
}

type StatsCoverage struct {
	Coverage float64 `json:"coverage"`
}
