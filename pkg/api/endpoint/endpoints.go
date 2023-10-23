/*
Copyright 2021 Adevinta
*/

package endpoint

import (
	"encoding/json"
	"net/http"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"

	"github.com/adevinta/vulcan-api/pkg/api"
)

const (
	// Endpoints
	Healthcheck = "Healthcheck"

	FindJob = "FindJob"

	ListUsers        = "ListUsers"
	CreateUser       = "CreateUser"
	UpdateUser       = "UpdateUser"
	FindUser         = "FindUser"
	DeleteUser       = "DeleteUser"
	FindProfile      = "FindProfile"
	GenerateAPIToken = "GenerateAPIToken"

	CreateTeam      = "CreateTeam"
	UpdateTeam      = "UpdateTeam"
	FindTeam        = "FindTeam"
	ListTeams       = "ListTeams"
	DeleteTeam      = "DeleteTeam"
	FindTeamsByUser = "FindTeamsByUser"

	ListTeamMembers  = "ListTeamMembers"
	FindTeamMember   = "FindTeamMember"
	CreateTeamMember = "CreateTeamMember"
	UpdateTeamMember = "UpdateTeamMember"
	DeleteTeamMember = "DeleteTeamMember"

	ListRecipients   = "ListRecipients"
	UpdateRecipients = "UpdateRecipients"

	ListAssets             = "ListAssets"
	CreateAsset            = "CreateAsset"
	CreateAssetMultiStatus = "CreateAssetMultiStatus"
	MergeDiscoveredAssets  = "MergeDiscoveredAssets"
	FindAsset              = "FindAsset"
	UpdateAsset            = "UpdateAsset"
	DeleteAsset            = "DeleteAsset"

	ListAssetAnnotations   = "ListAssetAnnotations"
	CreateAssetAnnotations = "CreateAssetAnnotations"
	UpdateAssetAnnotations = "UpdateAssetAnnotations"
	PutAssetAnnotations    = "PutAssetAnnotations"
	DeleteAssetAnnotations = "DeleteAssetAnnotations"

	CreateGroup = "CreateGroup"
	ListGroups  = "ListGroups"
	UpdateGroup = "UpdateGroup"
	DeleteGroup = "DeleteGroup"
	FindGroup   = "FindGroup"

	GroupAsset     = "GroupAsset"
	UngroupAsset   = "UngroupAsset"
	ListAssetGroup = "ListAssetGroup"

	ListPrograms  = "ListPrograms"
	CreateProgram = "CreateProgram"
	FindProgram   = "FindProgram"
	UpdateProgram = "UpdateProgram"
	DeleteProgram = "DeleteProgram"

	CreateSchedule        = "CreateSchedule"
	DeleteSchedule        = "DeleteSchedule"
	ScheduleGlobalProgram = "ScheduleGlobalProgram"

	ListPolicies = "ListPolicies"
	CreatePolicy = "CreatePolicy"
	FindPolicy   = "FindPolicy"
	UpdatePolicy = "UpdatePolicy"
	DeletePolicy = "DeletePolicy"

	ListChecktypeSetting   = "ListChecktypeSetting"
	CreateChecktypeSetting = "CreateChecktypeSetting"
	FindChecktypeSetting   = "FindChecktypeSetting"
	UpdateChecktypeSetting = "UpdateChecktypeSetting"
	DeleteChecktypeSetting = "DeleteChecktypeSetting"

	ListProgramScans = "ListProgramScans"
	CreateScan       = "CreateScan"
	FindScan         = "FindScan"
	AbortScan        = "AbortScan"

	SendDigestReport = "SendDigestReport"

	StatsCoverage              = "StatsCoverage"
	ListFindings               = "ListFindings"
	ListFindingsIssues         = "ListFindingsIssues"
	ListFindingsByIssue        = "ListFindingsByIssue"
	ListFindingsTargets        = "ListFindingsTargets"
	ListFindingsByTarget       = "ListFindingsByTarget"
	FindFinding                = "FindFinding"
	CreateFindingOverwrite     = "CreateFindingOverwrite"
	ListFindingOverwrites      = "ListFindingOverwrites"
	ListFindingsLabels         = "ListFindingsLabels"
	CreateFindingTicket        = "CreateFindingTicket"
	StatsMTTR                  = "StatsMTTR"
	StatsExposure              = "StatsExposure"
	StatsCurrentExposure       = "StatsCurrentExposure"
	StatsOpen                  = "StatsOpen"
	StatsFixed                 = "StatsFixed"
	GlobalStatsMTTR            = "GlobalStatsMTTR"
	GlobalStatsExposure        = "GlobalStatsExposure"
	GlobalStatsCurrentExposure = "GlobalStatsCurrentExposure"
	GlobalStatsOpen            = "GlobalStatsOpen"
	GlobalStatsFixed           = "GlobalStatsFixed"
	GlobalStatsAssets          = "GlobalStatsAssets"
)

