/*
Copyright 2021 Adevinta
*/

package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/adevinta/vulcan-api/cmd/vulcan-api-cli/client"
	"github.com/goadesign/goa"
	goaclient "github.com/goadesign/goa/client"
)

var (
	ErrTeamNotFound = errors.New("Team not found")
)

type ErrorVulcanAPI struct {
	Code int
	Err  string `json:"error"`
	Type string
}

func (v ErrorVulcanAPI) String() string {
	return fmt.Sprintf("http status: %d, type: %s, message: %s", v.Code, v.Type, v.Err)
}

func (v ErrorVulcanAPI) Error() string {
	return v.String()
}

type Config struct {
	Key     string
	Format  string
	Scheme  string
	Host    string
	Timeout time.Duration
	Dump    bool
}

type CLI struct {
	ctx context.Context
	c   *client.Client
}

func NewCLI(ctx context.Context, cfg Config, logger *log.Logger) *CLI {
	httpClient := http.DefaultClient
	httpClient.Timeout = cfg.Timeout

	bearerSigner := newBearerSigner(cfg.Key, cfg.Format)

	c := client.New(goaclient.HTTPClientDoer(httpClient))
	c.UserAgent = "Vulcan-cli/0"
	c.Host = cfg.Host
	c.Scheme = cfg.Scheme
	c.Dump = cfg.Dump
	c.SetBearerSigner(bearerSigner)

	goaLogger := goa.NewLogger(logger)
	ctx = goa.WithLogger(ctx, goaLogger)

	return &CLI{
		c:   c,
		ctx: ctx,
	}
}

// newBearerSigner returns the request signer used for authenticating
// against the Bearer security scheme.
func newBearerSigner(key, format string) goaclient.Signer {
	return &goaclient.APIKeySigner{
		SignQuery: false,
		KeyName:   "authorization",
		KeyValue:  key,
		Format:    format,
	}
}

func (cli *CLI) Teams() ([]*Team, error) {
	ctx := cli.ctx
	c := cli.c

	resp, err := c.ListTeams(ctx, client.ListTeamsPath(), nil)
	if err != nil {
		return nil, err
	}

	apiTeams, err := c.DecodeTeamCollection(resp)
	if err != nil {
		return nil, err
	}

	var teams []*Team
	for _, apiT := range apiTeams {
		i := Info{
			Description: DereferenceString(apiT.Description),
			Tag:         DereferenceString(apiT.Tag),
		}

		t := &Team{
			ID:   DereferenceString(apiT.ID),
			Name: *apiT.Name,
			Info: i,
		}

		teams = append(teams, t)
	}

	return teams, nil
}

func (cli *CLI) TeamByName(name string) (*Team, error) {
	teams, err := cli.Teams()
	if err != nil {
		return nil, err
	}

	for _, t := range teams {
		if t.Name == name {
			return t, nil
		}
	}

	return nil, ErrTeamNotFound
}

func (cli *CLI) CreateTeam(t *Team) (string, error) {
	ctx := cli.ctx
	c := cli.c

	p := &client.TeamPayload{
		Name: t.Name,
	}
	if err := p.Validate(); err != nil {
		return "", err
	}

	resp, err := c.CreateTeams(ctx, client.CreateTeamsPath(), p)
	if err != nil {
		return "", err
	}

	apiTeam, err := c.DecodeTeam(resp)
	if err != nil {
		return "", err
	}

	return DereferenceString(apiTeam.ID), nil
}

func (cli *CLI) UpdateTeamInfo(t *Team) error {
	ctx := cli.ctx
	c := cli.c

	p := &client.TeamUpdatePayload{
		Name:        PtrString(t.Name),
		Description: PtrString(t.Description),
		Tag:         PtrString(t.Tag),
	}

	resp, err := c.UpdateTeams(ctx, client.UpdateTeamsPath(t.ID), p)
	if err != nil {
		return err
	}

	_, err = c.DecodeTeam(resp)
	return err
}

func (cli *CLI) Recipients(teamID string) (Recipients, error) {
	ctx := cli.ctx
	c := cli.c

	resp, err := c.ListRecipients(ctx, client.ListRecipientsPath(teamID))
	if err != nil {
		return nil, err
	}

	apiRecipients, err := c.DecodeRecipientCollection(resp)
	if err != nil {
		return nil, err
	}

	var recipients Recipients
	for _, apiRecipient := range apiRecipients {
		r := Recipient{
			Email: *apiRecipient.Email,
		}
		recipients = append(recipients, r)
	}

	return recipients, nil
}

