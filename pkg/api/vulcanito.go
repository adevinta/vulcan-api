/*
Copyright 2021 Adevinta
*/

package api

import (
	"context"
)

// VulcanitoService represents all operations provided by Vulcanito
type VulcanitoService interface {
	// Healthcheck
	Healthcheck(ctx context.Context) error

	// Jobs
	FindJob(ctx context.Context, jobID string) (*Job, error)
	UpdateJob(ctx context.Context, job Job) (*Job, error)

	// Users
	ListUsers(ctx context.Context) ([]*User, error)
	CreateUser(ctx context.Context, user User) (*User, error)
	UpdateUser(ctx context.Context, user User) (*User, error)
	FindUser(ctx context.Context, userID string) (*User, error)
	DeleteUser(ctx context.Context, userID string) error

	GenerateAPIToken(ctx context.Context, userID string) (*Token, error)

	// Teams
	CreateTeam(ctx context.Context, team Team, ownerEmail string) (*Team, error)
	UpdateTeam(ctx context.Context, team Team) (*Team, error)
	FindTeam(ctx context.Context, teamID string) (*Team, error)
	FindTeamByTag(ctx context.Context, tag string) (*Team, error)
	DeleteTeam(ctx context.Context, teamID string) error
	ListTeams(ctx context.Context) ([]*Team, error)
	FindTeamsByUser(ctx context.Context, userID string) ([]*Team, error)
	FindTeamsByTags(ctx context.Context, tags []string) ([]*Team, error)

	// TeamMembers
	FindTeamMember(ctx context.Context, teamID string, userID string) (*UserTeam, error)
	CreateTeamMember(ctx context.Context, teamUser UserTeam) (*UserTeam, error)
	UpdateTeamMember(ctx context.Context, teamUser UserTeam) (*UserTeam, error)
	DeleteTeamMember(ctx context.Context, teamID string, userID string) error

	// Recipients
	UpdateRecipients(ctx context.Context, teamID string, emails []string) error
	ListRecipients(ctx context.Context, teamID string) ([]*Recipient, error)

	// Assets
	ListAssets(ctx context.Context, teamID string, asset Asset) ([]*Asset, error)
	CreateAssets(ctx context.Context, assets []Asset, groups []Group, annotations []*AssetAnnotation) ([]Asset, error)
	CreateAssetsMultiStatus(ctx context.Context, assets []Asset, groups []Group, annotations []*AssetAnnotation) ([]AssetCreationResponse, error)
	MergeDiscoveredAssets(ctx context.Context, teamID string, assets []Asset, groupName string) error
	MergeDiscoveredAssetsAsync(ctx context.Context, teamID string, assets []Asset, groupName string) (*Job, error)
	FindAsset(ctx context.Context, asset Asset) (*Asset, error)
	UpdateAsset(ctx context.Context, asset Asset) (*Asset, error)
	DeleteAsset(ctx context.Context, asset Asset) error
	DeleteAllAssets(ctx context.Context, teamID string) error
	GetAssetType(ctx context.Context, assetTypeName string) (*AssetType, error)

	// Asset Annotations
	ListAssetAnnotations(ctx context.Context, teamID string, assetID string) ([]*AssetAnnotation, error)
	CreateAssetAnnotations(ctx context.Context, teamID string, assetID string, annotations []*AssetAnnotation) ([]*AssetAnnotation, error)
	UpdateAssetAnnotations(ctx context.Context, teamID string, assetID string, annotations []*AssetAnnotation) ([]*AssetAnnotation, error)
	PutAssetAnnotations(ctx context.Context, teamID string, assetID string, annotations []*AssetAnnotation) ([]*AssetAnnotation, error)
	DeleteAssetAnnotations(ctx context.Context, teamID string, assedID string, annotations []*AssetAnnotation) error

	ListGroups(ctx context.Context, teamID, groupName string) ([]*Group, error)
	CreateGroup(ctx context.Context, group Group) (*Group, error)
	FindGroup(ctx context.Context, group Group) (*Group, error)
	UpdateGroup(ctx context.Context, group Group) (*Group, error)
	DeleteGroup(ctx context.Context, group Group) error

	GroupAsset(ctx context.Context, assetGroup AssetGroup, teamID string) (*AssetGroup, error)
	UngroupAsset(ctx context.Context, assetGroup AssetGroup, teamID string) error
	ListAssetGroup(ctx context.Context, assetGroup AssetGroup, teamID string) ([]*Asset, error)

	ListPrograms(ctx context.Context, teamID string) ([]*Program, error)
	CreateProgram(ctx context.Context, program Program, teamID string) (*Program, error)
	FindProgram(ctx context.Context, programID string, teamID string) (*Program, error)
	UpdateProgram(ctx context.Context, program Program, teamID string) (*Program, error)
	DeleteProgram(ctx context.Context, program Program, teamID string) error

	// Schedules
	CreateSchedule(ctx context.Context, programID string, cronExpr string, teamID string) (*Program, error)
	DeleteSchedule(ctx context.Context, programID string, teamID string) (*Program, error)
	ScheduleGlobalProgram(ctx context.Context, programID string, cronExpr string) error

	ListPolicies(ctx context.Context, teamID string) ([]*Policy, error)
	CreatePolicy(ctx context.Context, policy Policy) (*Policy, error)
	FindPolicy(ctx context.Context, policyID string) (*Policy, error)
	UpdatePolicy(ctx context.Context, policy Policy) (*Policy, error)
	DeletePolicy(ctx context.Context, policy Policy) error

	ListChecktypeSetting(ctx context.Context, policyID string) ([]*ChecktypeSetting, error)
	CreateChecktypeSetting(ctx context.Context, setting ChecktypeSetting) (*ChecktypeSetting, error)
	FindChecktypeSetting(ctx context.Context, policyID, checktypeSettingID string) (*ChecktypeSetting, error)
	UpdateChecktypeSetting(ctx context.Context, checktypeSetting ChecktypeSetting) (*ChecktypeSetting, error)
	DeleteChecktypeSetting(ctx context.Context, checktypeSettingID string) error

	ListScans(ctx context.Context, teamID string, programID string) ([]*Scan, error)
	CreateScan(ctx context.Context, scan Scan, teamID string) (*Scan, error)
	FindScan(ctx context.Context, scanID, teamID string) (*Scan, error)
	AbortScan(ctx context.Context, scanID string, teamID string) (*Scan, error)
	UpdateScan(ctx context.Context, scan Scan) (*Scan, error)
	DeleteScan(ctx context.Context, scan Scan) error

	SendDigestReport(ctx context.Context, teamID string, startDate string, endDate string) error

	// Stats
	StatsCoverage(ctx context.Context, teamID string) (*StatsCoverage, error)

	// VulnerabilityDB Stats
	ListFindings(ctx context.Context, params FindingsParams, pagination Pagination) (*FindingsList, error)
	ListFindingsIssues(ctx context.Context, params FindingsParams, pagination Pagination) (*FindingsIssuesList, error)
	ListFindingsByIssue(ctx context.Context, params FindingsParams, pagination Pagination) (*FindingsList, error)
	ListFindingsTargets(ctx context.Context, params FindingsParams, pagination Pagination) (*FindingsTargetsList, error)
	ListFindingsByTarget(ctx context.Context, params FindingsParams, pagination Pagination) (*FindingsList, error)
	ListFindingsLabels(ctx context.Context, params FindingsParams) (*FindingsLabels, error)
	FindFinding(ctx context.Context, findingID string) (*Finding, error)
	CreateFindingOverwrite(ctx context.Context, findingOverwrite FindingOverwrite) error
	ListFindingOverwrites(ctx context.Context, findingID string) ([]*FindingOverwrite, error)
	StatsMTTR(ctx context.Context, params StatsParams) (*StatsMTTR, error)
	StatsExposure(ctx context.Context, params StatsParams) (*StatsExposure, error)
	StatsCurrentExposure(ctx context.Context, params StatsParams) (*StatsCurrentExposure, error)
	StatsOpen(ctx context.Context, params StatsParams) (*StatsOpen, error)
	StatsFixed(ctx context.Context, params StatsParams) (*StatsFixed, error)
	StatsAssets(ctx context.Context, params StatsParams) (*StatsAssets, error)

	// Vulcan Tracker
	CreateFindingTicket(ctx context.Context, ticket FindingTicketCreate) (*Ticket, error)
	GetFindingTicket(ctx context.Context, findingID, teamID string) (*Ticket, error)
	IsATeamOnboardedInVulcanTracker(ctx context.Context, teamID string, onboardedTeams []string) bool
}
