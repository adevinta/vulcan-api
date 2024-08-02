/*
Copyright 2021 Adevinta
*/

package api

import "github.com/adevinta/vulcan-api/pkg/saml"

type VulcanitoStore interface {
	Close() error

	NotFoundError(err error) bool

	Healthcheck() error

	ListIssues() ([]*Issue, error)

	FindJob(jobID string) (*Job, error)
	UpdateJob(job Job) (*Job, error)

	CreateUserIfNotExists(userData saml.UserData) error

	ListUsers() ([]*User, error)
	CreateUser(user User) (*User, error)
	UpdateUser(user User) (*User, error)
	FindUserByID(userID string) (*User, error)
	FindUserByEmail(email string) (*User, error)
	DeleteUserByID(userID string) error

	CreateTeam(team Team, ownerEmail string) (*Team, error)
	UpdateTeam(team Team) (*Team, error)
	FindTeam(teamID string) (*Team, error)
	FindTeamByIDForUser(ID, userID string) (*UserTeam, error)
	FindTeamsByUser(userID string) ([]*Team, error)
	FindTeamByName(name string) (*Team, error)
	FindTeamByTag(tag string) (*Team, error)
	FindTeamsByTags(tags []string) ([]*Team, error)
	FindTeamByProgram(programID string) (*Team, error)
	DeleteTeam(teamID string) error
	ListTeams() ([]*Team, error)

	CreateTeamMember(teamMember UserTeam) (*UserTeam, error)
	DeleteTeamMember(teamID string, userID string) error
	FindTeamMember(teamID string, userID string) (*UserTeam, error)
	UpdateTeamMember(teamMember UserTeam) (*UserTeam, error)

	UpdateRecipients(teamID string, emails []string) error
	ListRecipients(teamID string) ([]*Recipient, error)

	ListAssets(teamID string, asset Asset) ([]*Asset, error)
	FindAsset(teamID, assetID string) (*Asset, error)
	CreateAsset(asset Asset, groups []Group) (*Asset, error)
	CreateAssets(assets []Asset, groups []Group) ([]Asset, error)
	DeleteAsset(asset Asset) error
	DeleteAllAssets(teamID string) error
	UpdateAsset(asset Asset) (*Asset, error)
	MergeAssets(mergeOps AssetMergeOperations) error
	MergeAssetsAsync(teamID string, assets []Asset, groupName string) (*Job, error)

	GetAssetType(assetTypeName string) (*AssetType, error)

	ListAssetAnnotations(teamID string, assetID string) ([]*AssetAnnotation, error)
	CreateAssetAnnotations(teamID string, assetID string, annotations []*AssetAnnotation) ([]*AssetAnnotation, error)
	UpdateAssetAnnotations(teamID string, assetID string, annotations []*AssetAnnotation) ([]*AssetAnnotation, error)
	PutAssetAnnotations(teamID string, assetID string, annotations []*AssetAnnotation) ([]*AssetAnnotation, error)
	DeleteAssetAnnotations(teamID string, assetID string, annotations []*AssetAnnotation) error

	CreateGroup(group Group) (*Group, error)
	ListGroups(teamID, groupName string) ([]*Group, error)
	UpdateGroup(group Group) (*Group, error)
	DeleteGroup(group Group) error
	FindGroup(group Group) (*Group, error)
	// FindGroupInfo returns the info of the specified group
	// without loading the assets and teams associated to it.
	FindGroupInfo(group Group) (*Group, error)
	// DisjoinAssetsInGroups returns assets belonging to a team that are in a given
	// group but not in other groups.
	DisjoinAssetsInGroups(teamID, inGroupID string, notInGroupIDs []string) ([]*Asset, error)

	CountAssetsInGroups(teamID string, groupIDs []string) (int, error)

	GroupAsset(assetsGroup AssetGroup, teamID string) (*AssetGroup, error)
	ListAssetGroup(assetGroup AssetGroup, teamID string) ([]*AssetGroup, error)
	UngroupAssets(assetGroup AssetGroup, teamID string) error

	ListPrograms(teamID string) ([]*Program, error)
	CreateProgram(program Program, teamID string) (*Program, error)
	FindProgram(programID string, teamID string) (*Program, error)
	UpdateProgram(program Program, teamID string) (*Program, error)
	DeleteProgram(program Program, teamID string) error

	ListPolicies(teamID string) ([]*Policy, error)
	CreatePolicy(policy Policy) (*Policy, error)
	FindPolicy(policyID string) (*Policy, error)
	UpdatePolicy(policy Policy) (*Policy, error)
	DeletePolicy(policy Policy) error

	ListChecktypeSetting(policyID string) ([]*ChecktypeSetting, error)
	CreateChecktypeSetting(setting ChecktypeSetting) (*ChecktypeSetting, error)
	FindChecktypeSetting(checktypeSettingID string) (*ChecktypeSetting, error)
	UpdateChecktypeSetting(checktypeSetting ChecktypeSetting) (*ChecktypeSetting, error)
	DeleteChecktypeSetting(checktypeSettingID string) error

	FindGlobalProgramMetadata(programID string, teamID string) (*GlobalProgramsMetadata, error)
	UpsertGlobalProgramMetadata(teamID, program string, defaultAutosend bool, defaultDisabled bool, defaultCron string, autosend *bool, disabled *bool, cron *string) error
	DeleteProgramMetadata(program string) error

	CreateFindingOverwrite(findingOverwrite FindingOverwrite) error
	ListFindingOverwrites(findingID string) ([]*FindingOverwrite, error)
}