func (cli *CLI) AddRecipients(teamID string, recipients []Recipient) error {
	ctx := cli.ctx
	c := cli.c

	var r []string
	for _, recipient := range recipients {
		r = append(r, recipient.Email)
	}

	p := &client.RecipientsPayload{
		Emails: r,
	}
	if err := p.Validate(); err != nil {
		return err
	}

	resp, err := c.UpdateRecipients(ctx, client.UpdateRecipientsPath(teamID), p)
	if err != nil {
		return err
	}

	_, err = c.DecodeRecipientCollection(resp)
	return err
}

func (cli *CLI) Members(teamID string) (Members, error) {
	ctx := cli.ctx
	c := cli.c

	resp, err := c.ListTeamMembers(ctx, client.ListTeamMembersPath(teamID))
	if err != nil {
		return nil, err
	}

	apiMembers, err := c.DecodeTeammemberCollection(resp)
	if err != nil {
		return nil, err
	}

	var members Members
	for _, apiM := range apiMembers {
		if apiM.User == nil {
			return nil, fmt.Errorf("nil user for a member of team %s", teamID)
		}
		u := &Member{
			User: User{
				ID:       DereferenceString(apiM.User.ID),
				Email:    DereferenceString(apiM.User.Email),
				Admin:    DereferenceBool(apiM.User.Admin),
				Observer: DereferenceBool(apiM.User.Observer),
			},
			Role: DereferenceString(apiM.Role),
		}

		members = append(members, u)
	}

	return members, nil
}

func (cli *CLI) CreateMember(teamID, email, role string) (string, error) {
	ctx := cli.ctx
	c := cli.c

	p := &client.TeamMemberPayload{
		Email: &email,
		Role:  &role,
	}

	resp, err := c.CreateTeamMembers(ctx, client.CreateTeamMembersPath(teamID), p)
	if err != nil {
		return "", err
	}

	memberEntity, err := c.DecodeTeammember(resp)
	if err != nil {
		return "", err
	}
	if memberEntity.User == nil {
		return "", errors.New("user is nil")
	}

	return DereferenceString(memberEntity.User.ID), nil
}

func (cli *CLI) UpdateMember(teamID, userID, role string) error {
	ctx := cli.ctx
	c := cli.c

	p := &client.TeamMemberUpdatePayload{
		Role: &role,
	}

	resp, err := c.UpdateTeamMembers(ctx, client.UpdateTeamMembersPath(teamID, userID), p)
	if err != nil {
		return err
	}

	memberEntity, err := c.DecodeTeammember(resp)
	if err != nil {
		return err
	}
	if memberEntity.User == nil {
		return errors.New("user is nil")
	}

	return nil
}

func (cli *CLI) UpdateAsset(teamID, ID, rolfp string, alias string) error {
	ctx := cli.ctx
	c := cli.c

	p := &client.AssetUpdatePayload{Rolfp: &rolfp, Alias: &alias}
	resp, err := c.UpdateAssets(ctx, client.UpdateAssetsPath(teamID, ID), p)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusOK {
		_, err := c.DecodeAsset(resp)
		return err
	}
	// Here err is nil and status code is not OK.
	var verr ErrorVulcanAPI
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error updating the asset %s of the team %s, %w", ID, teamID, err)
	}
	err = json.Unmarshal(content, &verr)
	if err == nil {
		return fmt.Errorf("error updating the asset %s of the team %s, %w", ID, teamID, verr)
	}
	return fmt.Errorf("error updating the asset %s of the team %s, response status: %d", ID, teamID, resp.StatusCode)
}

func (cli *CLI) UpdateSchedule(teamID, programID, cron string) error {
	ctx := cli.ctx
	c := cli.c

	p := &client.ScheduleUpdatePayload{Cron: &cron}
	resp, err := c.UpdateSchedule(ctx, client.UpdateSchedulePath(teamID, programID), p)
	if err != nil {
		return err
	}

	_, err = c.DecodeProgram(resp)
	return err
}

func (cli *CLI) DeleteMember(teamID, userID string) error {
	ctx := cli.ctx
	c := cli.c

	resp, err := c.DeleteTeamMembers(ctx, client.DeleteTeamMembersPath(teamID, userID))
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("wrong status when removing member '%s' from team '%s' (%s)", userID, teamID, resp.Status)
	}

	return nil
}

