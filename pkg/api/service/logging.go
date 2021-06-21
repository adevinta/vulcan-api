// Code generated by Impl, DO NOT EDIT.
/*
Copyright 2021 Adevinta
*/

package service

import (
	"context"
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	kithttp "github.com/go-kit/kit/transport/http"

	"github.com/adevinta/vulcan-api/pkg/api"
)

type Middleware func(service api.VulcanitoService) api.VulcanitoService

func LoggingMiddleware(logger log.Logger) Middleware {
	return func(next api.VulcanitoService) api.VulcanitoService {
		return loggingMiddleware{logger, next}
	}
}

type loggingMiddleware struct {
	logger log.Logger
	next   api.VulcanitoService
}

// mySprintf returns a formatted string where if an argument is a byte array is printed as a string,
// and all the other types are in the default format
func mySprintf(a interface{}) string {
	t, ok := a.([]byte)
	if ok {
		return fmt.Sprintf("%s", t)
	}
	return fmt.Sprintf("%+v", a)
}
func (middleware loggingMiddleware) Healthcheck(ctx context.Context) error {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "Healthcheck")
	}()
	return middleware.next.Healthcheck(ctx)
}

func (middleware loggingMiddleware) ListUsers(ctx context.Context) ([]*api.User, error) {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "ListUsers")
	}()
	return middleware.next.ListUsers(ctx)
}

func (middleware loggingMiddleware) CreateUser(ctx context.Context, user api.User) (*api.User, error) {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "CreateUser", "user", mySprintf(user))
	}()
	return middleware.next.CreateUser(ctx, user)
}

func (middleware loggingMiddleware) UpdateUser(ctx context.Context, user api.User) (*api.User, error) {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "UpdateUser", "user", mySprintf(user))
	}()
	return middleware.next.UpdateUser(ctx, user)
}

func (middleware loggingMiddleware) FindUser(ctx context.Context, userID string) (*api.User, error) {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "FindUser", "userID", mySprintf(userID))
	}()
	return middleware.next.FindUser(ctx, userID)
}

func (middleware loggingMiddleware) DeleteUser(ctx context.Context, userID string) error {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "DeleteUser", "userID", mySprintf(userID))
	}()
	return middleware.next.DeleteUser(ctx, userID)
}

func (middleware loggingMiddleware) GenerateAPIToken(ctx context.Context, userID string) (*api.Token, error) {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "GenerateAPIToken", "userID", mySprintf(userID))
	}()
	return middleware.next.GenerateAPIToken(ctx, userID)
}

func (middleware loggingMiddleware) CreateTeam(ctx context.Context, team api.Team, ownerEmail string) (*api.Team, error) {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "CreateTeam", "team", mySprintf(team), "ownerEmail", mySprintf(ownerEmail))
	}()
	return middleware.next.CreateTeam(ctx, team, ownerEmail)
}

func (middleware loggingMiddleware) UpdateTeam(ctx context.Context, team api.Team) (*api.Team, error) {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "UpdateTeam", "team", mySprintf(team))
	}()
	return middleware.next.UpdateTeam(ctx, team)
}

func (middleware loggingMiddleware) FindTeam(ctx context.Context, teamID string) (*api.Team, error) {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "FindTeam", "teamID", mySprintf(teamID))
	}()
	return middleware.next.FindTeam(ctx, teamID)
}

func (middleware loggingMiddleware) FindTeamByTag(ctx context.Context, tag string) (*api.Team, error) {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "FindTeamByTag", "tag", mySprintf(tag))
	}()
	return middleware.next.FindTeamByTag(ctx, tag)
}

func (middleware loggingMiddleware) DeleteTeam(ctx context.Context, teamID string) error {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "DeleteTeam", "teamID", mySprintf(teamID))
	}()
	return middleware.next.DeleteTeam(ctx, teamID)
}

func (middleware loggingMiddleware) ListTeams(ctx context.Context) ([]*api.Team, error) {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "ListTeams")
	}()
	return middleware.next.ListTeams(ctx)
}

func (middleware loggingMiddleware) FindTeamsByUser(ctx context.Context, userID string) ([]*api.Team, error) {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "FindTeamsByUser", "userID", mySprintf(userID))
	}()
	return middleware.next.FindTeamsByUser(ctx, userID)
}

