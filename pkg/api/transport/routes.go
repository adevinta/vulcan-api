/*
Copyright 2021 Adevinta
*/

package transport

import (
	"net/http"

	kitlog "github.com/go-kit/kit/log"
	"github.com/gorilla/mux"

	"github.com/adevinta/vulcan-api/pkg/api/endpoint"
)

// AttachRoutes wire handlers with routes
func AttachRoutes(e endpoint.Endpoints, logger kitlog.Logger) http.Handler {
	r := mux.NewRouter()

	r.Methods("GET").Path("/api/v1").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/api/v1/login?redirect_to=/api/v1/home.html", http.StatusFound)
	})

	// Healthcheck
	r.Methods("GET").Path("/api/v1/healthcheck").Handler(newServer(e[endpoint.Healthcheck], endpoint.HealthcheckJSONRequest{}, logger, endpoint.Healthcheck))

	// Users
	r.Methods("GET").Path("/api/v1/users").Handler(newServer(e[endpoint.ListUsers], endpoint.EmptyRequest{}, logger, endpoint.ListUsers))
	r.Methods("POST").Path("/api/v1/users").Handler(newServer(e[endpoint.CreateUser], endpoint.UserRequest{}, logger, endpoint.CreateUser))
	r.Methods("GET").Path("/api/v1/users/{user_id}").Handler(newServer(e[endpoint.FindUser], endpoint.UserRequest{}, logger, endpoint.FindUser))
	r.Methods("PATCH").Path("/api/v1/users/{user_id}").Handler(newServer(e[endpoint.UpdateUser], endpoint.UserRequest{}, logger, endpoint.UpdateUser))
	r.Methods("DELETE").Path("/api/v1/users/{user_id}").Handler(newServer(e[endpoint.DeleteUser], endpoint.UserRequest{}, logger, endpoint.DeleteUser))

	// Profile
	r.Methods("GET").Path("/api/v1/profile").Handler(newServer(e[endpoint.FindProfile], endpoint.EmptyRequest{}, logger, endpoint.FindProfile))

	// Token
	r.Methods("POST").Path("/api/v1/users/{user_id}/token").Handler(newServer(e[endpoint.GenerateAPIToken], endpoint.UserRequest{}, logger, endpoint.GenerateAPIToken))

	// Teams from a User
	r.Methods("GET").Path("/api/v1/users/{user_id}/teams").Handler(newServer(e[endpoint.FindTeamsByUser], endpoint.UserRequest{}, logger, endpoint.FindTeamsByUser))

	// Teams
	r.Methods("GET").Path("/api/v1/teams").Handler(newServer(e[endpoint.ListTeams], endpoint.TeamRequest{}, logger, endpoint.ListTeams))
	r.Methods("POST").Path("/api/v1/teams").Handler(newServer(e[endpoint.CreateTeam], endpoint.TeamRequest{}, logger, endpoint.CreateTeam))
	r.Methods("GET").Path("/api/v1/teams/{team_id}").Handler(newServer(e[endpoint.FindTeam], endpoint.TeamRequest{}, logger, endpoint.FindTeam))
	r.Methods("PATCH").Path("/api/v1/teams/{team_id}").Handler(newServer(e[endpoint.UpdateTeam], endpoint.TeamUpdateRequest{}, logger, endpoint.UpdateTeam))
	r.Methods("DELETE").Path("/api/v1/teams/{team_id}").Handler(newServer(e[endpoint.DeleteTeam], endpoint.TeamRequest{}, logger, endpoint.DeleteTeam))

	// Team members
	r.Methods("POST").Path("/api/v1/teams/{team_id}/members").Handler(newServer(e[endpoint.CreateTeamMember], endpoint.TeamMemberRequest{}, logger, endpoint.CreateTeamMember))
	r.Methods("GET").Path("/api/v1/teams/{team_id}/members").Handler(newServer(e[endpoint.ListTeamMembers], endpoint.TeamMemberRequest{}, logger, endpoint.ListTeamMembers))
	r.Methods("GET").Path("/api/v1/teams/{team_id}/members/{user_id}").Handler(newServer(e[endpoint.FindTeamMember], endpoint.TeamMemberRequest{}, logger, endpoint.FindTeamMember))
	r.Methods("PATCH").Path("/api/v1/teams/{team_id}/members/{user_id}").Handler(newServer(e[endpoint.UpdateTeamMember], endpoint.TeamMemberRequest{}, logger, endpoint.UpdateTeamMember))
	r.Methods("DELETE").Path("/api/v1/teams/{team_id}/members/{user_id}").Handler(newServer(e[endpoint.DeleteTeamMember], endpoint.TeamMemberRequest{}, logger, endpoint.DeleteTeamMember))

	// Team recipients
	r.Methods("GET").Path("/api/v1/teams/{team_id}/recipients").Handler(newServer(e[endpoint.ListRecipients], endpoint.RecipientsData{}, logger, endpoint.ListRecipients))
	r.Methods("PUT").Path("/api/v1/teams/{team_id}/recipients").Handler(newServer(e[endpoint.UpdateRecipients], endpoint.RecipientsData{}, logger, endpoint.UpdateRecipients))

	// Assets
	r.Methods("GET").Path("/api/v1/teams/{team_id}/assets").Handler(newServer(e[endpoint.ListAssets], endpoint.AssetRequest{}, logger, endpoint.ListAssets))
	r.Methods("POST").Path("/api/v1/teams/{team_id}/assets").Handler(newServer(e[endpoint.CreateAsset], endpoint.AssetsListRequest{}, logger, endpoint.CreateAsset))
	r.Methods("POST").Path("/api/v1/teams/{team_id}/assets/multistatus").Handler(newServer(e[endpoint.CreateAssetMultiStatus], endpoint.AssetsListRequest{}, logger, endpoint.CreateAssetMultiStatus))

	r.Methods("GET").Path("/api/v1/teams/{team_id}/assets/{asset_id}").Handler(newServer(e[endpoint.FindAsset], endpoint.AssetRequest{}, logger, endpoint.FindAsset))
	r.Methods("PATCH").Path("/api/v1/teams/{team_id}/assets/{asset_id}").Handler(newServer(e[endpoint.UpdateAsset], endpoint.AssetRequest{}, logger, endpoint.UpdateAsset))
	r.Methods("DELETE").Path("/api/v1/teams/{team_id}/assets/{asset_id}").Handler(newServer(e[endpoint.DeleteAsset], endpoint.AssetRequest{}, logger, endpoint.DeleteAsset))

	// Groups
	r.Methods("POST").Path("/api/v1/teams/{team_id}/groups").Handler(newServer(e[endpoint.CreateGroup], endpoint.AssetsGroupRequest{}, logger, endpoint.CreateGroup))
	r.Methods("GET").Path("/api/v1/teams/{team_id}/groups").Handler(newServer(e[endpoint.ListGroups], endpoint.ListGroupsRequest{}, logger, endpoint.ListGroups))
	r.Methods("PATCH").Path("/api/v1/teams/{team_id}/groups/{group_id}").Handler(newServer(e[endpoint.UpdateGroup], endpoint.AssetsGroupRequest{}, logger, endpoint.UpdateGroup))
	r.Methods("DELETE").Path("/api/v1/teams/{team_id}/groups/{group_id}").Handler(newServer(e[endpoint.DeleteGroup], endpoint.AssetsGroupRequest{}, logger, endpoint.DeleteGroup))
	r.Methods("GET").Path("/api/v1/teams/{team_id}/groups/{group_id}").Handler(newServer(e[endpoint.FindGroup], endpoint.AssetsGroupRequest{}, logger, endpoint.FindGroup))

	// Group-assets association
	r.Methods("GET").Path("/api/v1/teams/{team_id}/groups/{group_id}/assets").Handler(newServer(e[endpoint.ListAssetGroup], endpoint.GroupAssetRequest{}, logger, endpoint.ListAssetGroup))
	r.Methods("POST").Path("/api/v1/teams/{team_id}/groups/{group_id}/assets").Handler(newServer(e[endpoint.GroupAsset], endpoint.GroupAssetRequest{}, logger, endpoint.GroupAsset))
	r.Methods("DELETE").Path("/api/v1/teams/{team_id}/groups/{group_id}/assets/{asset_id}").Handler(newServer(e[endpoint.UngroupAsset], endpoint.GroupAssetRequest{}, logger, endpoint.UngroupAsset))

	// Programs
	r.Methods("GET").Path("/api/v1/teams/{team_id}/programs").Handler(newServer(e[endpoint.ListPrograms], endpoint.ProgramRequest{}, logger, endpoint.ListPrograms))
	r.Methods("POST").Path("/api/v1/teams/{team_id}/programs").Handler(newServer(e[endpoint.CreateProgram], endpoint.ProgramRequest{}, logger, endpoint.CreateProgram))
	r.Methods("GET").Path("/api/v1/teams/{team_id}/programs/{program_id}").Handler(newServer(e[endpoint.FindProgram], endpoint.ProgramRequest{}, logger, endpoint.FindProgram))
	r.Methods("PATCH").Path("/api/v1/teams/{team_id}/programs/{program_id}").Handler(newServer(e[endpoint.UpdateProgram], endpoint.ProgramRequest{}, logger, endpoint.UpdateProgram))
	r.Methods("DELETE").Path("/api/v1/teams/{team_id}/programs/{program_id}").Handler(newServer(e[endpoint.DeleteProgram], endpoint.ProgramRequest{}, logger, endpoint.DeleteProgram))
	r.Methods("GET").Path("/api/v1/teams/{team_id}/programs/{program_id}/scans").Handler(newServer(e[endpoint.ListProgramScans], endpoint.ListProgramScansRequest{}, logger, endpoint.ListProgramScans))

	// Schedules
	r.Methods("POST").Path("/api/v1/teams/{team_id}/programs/{program_id}/schedule").Handler(newServer(e[endpoint.CreateSchedule], endpoint.ScheduleRequest{}, logger, endpoint.CreateSchedule))
	r.Methods("DELETE").Path("/api/v1/teams/{team_id}/programs/{program_id}/schedule").Handler(newServer(e[endpoint.DeleteSchedule], endpoint.ScheduleRequest{}, logger, endpoint.DeleteSchedule))
	r.Methods("PUT").Path("/api/v1/programs/{program_id}/schedule").Handler(newServer(e[endpoint.ScheduleGlobalProgram], endpoint.ScheduleGlobalRequest{}, logger, endpoint.ScheduleGlobalProgram))

	// Policies
	r.Methods("GET").Path("/api/v1/teams/{team_id}/policies").Handler(newServer(e[endpoint.ListPolicies], endpoint.PolicyRequest{}, logger, endpoint.ListPolicies))
	r.Methods("POST").Path("/api/v1/teams/{team_id}/policies").Handler(newServer(e[endpoint.CreatePolicy], endpoint.PolicyRequest{}, logger, endpoint.CreatePolicy))
	r.Methods("GET").Path("/api/v1/teams/{team_id}/policies/{policy_id}").Handler(newServer(e[endpoint.FindPolicy], endpoint.PolicyRequest{}, logger, endpoint.FindPolicy))
	r.Methods("PATCH").Path("/api/v1/teams/{team_id}/policies/{policy_id}").Handler(newServer(e[endpoint.UpdatePolicy], endpoint.PolicyRequest{}, logger, endpoint.UpdatePolicy))
	r.Methods("DELETE").Path("/api/v1/teams/{team_id}/policies/{policy_id}").Handler(newServer(e[endpoint.DeletePolicy], endpoint.PolicyRequest{}, logger, endpoint.DeletePolicy))

	// Policiy x CheckType Settings
	r.Methods("GET").Path("/api/v1/teams/{team_id}/policies/{policy_id}/settings").Handler(newServer(e[endpoint.ListChecktypeSetting], endpoint.ChecktypeSettingRequest{}, logger, endpoint.ListChecktypeSetting))
	r.Methods("POST").Path("/api/v1/teams/{team_id}/policies/{policy_id}/settings").Handler(newServer(e[endpoint.CreateChecktypeSetting], endpoint.ChecktypeSettingRequest{}, logger, endpoint.CreateChecktypeSetting))
	r.Methods("GET").Path("/api/v1/teams/{team_id}/policies/{policy_id}/settings/{setting_id}").Handler(newServer(e[endpoint.FindChecktypeSetting], endpoint.ChecktypeSettingRequest{}, logger, endpoint.FindChecktypeSetting))
	r.Methods("PATCH").Path("/api/v1/teams/{team_id}/policies/{policy_id}/settings/{setting_id}").Handler(newServer(e[endpoint.UpdateChecktypeSetting], endpoint.ChecktypeSettingRequest{}, logger, endpoint.UpdateChecktypeSetting))
	r.Methods("DELETE").Path("/api/v1/teams/{team_id}/policies/{policy_id}/settings/{setting_id}").Handler(newServer(e[endpoint.DeleteChecktypeSetting], endpoint.ChecktypeSettingRequest{}, logger, endpoint.DeleteChecktypeSetting))

	// scans
	r.Methods("POST").Path("/api/v1/teams/{team_id}/scans").Handler(newServer(e[endpoint.CreateScan], endpoint.ScanRequest{}, logger, endpoint.CreateScan))
	r.Methods("GET").Path("/api/v1/teams/{team_id}/scans/{scan_id}").Handler(newServer(e[endpoint.FindScan], endpoint.ScanRequest{}, logger, endpoint.FindScan))
	r.Methods("PUT").Path("/api/v1/teams/{team_id}/scans/{scan_id}/abort").Handler(newServer(e[endpoint.AbortScan], endpoint.ScanRequest{}, logger, endpoint.AbortScan))

	// Reports
	r.Methods("GET").Path("/api/v1/teams/{team_id}/scans/{scan_id}/report").Handler(newServer(e[endpoint.FindReport], endpoint.ReportRequest{}, logger, endpoint.FindReport))
	r.Methods("POST").Path("/api/v1/teams/{team_id}/scans/{scan_id}/report").Handler(newServer(e[endpoint.CreateReport], endpoint.ReportRequest{}, logger, endpoint.CreateReport))
	r.Methods("POST").Path("/api/v1/teams/{team_id}/scans/{scan_id}/report/send").Handler(newServer(e[endpoint.SendReport], endpoint.ReportRequest{}, logger, endpoint.SendReport))
	r.Methods("GET").Path("/api/v1/teams/{team_id}/scans/{scan_id}/report/email").Handler(newServer(e[endpoint.FindReportEmail], endpoint.ReportRequest{}, logger, endpoint.FindReportEmail))

	// Send Digest Report
	r.Methods("POST").Path("/api/v1/teams/{team_id}/report/digest").Handler(newServer(e[endpoint.SendDigestReport], endpoint.SendDigestReportRequest{}, logger, endpoint.SendDigestReport))

	// Stats
	r.Methods("GET").Path("/api/v1/teams/{team_id}/stats/coverage").Handler(newServer(e[endpoint.StatsCoverage], endpoint.StatsCoverageRequest{}, logger, endpoint.StatsCoverage))

	// Vulnerability DB
	r.Methods("GET").Path("/api/v1/teams/{team_id}/findings").Handler(newServer(e[endpoint.ListFindings], endpoint.FindingsRequest{}, logger, endpoint.ListFindings))
	r.Methods("GET").Path("/api/v1/teams/{team_id}/findings/issues").Handler(newServer(e[endpoint.ListFindingsIssues], endpoint.FindingsRequest{}, logger, endpoint.ListFindingsIssues))
	r.Methods("GET").Path("/api/v1/teams/{team_id}/findings/issues/{issue_id}").Handler(newServer(e[endpoint.ListFindingsByIssue], endpoint.FindingsByIssueRequest{}, logger, endpoint.ListFindingsByIssue))
	r.Methods("GET").Path("/api/v1/teams/{team_id}/findings/targets").Handler(newServer(e[endpoint.ListFindingsTargets], endpoint.FindingsRequest{}, logger, endpoint.ListFindingsTargets))
	r.Methods("GET").Path("/api/v1/teams/{team_id}/findings/targets/{target_id}").Handler(newServer(e[endpoint.ListFindingsByTarget], endpoint.FindingsByTargetRequest{}, logger, endpoint.ListFindingsByTarget))
	r.Methods("GET").Path("/api/v1/teams/{team_id}/findings/{finding_id}").Handler(newServer(e[endpoint.FindFinding], endpoint.FindingsRequest{}, logger, endpoint.FindFinding))
	r.Methods("PATCH").Path("/api/v1/teams/{team_id}/findings/{finding_id}").Handler(newServer(e[endpoint.UpdateFinding], endpoint.UpdateFindingRequest{}, logger, endpoint.UpdateFinding))
	r.Methods("GET").Path("/api/v1/teams/{team_id}/stats/mttr").Handler(newServer(e[endpoint.StatsMTTR], endpoint.StatsRequest{}, logger, endpoint.StatsMTTR))
	r.Methods("GET").Path("/api/v1/teams/{team_id}/stats/open").Handler(newServer(e[endpoint.StatsOpen], endpoint.StatsRequest{}, logger, endpoint.StatsOpen))
	r.Methods("GET").Path("/api/v1/teams/{team_id}/stats/fixed").Handler(newServer(e[endpoint.StatsFixed], endpoint.StatsRequest{}, logger, endpoint.StatsFixed))
	r.Methods("GET").Path("/api/v1/stats/mttr").Handler(newServer(e[endpoint.GlobalStatsMTTR], endpoint.StatsRequest{}, logger, endpoint.GlobalStatsMTTR))

	return r
}