func (cli *CLI) Assets(teamID string) (Assets, error) {
	ctx := cli.ctx
	c := cli.c

	resp, err := c.ListAssets(ctx, client.ListAssetsPath(teamID), nil)
	if err != nil {
		return nil, err
	}

	apiAssets, err := c.DecodeAssetCollection(resp)
	if err != nil {
		return nil, err
	}

	var assets Assets
	for _, apiA := range apiAssets {
		if apiA.Type == nil {
			return nil, fmt.Errorf("nil assettype for asset %s in team %s", DereferenceString(apiA.ID), teamID)
		}

		a := &Asset{
			ID:        DereferenceString(apiA.ID),
			Target:    DereferenceString(apiA.Identifier),
			AssetType: DereferenceString(apiA.Type.Name),
			Rolfp:     DereferenceString(apiA.Rolfp),
			Alias:     DereferenceString(apiA.Alias),
		}

		assets = append(assets, a)
	}

	return assets, nil
}

func (cli *CLI) CreateAsset(teamID, target, assetType, rolfp, alias string) (Assets, error) {
	ctx := cli.ctx
	c := cli.c

	p := &client.CreateAssetPayload{
		Assets: []*client.AssetPayload{
			&client.AssetPayload{
				Identifier: target,
				Type:       &assetType,
				Rolfp:      &rolfp,
				Alias:      &alias,
			},
		},
	}
	if err := p.Validate(); err != nil {
		return nil, err
	}

	resp, err := c.CreateAssets(ctx, client.CreateAssetsPath(teamID), p)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated {
		errorMedia, e := c.DecodeError(resp)
		if e != nil {
			return nil, e
		}
		return nil, errors.New(errorMedia.Error)
	}

	createdAssets, err := c.DecodeAssetCollection(resp)
	if err != nil {
		return nil, err
	}

	assets := make([]*Asset, len(createdAssets))
	for i, asset := range createdAssets {
		assets[i] = &Asset{
			ID:        DereferenceString(asset.ID),
			Target:    DereferenceString(asset.Identifier),
			AssetType: DereferenceString(asset.Type.Name),
		}
	}

	return assets, nil
}

func (cli *CLI) DeleteAsset(teamID, assetID string) error {
	ctx := cli.ctx
	c := cli.c
	_, err := c.DeleteAssets(ctx, client.DeleteAssetsPath(teamID, assetID))
	return err
}

func (cli *CLI) OrphanAssets(assets Assets, groups Groups) (OrphanAssets, error) {
	var orphans OrphanAssets
loop:
	for _, asset := range assets {
		for _, g := range groups {
			for _, a := range g.Assets {
				if asset.ID == a.ID {
					continue loop
				}
			}
		}

		orphans.Assets = append(orphans.Assets, asset)
	}

	return orphans, nil
}

func (cli *CLI) ForeignAssets(assets Assets, groups Groups) (ForeignAssets, error) {
	var foreigns ForeignAssets
	for _, g := range groups {
		for _, a := range g.Assets {
			if _, ok := assets.Find(a.ID); !ok {
				foreigns.Assets = append(foreigns.Assets, a)
			}
		}
	}

	return foreigns, nil
}

func (cli *CLI) DuppedAssets(assets Assets) (DuppedAssets, error) {
	var dupped DuppedAssets
	for _, a := range assets {
		if assets.IsDupped(a.Target, a.AssetType, a.ID) {
			dupped.Assets = append(dupped.Assets, a)
		}
	}

	return dupped, nil
}

func (cli *CLI) Groups(teamID string) (Groups, error) {
	ctx := cli.ctx
	c := cli.c

	resp, err := c.ListGroup(ctx, client.ListGroupPath(teamID))
	if err != nil {
		return nil, err
	}

	apiGroups, err := c.DecodeGroupCollection(resp)
	if err != nil {
		return nil, err
	}

	var groups Groups
	for _, apiG := range apiGroups {
		resp, err := c.ListAssetGroup(ctx, client.ListAssetGroupPath(teamID, *apiG.ID))
		if err != nil {
			return nil, err
		}
		apiAssets, err := c.DecodeAssetCollection(resp)
		if err != nil {
			return nil, err
		}

		var assets Assets
		for _, apiA := range apiAssets {
			if apiA.Type == nil {
				return nil, fmt.Errorf("nil assettype for asset %s in team %s", DereferenceString(apiA.ID), teamID)
			}

			a := &Asset{
				ID:        DereferenceString(apiA.ID),
				Target:    DereferenceString(apiA.Identifier),
				AssetType: DereferenceString(apiA.Type.Name),
				Rolfp:     DereferenceString(apiA.Rolfp),
				Alias:     DereferenceString(apiA.Alias),
			}

			assets = append(assets, a)
		}

		g := &Group{
			ID:     DereferenceString(apiG.ID),
			Name:   DereferenceString(apiG.Name),
			Assets: assets,
		}

		groups = append(groups, g)
	}

	return groups, nil
}