func (middleware loggingMiddleware) FindTeamMember(ctx context.Context, teamID string, userID string) (*api.UserTeam, error) {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "FindTeamMember", "teamID", mySprintf(teamID), "userID", mySprintf(userID))
	}()
	return middleware.next.FindTeamMember(ctx, teamID, userID)
}

func (middleware loggingMiddleware) CreateTeamMember(ctx context.Context, teamUser api.UserTeam) (*api.UserTeam, error) {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "CreateTeamMember", "teamUser", mySprintf(teamUser))
	}()
	return middleware.next.CreateTeamMember(ctx, teamUser)
}

func (middleware loggingMiddleware) UpdateTeamMember(ctx context.Context, teamUser api.UserTeam) (*api.UserTeam, error) {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "UpdateTeamMember", "teamUser", mySprintf(teamUser))
	}()
	return middleware.next.UpdateTeamMember(ctx, teamUser)
}

func (middleware loggingMiddleware) DeleteTeamMember(ctx context.Context, teamID string, userID string) error {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "DeleteTeamMember", "teamID", mySprintf(teamID), "userID", mySprintf(userID))
	}()
	return middleware.next.DeleteTeamMember(ctx, teamID, userID)
}

func (middleware loggingMiddleware) UpdateRecipients(ctx context.Context, teamID string, emails []string) error {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "UpdateRecipients", "teamID", mySprintf(teamID), "emails", mySprintf(emails))
	}()
	return middleware.next.UpdateRecipients(ctx, teamID, emails)
}

func (middleware loggingMiddleware) ListRecipients(ctx context.Context, teamID string) ([]*api.Recipient, error) {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "ListRecipients", "teamID", mySprintf(teamID))
	}()
	return middleware.next.ListRecipients(ctx, teamID)
}

func (middleware loggingMiddleware) ListAssets(ctx context.Context, teamID string, asset api.Asset) ([]*api.Asset, error) {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "ListAssets", "teamID", mySprintf(teamID), "asset", mySprintf(asset))
	}()
	return middleware.next.ListAssets(ctx, teamID, asset)
}

func (middleware loggingMiddleware) CreateAssets(ctx context.Context, assets []api.Asset, groups []api.Group, annotations []api.AssetAnnotation) ([]api.Asset, error) {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "CreateAssets", "assets", mySprintf(assets), "groups", mySprintf(groups), "annotations", mySprintf(annotations))
	}()
	return middleware.next.CreateAssets(ctx, assets, groups, annotations)
}

func (middleware loggingMiddleware) CreateAssetsMultiStatus(ctx context.Context, assets []api.Asset, groups []api.Group, annotations []api.AssetAnnotation) ([]api.AssetCreationResponse, error) {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "CreateAssetsMultiStatus", "assets", mySprintf(assets), "groups", mySprintf(groups), "annotations", mySprintf(annotations))
	}()
	return middleware.next.CreateAssetsMultiStatus(ctx, assets, groups, annotations)
}

func (middleware loggingMiddleware) FindAsset(ctx context.Context, asset api.Asset) (*api.Asset, error) {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "FindAsset", "asset", mySprintf(asset))
	}()
	return middleware.next.FindAsset(ctx, asset)
}

func (middleware loggingMiddleware) UpdateAsset(ctx context.Context, asset api.Asset) (*api.Asset, error) {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "UpdateAsset", "asset", mySprintf(asset))
	}()
	return middleware.next.UpdateAsset(ctx, asset)
}

func (middleware loggingMiddleware) DeleteAsset(ctx context.Context, asset api.Asset) error {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "DeleteAsset", "asset", mySprintf(asset))
	}()
	return middleware.next.DeleteAsset(ctx, asset)
}

func (middleware loggingMiddleware) DeleteAllAssets(ctx context.Context, teamID string) error {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "DeleteAllAssets", "teamID", mySprintf(teamID))
	}()
	return middleware.next.DeleteAllAssets(ctx, teamID)
}

func (middleware loggingMiddleware) GetAssetType(ctx context.Context, assetTypeName string) (*api.AssetType, error) {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "GetAssetType", "assetTypeName", mySprintf(assetTypeName))
	}()
	return middleware.next.GetAssetType(ctx, assetTypeName)
}