// Endpoints contains all available endpoints for this api
type Endpoints map[string]endpoint.Endpoint

var endpoints = make(Endpoints)

// MakeEndpoints initialize endpoints using the given service
func MakeEndpoints(s api.VulcanitoService, isJiraIntEnabled bool, logger log.Logger) Endpoints {
	endpoints[Healthcheck] = makeHealthcheckEndpoint(s, logger)

	endpoints[FindJob] = makeFindJobEndpoint(s, logger)

	endpoints[ListUsers] = makeListUsersEndpoint(s, logger)
	endpoints[CreateUser] = makeCreateUserEndpoint(s, logger)
	endpoints[UpdateUser] = makeUpdateUserEndpoint(s, logger)
	endpoints[FindUser] = makeFindUserEndpoint(s, logger)
	endpoints[DeleteUser] = makeDeleteUserEndpoint(s, logger)
	endpoints[FindProfile] = makeFindProfileEndpoint(s, logger)
	endpoints[GenerateAPIToken] = makeGenerateAPITokenEndpoint(s, logger)

	endpoints[CreateTeam] = makeCreateTeamEndpoint(s, logger)
	endpoints[UpdateTeam] = makeUpdateTeamEndpoint(s, logger)
	endpoints[FindTeam] = makeFindTeamEndpoint(s, logger)
	endpoints[ListTeams] = makeListTeamsEndpoint(s, logger)
	endpoints[DeleteTeam] = makeDeleteTeamEndpoint(s, logger)
	endpoints[FindTeamsByUser] = makeFindTeamsByUserEndpoint(s, logger)

	endpoints[ListTeamMembers] = makeListTeamMembersEndpoint(s, logger)
	endpoints[FindTeamMember] = makeFindTeamMemberEndpoint(s, logger)
	endpoints[CreateTeamMember] = makeCreateTeamMemberEndpoint(s, logger)
	endpoints[UpdateTeamMember] = makeUpdateTeamMemberEndpoint(s, logger)
	endpoints[DeleteTeamMember] = makeDeleteTeamMemberEndpoint(s, logger)

	endpoints[ListRecipients] = makeListRecipientsEndpoint(s, logger)
	endpoints[UpdateRecipients] = makeUpdateRecipientsEndpoint(s, logger)

	endpoints[ListAssets] = makeListAssetsEndpoint(s, logger)
	endpoints[CreateAsset] = makeCreateAssetEndpoint(s, logger)
	endpoints[CreateAssetMultiStatus] = makeCreateAssetMultiStatusEndpoint(s, logger)
	endpoints[MergeDiscoveredAssets] = makeMergeDiscoveredAssetsEndpoint(s, logger)
	endpoints[FindAsset] = makeFindAssetEndpoint(s, logger)
	endpoints[UpdateAsset] = makeUpdateAssetEndpoint(s, logger)
	endpoints[DeleteAsset] = makeDeleteAssetEndpoint(s, logger)

	endpoints[ListAssetAnnotations] = makeListAssetAnnotationsEndpoint(s, logger)
	endpoints[CreateAssetAnnotations] = makeCreateAssetAnnotationsEndpoint(s, logger)
	endpoints[UpdateAssetAnnotations] = makeUpdateAssetAnnotationsEndpoint(s, logger)
	endpoints[PutAssetAnnotations] = makePutAssetAnnotationsEndpoint(s, logger)
	endpoints[DeleteAssetAnnotations] = makeDeleteAssetAnnotationsEndpoint(s, logger)

	endpoints[CreateGroup] = makeCreateGroupEndpoint(s, logger)
	endpoints[ListGroups] = makeListGroupsEndpoint(s, logger)
	endpoints[UpdateGroup] = makeUpdateGroupEndpoint(s, logger)
	endpoints[DeleteGroup] = makeDeleteGroupEndpoint(s, logger)
	endpoints[FindGroup] = makeFindGroupEndpoint(s, logger)

	endpoints[GroupAsset] = makeGroupAssetEndpoint(s, logger)
	endpoints[UngroupAsset] = makeUngroupAssetEndpoint(s, logger)
	endpoints[ListAssetGroup] = makeListAssetGroupEndpoint(s, logger)

	endpoints[ListPrograms] = makeListProgramsEndpoint(s, logger)
	endpoints[CreateProgram] = makeCreateProgramEndpoint(s, logger)
	endpoints[FindProgram] = makeFindProgramEndpoint(s, logger)
	endpoints[UpdateProgram] = makeUpdateProgramEndpoint(s, logger)
	endpoints[DeleteProgram] = makeDeleteProgramEndpoint(s, logger)

	endpoints[CreateSchedule] = makeCreateScheduleEndpoint(s, logger)
	endpoints[DeleteSchedule] = makeDeleteScheduleEndpoint(s, logger)
	endpoints[ScheduleGlobalProgram] = makeScheduleGlobalProgramEndpoint(s, logger)

	endpoints[ListPolicies] = makeListPoliciesEndpoint(s, logger)
	endpoints[CreatePolicy] = makeCreatePolicyEndpoint(s, logger)
	endpoints[FindPolicy] = makeFindPolicyEndpoint(s, logger)
	endpoints[UpdatePolicy] = makeUpdatePolicyEndpoint(s, logger)
	endpoints[DeletePolicy] = makeDeletePolicyEndpoint(s, logger)

	endpoints[ListChecktypeSetting] = makeListChecktypeSettingEndpoint(s, logger)
	endpoints[CreateChecktypeSetting] = makeCreateChecktypeSettingEndpoint(s, logger)
	endpoints[FindChecktypeSetting] = makeFindChecktypeSettingEndpoint(s, logger)
	endpoints[UpdateChecktypeSetting] = makeUpdateChecktypeSettingEndpoint(s, logger)
	endpoints[DeleteChecktypeSetting] = makeDeleteChecktypeSettingEndpoint(s, logger)

	endpoints[ListProgramScans] = makeListProgramScansEndpoint(s, logger)
	endpoints[CreateScan] = makeCreateScanEndpoint(s, logger)
	endpoints[FindScan] = makeFindScanEndpoint(s, logger)
	endpoints[AbortScan] = makeAbortScanEndpoint(s, logger)

	endpoints[SendDigestReport] = makeSendDigestReportEndpoint(s, logger)

	endpoints[StatsCoverage] = makeStatsCoverageEndpoint(s, logger)
	endpoints[ListFindings] = makeListFindingsEndpoint(s, logger)
	endpoints[ListFindingsIssues] = makeListFindingsIssuesEndpoint(s, logger)
	endpoints[ListFindingsByIssue] = makeListFindingsByIssueEndpoint(s, logger)
	endpoints[ListFindingsTargets] = makeListFindingsTargetsEndpoint(s, logger)
	endpoints[ListFindingsByTarget] = makeListFindingsByTargetEndpoint(s, logger)
	endpoints[FindFinding] = makeFindFindingEndpoint(s, logger)
	endpoints[CreateFindingOverwrite] = makeCreateFindingOverwriteEndpoint(s, logger)
	endpoints[ListFindingOverwrites] = makeListFindingOverwritesEndpoint(s, logger)
	if isJiraIntEnabled {
		endpoints[CreateFindingTicket] = makeCreateFindingTicketEndpoint(s, logger)
	}
	endpoints[ListFindingsLabels] = makeListFindingsLabelsEndpoint(s, logger)
	endpoints[StatsMTTR] = makeStatsMTTREndpoint(s, logger)
	endpoints[StatsExposure] = makeStatsExposureEndpoint(s, logger)
	endpoints[StatsCurrentExposure] = makeStatsCurrentExposureEndpoint(s, logger)
	endpoints[StatsOpen] = makeStatsOpenEndpoint(s, logger)
	endpoints[StatsFixed] = makeStatsFixedEndpoint(s, logger)
	endpoints[GlobalStatsMTTR] = makeGlobalStatsMTTREndpoint(s, logger)
	endpoints[GlobalStatsExposure] = makeGlobalStatsExposureEndpoint(s, logger)
	endpoints[GlobalStatsCurrentExposure] = makeGlobalStatsCurrentExposureEndpoint(s, logger)
	endpoints[GlobalStatsOpen] = makeGlobalStatsOpenEndpoint(s, logger)
	endpoints[GlobalStatsFixed] = makeGlobalStatsFixedEndpoint(s, logger)
	endpoints[GlobalStatsAssets] = makeGlobalStatsAssetsEndpoint(s, logger)

	return endpoints
}