func (cli *CLI) CreateGroup(teamID, name string) (string, error) {
	ctx := cli.ctx
	c := cli.c

	p := &client.GroupPayload{
		Name: name,
	}
	if err := p.Validate(); err != nil {
		return "", err
	}

	resp, err := c.CreateGroup(ctx, client.CreateGroupPath(teamID), p)
	if err != nil {
		return "", err
	}

	groupEntity, err := c.DecodeGroup(resp)
	if err != nil {
		return "", err
	}

	return DereferenceString(groupEntity.ID), nil
}

func (cli *CLI) AssociateAsset(teamID, groupID, assetID string) error {
	ctx := cli.ctx
	c := cli.c

	p := &client.AssetGroupPayload{
		AssetID: assetID,
	}
	if err := p.Validate(); err != nil {
		return err
	}

	resp, err := c.CreateAssetGroup(ctx, client.CreateAssetGroupPath(teamID, groupID), p)
	if err != nil {
		return err
	}

	_, err = c.DecodeAssetgroup(resp)
	return err
}

func (cli *CLI) DeassociateAsset(teamID, groupID, assetID string) error {
	ctx := cli.ctx
	c := cli.c

	resp, err := c.DeleteAssetGroup(ctx, client.DeleteAssetGroupPath(teamID, groupID, assetID))
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("wrong status when deassociating group '%s' and asset '%s' in team '%s' (%s)", groupID, assetID, teamID, resp.Status)
	}

	return nil
}

func (cli *CLI) Users() (Users, error) {
	ctx := cli.ctx
	c := cli.c

	resp, err := c.ListUser(ctx, client.ListUserPath())
	if err != nil {
		return nil, err
	}

	apiUsers, err := c.DecodeUserCollection(resp)
	if err != nil {
		return nil, err
	}

	var users Users
	for _, apiU := range apiUsers {
		u := &User{
			ID:        DereferenceString(apiU.ID),
			Firstname: DereferenceString(apiU.Firstname),
			Lastname:  DereferenceString(apiU.Lastname),
			Email:     DereferenceString(apiU.Email),
			Admin:     DereferenceBool(apiU.Admin),
			Observer:  DereferenceBool(apiU.Observer),
			Active:    DereferenceBool(apiU.Active),
		}

		users = append(users, u)
	}

	return users, nil
}

func (cli *CLI) Unassigned(users Users, teams []*Team) (Unassigned, error) {
	var unassigned Unassigned
	for _, u := range users {
		var assigned bool
		for _, t := range teams {
			if _, ok := t.Members.Find(u.Email); ok {
				assigned = true
				break
			}
		}

		if !assigned {
			unassigned.Users = append(unassigned.Users, u)
		}
	}

	return unassigned, nil
}

func (cli *CLI) Policies(teamID string) (Policies, error) {
	ctx := cli.ctx
	c := cli.c

	resp, err := c.ListPolicies(ctx, client.ListPoliciesPath(teamID))
	if err != nil {
		return nil, err
	}

	apiPolicies, err := c.DecodePolicyCollection(resp)
	if err != nil {
		return nil, err
	}

	var policies Policies
	for _, apiP := range apiPolicies {
		p := &Policy{
			ID:   DereferenceString(apiP.ID),
			Name: DereferenceString(apiP.Name),
		}

		settings, err := cli.Settings(teamID, p.ID)
		if err != nil {
			return nil, err
		}
		p.Settings = settings

		policies = append(policies, p)
	}

	return policies, nil
}