func (middleware loggingMiddleware) ListAssetAnnotations(ctx context.Context, teamID string, assetID string) ([]*api.AssetAnnotation, error) {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "ListAssetAnnotations", "teamID", mySprintf(teamID), "assetID", mySprintf(assetID))
	}()
	return middleware.next.ListAssetAnnotations(ctx, teamID, assetID)
}

func (middleware loggingMiddleware) CreateAssetAnnotations(ctx context.Context, teamID string, assetID string, annotations []*api.AssetAnnotation) ([]*api.AssetAnnotation, error) {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "CreateAssetAnnotations", "teamID", mySprintf(teamID), "assetID", mySprintf(assetID), "annotations", mySprintf(annotations))
	}()
	return middleware.next.CreateAssetAnnotations(ctx, teamID, assetID, annotations)
}

func (middleware loggingMiddleware) UpdateAssetAnnotations(ctx context.Context, teamID string, assetID string, annotations []*api.AssetAnnotation) ([]*api.AssetAnnotation, error) {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "UpdateAssetAnnotations", "teamID", mySprintf(teamID), "assetID", mySprintf(assetID), "annotations", mySprintf(annotations))
	}()
	return middleware.next.UpdateAssetAnnotations(ctx, teamID, assetID, annotations)
}

func (middleware loggingMiddleware) DeleteAssetAnnotations(ctx context.Context, teamID string, assedID string, annotations []*api.AssetAnnotation) ([]*api.AssetAnnotation, error) {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "DeleteAssetAnnotations", "teamID", mySprintf(teamID), "assedID", mySprintf(assedID), "annotations", mySprintf(annotations))
	}()
	return middleware.next.DeleteAssetAnnotations(ctx, teamID, assedID, annotations)
}

func (middleware loggingMiddleware) ListGroups(ctx context.Context, teamID string, groupName string) ([]*api.Group, error) {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "ListGroups", "teamID", mySprintf(teamID), "groupName", mySprintf(groupName))
	}()
	return middleware.next.ListGroups(ctx, teamID, groupName)
}

func (middleware loggingMiddleware) CreateGroup(ctx context.Context, group api.Group) (*api.Group, error) {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "CreateGroup", "group", mySprintf(group))
	}()
	return middleware.next.CreateGroup(ctx, group)
}

func (middleware loggingMiddleware) FindGroup(ctx context.Context, group api.Group) (*api.Group, error) {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "FindGroup", "group", mySprintf(group))
	}()
	return middleware.next.FindGroup(ctx, group)
}

func (middleware loggingMiddleware) UpdateGroup(ctx context.Context, group api.Group) (*api.Group, error) {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "UpdateGroup", "group", mySprintf(group))
	}()
	return middleware.next.UpdateGroup(ctx, group)
}

func (middleware loggingMiddleware) DeleteGroup(ctx context.Context, group api.Group) error {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "DeleteGroup", "group", mySprintf(group))
	}()
	return middleware.next.DeleteGroup(ctx, group)
}

func (middleware loggingMiddleware) GroupAsset(ctx context.Context, assetGroup api.AssetGroup, teamID string) (*api.AssetGroup, error) {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "GroupAsset", "assetGroup", mySprintf(assetGroup), "teamID", mySprintf(teamID))
	}()
	return middleware.next.GroupAsset(ctx, assetGroup, teamID)
}

func (middleware loggingMiddleware) UngroupAsset(ctx context.Context, assetGroup api.AssetGroup, teamID string) error {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "UngroupAsset", "assetGroup", mySprintf(assetGroup), "teamID", mySprintf(teamID))
	}()
	return middleware.next.UngroupAsset(ctx, assetGroup, teamID)
}

func (middleware loggingMiddleware) ListAssetGroup(ctx context.Context, assetGroup api.AssetGroup, teamID string) ([]*api.Asset, error) {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "ListAssetGroup", "assetGroup", mySprintf(assetGroup), "teamID", mySprintf(teamID))
	}()
	return middleware.next.ListAssetGroup(ctx, assetGroup, teamID)
}

func (middleware loggingMiddleware) ListPrograms(ctx context.Context, teamID string) ([]*api.Program, error) {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "ListPrograms", "teamID", mySprintf(teamID))
	}()
	return middleware.next.ListPrograms(ctx, teamID)
}

func (middleware loggingMiddleware) CreateProgram(ctx context.Context, program api.Program, teamID string) (*api.Program, error) {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "CreateProgram", "program", mySprintf(program), "teamID", mySprintf(teamID))
	}()
	return middleware.next.CreateProgram(ctx, program, teamID)
}