type EmptyRequest struct{}

type HTTPResponse interface {
	StatusCode() int
}

type Created struct {
	Data interface{}
}

func (c Created) StatusCode() int {
	return http.StatusCreated
}

func (c Created) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.Data)
}

type Ok struct {
	Data interface{}
}

func (c Ok) StatusCode() int {
	return http.StatusOK
}

func (c Ok) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.Data)
}

type Accepted struct {
	Data interface{}
}

func (c Accepted) StatusCode() int {
	return http.StatusAccepted
}

func (c Accepted) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.Data)
}

type ServerDown struct {
	Data interface{}
}

func (c ServerDown) StatusCode() int {
	return http.StatusInternalServerError
}

func (c ServerDown) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.Data)
}

type NoContent struct {
	Data interface{}
}

func (c NoContent) StatusCode() int {
	return http.StatusNoContent
}

func (c NoContent) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.Data)
}

type MultiStatus struct {
	Data interface{}
}

func (makeCreateAssetMultiStatusEndpoint MultiStatus) StatusCode() int {
	return http.StatusMultiStatus
}

func (m MultiStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.Data)
}

type Forbidden struct {
	Data interface{}
}

func (c Forbidden) StatusCode() int {
	return http.StatusForbidden
}

func (c Forbidden) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.Data)
}