func (cli *CLI) Settings(teamID, policyID string) (SettingsCollection, error) {
	ctx := cli.ctx
	c := cli.c

	resp, err := c.ListPolicySettings(ctx, client.ListPolicySettingsPath(teamID, policyID))
	if err != nil {
		return nil, err
	}

	apiSettingsResponse, err := c.DecodePolicysettingCollection(resp)
	if err != nil {
		return nil, err
	}

	var settings SettingsCollection
	for _, apiS := range apiSettingsResponse {
		s := &Settings{
			ID:      DereferenceString(apiS.ID),
			Name:    DereferenceString(apiS.ChecktypeName),
			Options: DereferenceString(apiS.Options),
		}

		settings = append(settings, s)
	}

	return settings, nil
}

func (cli *CLI) Programs(teamID string) (Programs, error) {
	ctx := cli.ctx
	c := cli.c

	resp, err := c.ListPrograms(ctx, client.ListProgramsPath(teamID))
	if err != nil {
		return nil, err
	}

	apiPrograms, err := c.DecodeProgramCollection(resp)
	if err != nil {
		return nil, err
	}

	var programs Programs

	for _, apiP := range apiPrograms {
		p := &Program{
			ID:           DereferenceString(apiP.ID),
			Name:         DereferenceString(apiP.Name),
			Autosend:     DereferenceBool(apiP.Autosend),
			Cron:         DereferenceString(&apiP.Schedule.Cron),
			PolicyGroups: convertPolicyGroups(apiP.PolicyGroups),
		}

		programs = append(programs, p)
	}

	return programs, nil
}

func (cli *CLI) Program(teamID, programID string) (*Program, error) {
	ctx := cli.ctx
	c := cli.c

	resp, err := c.ShowPrograms(ctx, client.ShowProgramsPath(teamID, programID))
	if err != nil {
		return nil, err
	}

	apiProgram, err := c.DecodeProgram(resp)
	if err != nil {
		return nil, err
	}

	return &Program{
		ID:           DereferenceString(apiProgram.ID),
		Name:         DereferenceString(apiProgram.Name),
		Autosend:     DereferenceBool(apiProgram.Autosend),
		Cron:         DereferenceString(&apiProgram.Schedule.Cron),
		PolicyGroups: convertPolicyGroups(apiProgram.PolicyGroups),
	}, nil
}

func (cli *CLI) ProgramByName(teamName, programName string) (*Program, error) {
	t, err := cli.TeamByName(teamName)
	if err != nil {
		return nil, err
	}

	programs, err := cli.Programs(t.ID)
	if err != nil {
		return nil, err
	}

	for _, p := range programs {
		if p.Name == programName {
			return p, nil
		}
	}

	return nil, fmt.Errorf("program '%s' not found for team '%s'", programName, teamName)
}

func convertPolicyGroups(apiPGroups []*client.ProgramPolicyGroup) []PolicyGroup {
	pgroups := []PolicyGroup{}
	for _, v := range apiPGroups {
		// This copies the current value to avoid having the usual troubles
		// when ranging over an slice of pointers.
		v := v
		if v == nil || v.Group == nil || v.Policy == nil {
			continue
		}
		if v.Group.ID == nil || v.Policy.ID == nil {
			continue
		}
		pgroups = append(pgroups, PolicyGroup{
			GroupID:  *v.Group.ID,
			PolicyID: *v.Policy.ID,
		})
	}
	return pgroups
}

func (cli *CLI) AddProgramsToPolicies(programs Programs, policies Policies) {
	for _, program := range programs {
		for _, g := range program.PolicyGroups {
			if policy, ok := policies.Find(g.PolicyID); ok {
				policy.Programs = append(policy.Programs, program)
			}
		}
	}
}

func (cli *CLI) LaunchScan(t *Team, programID string) (*Scan, error) {
	ctx := cli.ctx
	c := cli.c

	p := &client.ScanPayload{
		ProgramID: programID,
	}
	if err := p.Validate(); err != nil {
		return nil, err
	}

	resp, err := c.CreateScan(ctx, client.CreateScanPath(t.ID), p)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("wrong status when creating scan for program '%s' in team '%s' (%s)", programID, t.ID, resp.Status)
	}

	scan, err := c.DecodeScan(resp)
	if err != nil {
		return nil, err
	}

	s := &Scan{
		ID:      DereferenceString(scan.ID),
		Program: programID,
		Status:  "CREATED",
		Team:    t.Name,
	}

	return s, nil
}