func (middleware loggingMiddleware) FindProgram(ctx context.Context, programID string, teamID string) (*api.Program, error) {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "FindProgram", "programID", mySprintf(programID), "teamID", mySprintf(teamID))
	}()
	return middleware.next.FindProgram(ctx, programID, teamID)
}

func (middleware loggingMiddleware) UpdateProgram(ctx context.Context, program api.Program, teamID string) (*api.Program, error) {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "UpdateProgram", "program", mySprintf(program), "teamID", mySprintf(teamID))
	}()
	return middleware.next.UpdateProgram(ctx, program, teamID)
}

func (middleware loggingMiddleware) DeleteProgram(ctx context.Context, program api.Program, teamID string) error {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "DeleteProgram", "program", mySprintf(program), "teamID", mySprintf(teamID))
	}()
	return middleware.next.DeleteProgram(ctx, program, teamID)
}

func (middleware loggingMiddleware) CreateSchedule(ctx context.Context, programID string, cronExpr string, teamID string) (*api.Program, error) {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "CreateSchedule", "programID", mySprintf(programID), "cronExpr", mySprintf(cronExpr), "teamID", mySprintf(teamID))
	}()
	return middleware.next.CreateSchedule(ctx, programID, cronExpr, teamID)
}

func (middleware loggingMiddleware) DeleteSchedule(ctx context.Context, programID string, teamID string) (*api.Program, error) {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "DeleteSchedule", "programID", mySprintf(programID), "teamID", mySprintf(teamID))
	}()
	return middleware.next.DeleteSchedule(ctx, programID, teamID)
}

func (middleware loggingMiddleware) ScheduleGlobalProgram(ctx context.Context, programID string, cronExpr string) error {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "ScheduleGlobalProgram", "programID", mySprintf(programID), "cronExpr", mySprintf(cronExpr))
	}()
	return middleware.next.ScheduleGlobalProgram(ctx, programID, cronExpr)
}

func (middleware loggingMiddleware) ListPolicies(ctx context.Context, teamID string) ([]*api.Policy, error) {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "ListPolicies", "teamID", mySprintf(teamID))
	}()
	return middleware.next.ListPolicies(ctx, teamID)
}

func (middleware loggingMiddleware) CreatePolicy(ctx context.Context, policy api.Policy) (*api.Policy, error) {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "CreatePolicy", "policy", mySprintf(policy))
	}()
	return middleware.next.CreatePolicy(ctx, policy)
}

func (middleware loggingMiddleware) FindPolicy(ctx context.Context, policyID string) (*api.Policy, error) {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "FindPolicy", "policyID", mySprintf(policyID))
	}()
	return middleware.next.FindPolicy(ctx, policyID)
}

func (middleware loggingMiddleware) UpdatePolicy(ctx context.Context, policy api.Policy) (*api.Policy, error) {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "UpdatePolicy", "policy", mySprintf(policy))
	}()
	return middleware.next.UpdatePolicy(ctx, policy)
}

func (middleware loggingMiddleware) DeletePolicy(ctx context.Context, policy api.Policy) error {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "DeletePolicy", "policy", mySprintf(policy))
	}()
	return middleware.next.DeletePolicy(ctx, policy)
}

func (middleware loggingMiddleware) ListChecktypeSetting(ctx context.Context, policyID string) ([]*api.ChecktypeSetting, error) {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "ListChecktypeSetting", "policyID", mySprintf(policyID))
	}()
	return middleware.next.ListChecktypeSetting(ctx, policyID)
}

func (middleware loggingMiddleware) CreateChecktypeSetting(ctx context.Context, setting api.ChecktypeSetting) (*api.ChecktypeSetting, error) {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "CreateChecktypeSetting", "setting", mySprintf(setting))
	}()
	return middleware.next.CreateChecktypeSetting(ctx, setting)
}

func (middleware loggingMiddleware) FindChecktypeSetting(ctx context.Context, policyID string, checktypeSettingID string) (*api.ChecktypeSetting, error) {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "FindChecktypeSetting", "policyID", mySprintf(policyID), "checktypeSettingID", mySprintf(checktypeSettingID))
	}()
	return middleware.next.FindChecktypeSetting(ctx, policyID, checktypeSettingID)
}

func (middleware loggingMiddleware) UpdateChecktypeSetting(ctx context.Context, checktypeSetting api.ChecktypeSetting) (*api.ChecktypeSetting, error) {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "UpdateChecktypeSetting", "checktypeSetting", mySprintf(checktypeSetting))
	}()
	return middleware.next.UpdateChecktypeSetting(ctx, checktypeSetting)
}

func (middleware loggingMiddleware) DeleteChecktypeSetting(ctx context.Context, checktypeSettingID string) error {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "DeleteChecktypeSetting", "checktypeSettingID", mySprintf(checktypeSettingID))
	}()
	return middleware.next.DeleteChecktypeSetting(ctx, checktypeSettingID)
}

func (middleware loggingMiddleware) ListScans(ctx context.Context, teamID string, programID string) ([]*api.Scan, error) {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "ListScans", "teamID", mySprintf(teamID), "programID", mySprintf(programID))
	}()
	return middleware.next.ListScans(ctx, teamID, programID)
}

func (middleware loggingMiddleware) CreateScan(ctx context.Context, scan api.Scan, teamID string) (*api.Scan, error) {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "CreateScan", "scan", mySprintf(scan), "teamID", mySprintf(teamID))
	}()
	return middleware.next.CreateScan(ctx, scan, teamID)
}

func (middleware loggingMiddleware) FindScan(ctx context.Context, scanID string, teamID string) (*api.Scan, error) {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "FindScan", "scanID", mySprintf(scanID), "teamID", mySprintf(teamID))
	}()
	return middleware.next.FindScan(ctx, scanID, teamID)
}

func (middleware loggingMiddleware) AbortScan(ctx context.Context, scanID string, teamID string) (*api.Scan, error) {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "AbortScan", "scanID", mySprintf(scanID), "teamID", mySprintf(teamID))
	}()
	return middleware.next.AbortScan(ctx, scanID, teamID)
}

func (middleware loggingMiddleware) UpdateScan(ctx context.Context, scan api.Scan) (*api.Scan, error) {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "UpdateScan", "scan", mySprintf(scan))
	}()
	return middleware.next.UpdateScan(ctx, scan)
}

func (middleware loggingMiddleware) DeleteScan(ctx context.Context, scan api.Scan) error {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "DeleteScan", "scan", mySprintf(scan))
	}()
	return middleware.next.DeleteScan(ctx, scan)
}

func (middleware loggingMiddleware) FindReport(ctx context.Context, scanID string) (*api.Report, error) {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "FindReport", "scanID", mySprintf(scanID))
	}()
	return middleware.next.FindReport(ctx, scanID)
}

func (middleware loggingMiddleware) SendReport(ctx context.Context, scanID string, teamID string) error {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "SendReport", "scanID", mySprintf(scanID), "teamID", mySprintf(teamID))
	}()
	return middleware.next.SendReport(ctx, scanID, teamID)
}

func (middleware loggingMiddleware) GenerateReport(ctx context.Context, teamID string, teamName string, scanID string, autosend bool) error {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "GenerateReport", "teamID", mySprintf(teamID), "teamName", mySprintf(teamName), "scanID", mySprintf(scanID), "autosend", mySprintf(autosend))
	}()
	return middleware.next.GenerateReport(ctx, teamID, teamName, scanID, autosend)
}

func (middleware loggingMiddleware) RunGenerateReport(ctx context.Context, autosend bool, scanID string, programName string, teamID string, teamName string) error {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "RunGenerateReport", "autosend", mySprintf(autosend), "scanID", mySprintf(scanID), "programName", mySprintf(programName), "teamID", mySprintf(teamID), "teamName", mySprintf(teamName))
	}()
	return middleware.next.RunGenerateReport(ctx, autosend, scanID, programName, teamID, teamName)
}

func (middleware loggingMiddleware) ProcessScanCheckNotification(ctx context.Context, msg []byte) error {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "ProcessScanCheckNotification", "msg", mySprintf(msg))
	}()
	return middleware.next.ProcessScanCheckNotification(ctx, msg)
}

func (middleware loggingMiddleware) SendDigestReport(ctx context.Context, teamID string, startDate string, endDate string) error {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "SendDigestReport", "teamID", mySprintf(teamID), "startDate", mySprintf(startDate), "endDate", mySprintf(endDate))
	}()
	return middleware.next.SendDigestReport(ctx, teamID, startDate, endDate)
}