func (cli *CLI) Scan(teamID, scanID string) (*Scan, error) {
	ctx := cli.ctx
	c := cli.c

	resp, err := c.ShowScan(ctx, client.ShowScanPath(teamID, scanID))
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("wrong status when getting scan '%s' in team '%s' (%s)", scanID, teamID, resp.Status)
	}

	scan, err := c.DecodeScan(resp)
	if err != nil {
		return nil, err
	}

	s := &Scan{
		ID:     scanID,
		Status: DereferenceString(scan.Status),
	}
	if scan.Program != nil {
		s.Program = DereferenceString(scan.Program.ID)
	}

	return s, nil
}

func (cli *CLI) RefreshScan(s *Scan) (*Scan, error) {
	ctx := cli.ctx
	c := cli.c

	t, err := cli.TeamByName(s.Team)
	if err != nil {
		return nil, err
	}

	resp, err := c.ShowScan(ctx, client.ShowScanPath(t.ID, s.ID))
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("wrong status when refreshing scan '%s' in team '%s' (%s)", s.ID, s.Team, resp.Status)
	}

	scan, err := c.DecodeScan(resp)
	if err != nil {
		return nil, err
	}

	s.Status = DereferenceString(scan.Status)

	return s, nil
}

func (cli *CLI) ReportEmail(teamName string, scanID string) (string, error) {
	ctx := cli.ctx
	c := cli.c

	t, err := cli.TeamByName(teamName)
	if err != nil {
		return "", err
	}

	resp, err := c.EmailScanReport(ctx, client.EmailScanReportPath(t.ID, scanID))
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("wrong status when retrieving report for scan '%s' in team '%s' (%s)", scanID, teamName, resp.Status)
	}

	email, err := c.DecodeReportemail(resp)
	if err != nil {
		return "", err
	}

	return DereferenceString(email.EmailBody), nil
}

func (cli *CLI) SendReport(teamName string, scanID string) error {
	ctx := cli.ctx
	c := cli.c

	t, err := cli.TeamByName(teamName)
	if err != nil {
		return err
	}

	resp, err := c.SendScanReport(ctx, client.SendScanReportPath(t.ID, scanID))
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("wrong status when sending report for scan '%s' in team '%s' (%s)", scanID, teamName, resp.Status)
	}

	return nil
}

func (cli *CLI) Findings(teamID string, minScore float64, status *string) ([]*Finding, error) {
	ctx := cli.ctx
	c := cli.c

	var findings []*Finding
	more := true
	page := 0.0
	for more {
		resp, err := c.ListFindingsFindings(
			ctx,                                     // ctx
			client.ListFindingsFindingsPath(teamID), // path
			nil,                                     // atDate
			nil,                                     // identifier
			nil,                                     // identifiers
			nil,                                     // issueID
			nil,                                     // maxDate
			nil,                                     // maxScore
			nil,                                     // minDate
			&minScore,                               // minScore
			&page,                                   // page
			nil,                                     // size
			nil,                                     // sortBy
			status,                                  // status
			nil,                                     // targetID
		)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("wrong status when retrieving findings for team '%s' (%s)", teamID, resp.Status)
		}

		apiFindingsList, err := c.DecodeFindingsList(resp)
		if err != nil {
			return nil, err
		}

		for _, apiF := range apiFindingsList.Findings {
			f := Finding{
				Summary:   *apiF.Issue.Summary,
				Score:     *apiF.Score,
				Target:    *apiF.Target.Identifier,
				Checktype: *apiF.Source.Component,
			}

			findings = append(findings, &f)
		}

		more = *apiFindingsList.Pagination.More
		page++
	}

	return findings, nil
}

func (cli *CLI) Coverage(teamID string) (Coverage, error) {
	ctx := cli.ctx
	c := cli.c

	var coverage Coverage

	resp, err := c.CoverageStats(ctx, client.CoverageStatsPath(teamID))
	if err != nil {
		return coverage, err
	}

	if resp.StatusCode != http.StatusOK {
		return coverage, fmt.Errorf("wrong status when retrieving coverage for team '%s' (%s)", teamID, resp.Status)
	}

	cAPI, err := c.DecodeStatscoverage(resp)
	if err != nil {
		return coverage, err
	}

	if cAPI.Coverage == nil {
		return coverage, errors.New("unepected nil value from coverage response")
	}

	coverage = Coverage(*cAPI.Coverage)

	return coverage, nil
}