func (middleware loggingMiddleware) StatsCoverage(ctx context.Context, teamID string) (*api.StatsCoverage, error) {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "StatsCoverage", "teamID", mySprintf(teamID))
	}()
	return middleware.next.StatsCoverage(ctx, teamID)
}

func (middleware loggingMiddleware) ListFindings(ctx context.Context, params api.FindingsParams, pagination api.Pagination) (*api.FindingsList, error) {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "ListFindings", "params", mySprintf(params), "pagination", mySprintf(pagination))
	}()
	return middleware.next.ListFindings(ctx, params, pagination)
}

func (middleware loggingMiddleware) ListFindingsIssues(ctx context.Context, params api.FindingsParams, pagination api.Pagination) (*api.FindingsIssuesList, error) {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "ListFindingsIssues", "params", mySprintf(params), "pagination", mySprintf(pagination))
	}()
	return middleware.next.ListFindingsIssues(ctx, params, pagination)
}

func (middleware loggingMiddleware) ListFindingsByIssue(ctx context.Context, params api.FindingsParams, pagination api.Pagination) (*api.FindingsList, error) {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "ListFindingsByIssue", "params", mySprintf(params), "pagination", mySprintf(pagination))
	}()
	return middleware.next.ListFindingsByIssue(ctx, params, pagination)
}

func (middleware loggingMiddleware) ListFindingsTargets(ctx context.Context, params api.FindingsParams, pagination api.Pagination) (*api.FindingsTargetsList, error) {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "ListFindingsTargets", "params", mySprintf(params), "pagination", mySprintf(pagination))
	}()
	return middleware.next.ListFindingsTargets(ctx, params, pagination)
}

func (middleware loggingMiddleware) ListFindingsByTarget(ctx context.Context, params api.FindingsParams, pagination api.Pagination) (*api.FindingsList, error) {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "ListFindingsByTarget", "params", mySprintf(params), "pagination", mySprintf(pagination))
	}()
	return middleware.next.ListFindingsByTarget(ctx, params, pagination)
}

func (middleware loggingMiddleware) FindFinding(ctx context.Context, findingID string) (*api.Finding, error) {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "FindFinding", "findingID", mySprintf(findingID))
	}()
	return middleware.next.FindFinding(ctx, findingID)
}

func (middleware loggingMiddleware) CreateFindingOverwrite(ctx context.Context, findingOverwrite api.FindingOverwrite) error {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "CreateFindingOverwrite", "findingOverwrite", mySprintf(findingOverwrite))
	}()
	return middleware.next.CreateFindingOverwrite(ctx, findingOverwrite)
}

func (middleware loggingMiddleware) ListFindingOverwrites(ctx context.Context, findingID string) ([]*api.FindingOverwrite, error) {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "ListFindingOverwrites", "findingID", mySprintf(findingID))
	}()
	return middleware.next.ListFindingOverwrites(ctx, findingID)
}

func (middleware loggingMiddleware) StatsMTTR(ctx context.Context, params api.StatsParams) (*api.StatsMTTR, error) {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "StatsMTTR", "params", mySprintf(params))
	}()
	return middleware.next.StatsMTTR(ctx, params)
}

func (middleware loggingMiddleware) StatsExposure(ctx context.Context, params api.StatsParams) (*api.StatsExposure, error) {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "StatsExposure", "params", mySprintf(params))
	}()
	return middleware.next.StatsExposure(ctx, params)
}

func (middleware loggingMiddleware) StatsOpen(ctx context.Context, params api.StatsParams) (*api.StatsOpen, error) {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "StatsOpen", "params", mySprintf(params))
	}()
	return middleware.next.StatsOpen(ctx, params)
}

func (middleware loggingMiddleware) StatsFixed(ctx context.Context, params api.StatsParams) (*api.StatsFixed, error) {
	defer func() {
		XRequestID := ""
		if ctx != nil {
			XRequestID, _ = ctx.Value(kithttp.ContextKeyRequestXRequestID).(string)
		}
		_ = level.Debug(middleware.logger).Log("X-Request-ID", XRequestID, "service", "StatsFixed", "params", mySprintf(params))
	}()
	return middleware.next.StatsFixed(ctx, params)
}
