/*
Copyright 2021 Adevinta
*/

package cli

import (
	"errors"
	"fmt"
	"strings"

	types "github.com/adevinta/vulcan-types"
	"github.com/google/go-cmp/cmp"
)

type Op struct {
	cli *CLI
}

type UpdateTeamInfoOp struct {
	NewInfo Info
	OldInfo Info

	Team *Team

	Op
}

func (o UpdateTeamInfoOp) String() string {
	return fmt.Sprintf("Information updated for team '%s':\n%s", o.Team.Name, cmp.Diff(o.OldInfo, o.NewInfo))
}

func (o UpdateTeamInfoOp) Apply() error {
	o.Team.Info = o.NewInfo
	return o.cli.UpdateTeamInfo(o.Team)
}

type UpdateRecipientsOp struct {
	NewRecipients Recipients
	OldRecipients Recipients
	Team          *Team

	Op
}

func (o UpdateRecipientsOp) String() string {
	return fmt.Sprintf("Recipients updated for team '%s':\n%s", o.Team.Name, cmp.Diff(o.OldRecipients, o.NewRecipients))
}

func (o UpdateRecipientsOp) Apply() error {
	return o.cli.AddRecipients(o.Team.ID, o.NewRecipients)
}

type CreateMemberOp struct {
	NewMember *Member
	Team      *Team

	Op
}

func (o CreateMemberOp) String() string {
	return fmt.Sprintf("Adding member '%s' with role '%s' to team '%s'", o.NewMember.Email, o.NewMember.Role, o.Team.Name)
}

func (o CreateMemberOp) Apply() error {
	id, err := o.cli.CreateMember(o.Team.ID, o.NewMember.Email, o.NewMember.Role)
	if err != nil {
		return err
	}
	o.NewMember.ID = id

	return nil
}

type UpdateProgramOp struct {
	NewProgram *Program
	OldProgram *Program
	Team       *Team

	Op
}

func (o UpdateProgramOp) String() string {
	return fmt.Sprintf("Updating schedule of program '%s' from '%s' to '%s' for team '%s'", o.NewProgram.ID, o.OldProgram.Cron, o.NewProgram.Cron, o.Team.Name)
}

func (o UpdateProgramOp) Apply() error {
	return o.cli.UpdateSchedule(o.Team.ID, o.NewProgram.ID, o.NewProgram.Cron)
}

type UpdateMemberOp struct {
	NewMember *Member
	OldMember *Member
	Team      *Team

	Op
}

func (o UpdateMemberOp) String() string {
	return fmt.Sprintf("Updating member '%s' from role '%s' to role '%s' for team '%s'", o.NewMember.Email, o.OldMember.Role, o.NewMember.Role, o.Team.Name)
}

func (o UpdateMemberOp) Apply() error {
	return o.cli.UpdateMember(o.Team.ID, o.NewMember.ID, o.NewMember.Role)
}

type UpdateAssetOp struct {
	NewAsset *Asset
	OldAsset *Asset
	Team     *Team

	Op
}

func (o UpdateAssetOp) String() string {
	return fmt.Sprintf("Updating asset '%s' from rolfp '%s' to rolfp '%s' and from alias '%s' to alias '%s' for team '%s'", o.OldAsset.ID, o.OldAsset.Rolfp, o.NewAsset.Rolfp, o.OldAsset.Alias, o.NewAsset.Alias, o.Team.Name)
}

func (o UpdateAssetOp) Apply() error {
	return o.cli.UpdateAsset(o.Team.ID, o.NewAsset.ID, o.NewAsset.Rolfp, o.NewAsset.Alias)
}

type DeleteMemberOp struct {
	Member *Member
	Team   *Team

	Op
}

func (o DeleteMemberOp) String() string {
	return fmt.Sprintf("Deleting member '%s' with role '%s' from team '%s'", o.Member.Email, o.Member.Role, o.Team.Name)
}

func (o DeleteMemberOp) Apply() error {
	return o.cli.DeleteMember(o.Team.ID, o.Member.ID)
}

type CreateGroupOp struct {
	NewGroup *Group
	Team     *Team

	Op
}

func (o CreateGroupOp) String() string {
	return fmt.Sprintf("Adding group '%s' to team '%s'", o.NewGroup.Name, o.Team.Name)
}

func (o CreateGroupOp) Apply() error {
	id, err := o.cli.CreateGroup(o.Team.ID, o.NewGroup.Name)
	if err != nil {
		return err
	}
	o.NewGroup.ID = id

	return nil
}

type CreateAssetOp struct {
	NewAsset *Asset
	Group    *Group
	Team     *Team

	Op
}

func (o CreateAssetOp) String() string {
	return fmt.Sprintf("Creating asset '%s' with asset type '%s' and roflp '%s' for team '%s'", o.NewAsset.Target, o.NewAsset.AssetType, o.NewAsset.Rolfp, o.Team.Name)
}

func (o CreateAssetOp) Apply() error {
	createdAssets, err := o.cli.CreateAsset(o.Team.ID, o.NewAsset.Target, o.NewAsset.AssetType, o.NewAsset.Rolfp, o.NewAsset.Alias)
	if err != nil {
		return fmt.Errorf("error applying operation: %s, error: %w", o, err)
	}
	// For 'normal assets' we will only have one created asset returned from API.
	// For assets created usign smart capability we can have N assets created,
	// so pick first one to comply with the asset association operation already
	// created, which references o.NewAsset.ID.
	o.NewAsset.ID = createdAssets[0].ID
	o.NewAsset.AssetType = createdAssets[0].AssetType

	// In case of assets being created using smart capability, for every auto
	// generated asset from server, add asset to the group specified for the new asset,
	// and create and execute a new asset association operation.
	for _, smartAsset := range createdAssets[1:] {
		o.Group.Assets = append(o.Group.Assets, smartAsset)

		associateOp := AssociationOp{
			smartAsset,
			o.Group,
			o.Team,
			o.Op,
		}
		associateOp.Apply() // nolint
	}

	return nil
}

type DeleteAssetOp struct {
	Asset *Asset
	Team  *Team
	Op
}

func (o DeleteAssetOp) String() string {
	return fmt.Sprintf("Deleting asset '%s' with asset type '%s' for team '%s'", o.Asset.Target, o.Asset.AssetType, o.Team.Name)
}

func (o DeleteAssetOp) Apply() error {
	err := o.cli.DeleteAsset(o.Team.ID, o.Asset.ID)
	return err
}

type AssociationOp struct {
	Asset    *Asset
	NewGroup *Group
	Team     *Team

	Op
}

func (o AssociationOp) String() string {
	return fmt.Sprintf("Adding asset '%s' with target '%s' and asset type '%s'to group '%s' of team '%s'", o.Asset.ID, o.Asset.Target, o.Asset.AssetType, o.NewGroup.Name, o.Team.Name)
}

func (o AssociationOp) Apply() error {
	return o.cli.AssociateAsset(o.Team.ID, o.NewGroup.ID, o.Asset.ID)
}

type DeassociationOp struct {
	Asset    *Asset
	OldGroup *Group
	Team     *Team

	Op
}

func (o DeassociationOp) String() string {
	return fmt.Sprintf("Removing asset '%s' with target '%s' from group '%s' of team '%s'", o.Asset.ID, o.Asset.Target, o.OldGroup.Name, o.Team.Name)
}

func (o DeassociationOp) Apply() error {
	return o.cli.DeassociateAsset(o.Team.ID, o.OldGroup.ID, o.Asset.ID)
}

type Journal struct {
	LocalTeams  map[string]*Team
	RemoteTeams map[string]*Team

	UpdatedTeams        []UpdateTeamInfoOp
	UpdatedRecipients   []UpdateRecipientsOp
	NewMembers          []CreateMemberOp
	DeletedMembers      []DeleteMemberOp
	UpdatedMembers      []UpdateMemberOp
	NewGroups           []CreateGroupOp
	NewAssets           []CreateAssetOp
	UpdateAssets        []UpdateAssetOp
	NewAssociations     []AssociationOp
	DeletedAssociations []DeassociationOp
	DeleteAssets        []DeleteAssetOp

	UpdatedPrograms []UpdateProgramOp

	cli *CLI
}

func indexByTeamName(localTeams []*Team, remoteTeams []*Team) (indexedLocalTeams map[string]*Team, indexedRemoteTeams map[string]*Team) {
	indexedLocalTeams = make(map[string]*Team)
	indexedRemoteTeams = make(map[string]*Team)

	for _, lt := range localTeams {
		indexedLocalTeams[lt.Name] = lt
	}

	for _, rt := range remoteTeams {
		indexedRemoteTeams[rt.Name] = rt
	}

	return indexedLocalTeams, indexedRemoteTeams
}

func NewJournal(localTeams []*Team, remoteTeams []*Team, cli *CLI) (*Journal, error) {
	ilt, irt := indexByTeamName(localTeams, remoteTeams)

	j := &Journal{
		LocalTeams:  ilt,
		RemoteTeams: irt,
		cli:         cli,
	}

	return j, nil
}

func (j *Journal) BuildModifications() error {
	teamsToUpdate, err := j.GuessTeamsToUpdate()
	if err != nil {
		return err
	}
	j.UpdatedTeams = teamsToUpdate

	recipientsToUpdate, err := j.GuessRecipientsToUpdate()
	if err != nil {
		return err
	}
	j.UpdatedRecipients = recipientsToUpdate

	membersToCreate, err := j.GuessMembersToCreate()
	if err != nil {
		return err
	}
	j.NewMembers = membersToCreate

	membersToDelete, err := j.GuessMembersToDelete()
	if err != nil {
		return err
	}
	j.DeletedMembers = membersToDelete

	membersToUpdate, err := j.GuessMembersToUpdate()
	if err != nil {
		return err
	}
	j.UpdatedMembers = membersToUpdate

	groupsToCreate, err := j.GuessGroupsToCreate()
	if err != nil {
		return err
	}
	j.NewGroups = groupsToCreate

	assetsToCreate, err := j.GuessAssetsToCreate()
	if err != nil {
		return err
	}
	j.NewAssets = assetsToCreate

	assetsToAssociate, err := j.GuessAssetsToAssociate()
	if err != nil {
		return err
	}
	j.NewAssociations = assetsToAssociate

	assetsToDeassociate, err := j.GuessAssetsToDeassociate()
	if err != nil {
		return err
	}

	j.DeletedAssociations = assetsToDeassociate

	assetsToUpdate, err := j.GuessAssetsToUpdate()
	if err != nil {
		return err
	}
	j.UpdateAssets = assetsToUpdate

	programsToUpdate, err := j.GuessProgramsToUpdate()
	if err != nil {
		return err
	}
	j.UpdatedPrograms = programsToUpdate

	return nil
}

func (j *Journal) BuildPruneModifications() {
	assets := []DeleteAssetOp{}
	for _, lt := range j.LocalTeams {
		for _, a := range lt.OrphanAssets.Assets {
			op := DeleteAssetOp{
				Asset: a,
				Team:  lt,
			}
			op.cli = j.cli
			assets = append(assets, op)
		}
	}
	j.DeleteAssets = assets
}

func (j *Journal) Apply() error {
	for _, op := range j.UpdatedTeams {
		if err := op.Apply(); err != nil {
			return err
		}
	}

	for _, op := range j.UpdatedRecipients {
		if err := op.Apply(); err != nil {
			return err
		}
	}

	for _, op := range j.NewMembers {
		if err := op.Apply(); err != nil {
			return err
		}
	}

	for _, op := range j.DeletedMembers {
		if err := op.Apply(); err != nil {
			return err
		}
	}

	for _, op := range j.UpdatedMembers {
		if err := op.Apply(); err != nil {
			return err
		}
	}

	for _, op := range j.NewGroups {
		if err := op.Apply(); err != nil {
			return err
		}
	}

	for _, op := range j.NewAssets {
		if err := op.Apply(); err != nil {

			return err
		}
	}

	for _, op := range j.NewAssociations {
		if err := op.Apply(); err != nil {
			return err
		}
	}

	for _, op := range j.DeletedAssociations {
		if err := op.Apply(); err != nil {
			return err
		}
	}

	for _, op := range j.DeleteAssets {
		if err := op.Apply(); err != nil {
			return err
		}
	}

	for _, op := range j.UpdateAssets {
		if err := op.Apply(); err != nil {
			return err
		}
	}

	for _, op := range j.UpdatedPrograms {
		if err := op.Apply(); err != nil {
			return err
		}
	}

	return nil
}

func (j *Journal) String() string {
	var b strings.Builder

	for _, t := range j.UpdatedTeams {
		printInfo(&b, t)
	}

	for _, r := range j.UpdatedRecipients {
		printInfo(&b, r)
	}

	for _, m := range j.NewMembers {
		printInfo(&b, m)
	}

	for _, m := range j.DeletedMembers {
		printInfo(&b, m)
	}

	for _, m := range j.UpdatedMembers {
		printInfo(&b, m)
	}

	for _, g := range j.NewGroups {
		printInfo(&b, g)
	}

	for _, a := range j.NewAssets {
		printInfo(&b, a)
	}

	for _, a := range j.NewAssociations {
		printInfo(&b, a)
	}

	for _, a := range j.DeletedAssociations {
		printInfo(&b, a)
	}

	for _, a := range j.DeleteAssets {
		printInfo(&b, a)
	}

	for _, a := range j.UpdateAssets {
		printInfo(&b, a)
	}

	for _, p := range j.UpdatedPrograms {
		printInfo(&b, p)
	}

	return b.String()
}

func printInfo(b *strings.Builder, str fmt.Stringer) {
	fmt.Fprintf(b, "[*] %s\n", str)
}

func (j *Journal) GuessTeamsToUpdate() ([]UpdateTeamInfoOp, error) {
	var t []UpdateTeamInfoOp
	for name, lt := range j.LocalTeams {
		rt, ok := j.RemoteTeams[name]
		if !ok {
			return nil, fmt.Errorf("local team %v doesn't exist remotely", name)
		}

		if !cmp.Equal(lt.Info, rt.Info) {
			op := UpdateTeamInfoOp{
				NewInfo: lt.Info,
				OldInfo: rt.Info,
				Team:    lt,
			}
			op.cli = j.cli
			t = append(t, op)
		}
	}

	return t, nil
}

func (j *Journal) GuessRecipientsToUpdate() ([]UpdateRecipientsOp, error) {
	var r []UpdateRecipientsOp
	for name, lt := range j.LocalTeams {
		rt, ok := j.RemoteTeams[name]
		if !ok {
			return nil, fmt.Errorf("local team %v doesn't exist remotely", name)
		}

		if !cmp.Equal(lt.Recipients, rt.Recipients) {
			op := UpdateRecipientsOp{
				NewRecipients: lt.Recipients,
				OldRecipients: rt.Recipients,
				Team:          rt,
			}
			op.cli = j.cli
			r = append(r, op)
		}
	}

	return r, nil
}

func (j *Journal) GuessMembersToCreate() ([]CreateMemberOp, error) {
	var c []CreateMemberOp
	for name, lt := range j.LocalTeams {
		rt, ok := j.RemoteTeams[name]
		if !ok {
			return nil, fmt.Errorf("local team %v doesn't exist remotely", name)
		}

		for _, lm := range lt.Members {
			rm, ok := rt.Members.Find(lm.Email)
			if ok && lm.ID != rm.ID && lm.ID != "" {
				return nil, fmt.Errorf("member %v has different IDs locally (%v) and remotely (%v)", lm.Email, lm.ID, rm.ID)
			}
			if !ok {
				m := CreateMemberOp{
					NewMember: lm,
					Team:      rt,
				}
				m.cli = j.cli
				c = append(c, m)
			}
		}
	}

	return c, nil
}

func (j *Journal) GuessMembersToUpdate() ([]UpdateMemberOp, error) {
	var c []UpdateMemberOp
	for name, lt := range j.LocalTeams {
		rt, ok := j.RemoteTeams[name]
		if !ok {
			return nil, fmt.Errorf("local team %v doesn't exist remotely", name)
		}

		for _, lm := range lt.Members {
			rm, ok := rt.Members.Find(lm.Email)
			// !ok is not a problem because member will be created in previous steps.
			if ok {
				if lm.ID != rm.ID && lm.ID != "" {
					return nil, fmt.Errorf("member %v has different IDs locally (%v) and remotely (%v)", lm.Email, lm.ID, rm.ID)
				}
				if lm.Role != rm.Role {
					lm.ID = rm.ID
					m := UpdateMemberOp{
						NewMember: lm,
						OldMember: rm,
						Team:      rt,
					}
					m.cli = j.cli
					c = append(c, m)
				}
			}
		}
	}

	return c, nil
}

func (j *Journal) GuessProgramsToUpdate() ([]UpdateProgramOp, error) {
	var ops []UpdateProgramOp
	for name, lt := range j.LocalTeams {
		lt := lt
		rt, ok := j.RemoteTeams[name]
		if !ok {
			return nil, fmt.Errorf("local team %v doesn't exist remotely", name)
		}
		for _, lp := range lt.Programs {
			rp, ok := rt.Programs.Find(lp.ID)
			// TODO !ok, c create a new program, is not taken into account yet.
			// By now the only feature regarding programs is to update the
			// schedule of and existing program.
			if !ok {
				continue
			}
			if lp.Cron != rp.Cron {
				op := UpdateProgramOp{
					NewProgram: lp,
					OldProgram: rp,
					Team:       rt,
				}
				op.cli = j.cli
				ops = append(ops, op)
			}

		}
	}
	return ops, nil
}

func (j *Journal) GuessMembersToDelete() ([]DeleteMemberOp, error) {
	var ms []DeleteMemberOp
	for name, lt := range j.LocalTeams {
		rt, ok := j.RemoteTeams[name]
		if !ok {
			return nil, fmt.Errorf("local team '%v' doesn't exist remotely", name)
		}

		for _, rm := range rt.Members {
			if _, ok := lt.Members.Find(rm.Email); !ok {
				m := DeleteMemberOp{
					Member: rm,
					Team:   rt,
				}
				m.cli = j.cli
				ms = append(ms, m)
			}
		}
	}

	return ms, nil
}

func (j *Journal) GuessGroupsToCreate() ([]CreateGroupOp, error) {
	var gs []CreateGroupOp
	for name, lt := range j.LocalTeams {
		rt, ok := j.RemoteTeams[name]
		if !ok {
			return nil, fmt.Errorf("local team %v doesn't exist remotely", name)
		}

		for _, lg := range lt.Groups {
			_, ok := rt.Groups.FindByName(lg.Name)
			if !ok {
				g := CreateGroupOp{
					NewGroup: lg,
					Team:     rt,
				}
				g.cli = j.cli
				gs = append(gs, g)
			}
		}
	}

	return gs, nil
}

func (j *Journal) GuessAssetsToCreate() ([]CreateAssetOp, error) {
	var as []CreateAssetOp
	for name, lt := range j.LocalTeams {
		_, ok := j.RemoteTeams[name]
		if !ok {
			return nil, fmt.Errorf("local team %v doesn't exist remotely", name)
		}

		for _, lg := range lt.Groups {
			for _, la := range lg.Assets {
				if la.ID != "" {
					continue
				}
				if err := validateAsset(*la); err != nil {
					fmt.Printf("invalid asset: %s, with type: %s, skipping its creation\n", la.Target, la.AssetType)
					continue
				}
				a := CreateAssetOp{
					NewAsset: la,
					Group:    lg,
					Team:     lt,
				}
				a.cli = j.cli
				as = append(as, a)
			}
		}
	}

	return as, nil
}

func validateAsset(a Asset) error {
	switch a.AssetType {
	case "Hostname":
		if !types.IsHostname(a.Target) {
			return errors.New("identifier is not a valid Hostname")
		}
	case "AWSAccount":
		if !types.IsAWSARN(a.Target) {
			return errors.New("identifier is not a valid AWSAccount")
		}
	case "DockerImage":
		if !types.IsDockerImage(a.Target) {
			return errors.New("identifier is not a valid DockerImage")
		}
	case "GitRepository":
		if !types.IsGitRepository(a.Target) {
			return errors.New("identifier is not a valid GitRepository")
		}
	case "IP":
		if strings.HasSuffix(a.Target, "/32") {
			if !types.IsHost(a.Target) {
				return errors.New("identifier is not a valid Host")
			}
		} else {
			if !types.IsIP(a.Target) {
				return errors.New("identifier is not a valid IP")
			}
		}
	case "IPRange":
		if !types.IsCIDR(a.Target) {
			return errors.New("identifier is not a valid CIDR block")
		}
	case "WebAddress":
		if !types.IsWebAddress(a.Target) {
			return errors.New("identifier is not a valid WebAddress")
		}
	case "DomainName":
		if ok, _ := types.IsDomainName(a.Target); !ok {
			return errors.New("identifier is not a valid DomainName")
		}
	default:
		// If none of the previous case match, force a validation error
		return errors.New("Asset type not supported")
	}
	return nil
}

func (j *Journal) GuessAssetsToUpdate() ([]UpdateAssetOp, error) {
	var c []UpdateAssetOp
	for name, lt := range j.LocalTeams {
		rt, ok := j.RemoteTeams[name]
		if !ok {
			return nil, fmt.Errorf("local team %v doesn't exist remotely", name)
		}

		for _, lg := range lt.Groups {
			for _, la := range lg.Assets {
				ra, ok := rt.Assets.Find(la.ID)
				// !ok is not a problem because the asset will be created in previous steps.
				if !ok {
					continue
				}
				// By now we are only taking into account the ROLFP and the
				// alias we are not considering, for instance, updating the
				// asset type.
				if ra.Rolfp == la.Rolfp && ra.Alias == la.Alias {
					continue
				}
				la.ID = ra.ID
				la.AssetType = ra.AssetType
				la.Target = ra.Target
				a := UpdateAssetOp{
					OldAsset: ra,
					NewAsset: la,
					Team:     rt,
				}
				a.cli = j.cli
				c = append(c, a)
			}
		}
	}

	return c, nil
}

func (j *Journal) GuessAssetsToAssociate() ([]AssociationOp, error) {
	var as []AssociationOp
	for name, lt := range j.LocalTeams {
		rt, ok := j.RemoteTeams[name]
		if !ok {
			return nil, fmt.Errorf("local team '%v' doesn't exist remotely", name)
		}

		for _, lg := range lt.Groups {
			rg, ok := rt.Groups.FindByName(lg.Name)
			if !ok {
				if lg.ID != "" {
					return nil, fmt.Errorf("local group with ID '%s' and name '%s' of team '%v' does not exist remotely", lg.ID, lg.Name, name)
				}
				// Group will be created when applying CreateGroupOp.
				rg = &Group{}
			}

			for _, la := range lg.Assets {
				// Only relevant for DryRun, because association comes after asset creation.
				if _, ok := rt.Assets.Find(la.ID); !ok && la.ID != "" {
					return nil, fmt.Errorf("local asset with ID '%s' and team '%v' does not exist remotely", la.ID, name)
				}

				if _, ok := rg.FindAssetByID(la.ID); !ok || la.ID == "" {
					a := AssociationOp{
						Asset:    la,
						NewGroup: lg,
						Team:     rt,
					}
					a.cli = j.cli
					as = append(as, a)
				}
			}
		}
	}

	return as, nil
}

func (j *Journal) GuessAssetsToDeassociate() ([]DeassociationOp, error) {
	var as []DeassociationOp
	for name, lt := range j.LocalTeams {
		rt, ok := j.RemoteTeams[name]
		if !ok {
			return nil, fmt.Errorf("local team '%v' doesn't exist remotely", name)
		}

		for _, lg := range lt.Groups {
			rg, ok := rt.Groups.FindByName(lg.Name)
			// Do not check deassociation for new groups
			if !ok {
				continue
			}

			for _, ra := range rg.Assets {
				if _, ok := lg.FindAssetByID(ra.ID); !ok {
					a := DeassociationOp{
						Asset:    ra,
						OldGroup: rg,
						Team:     rt,
					}
					a.cli = j.cli
					as = append(as, a)
				}
			}
		}
	}

	return as, nil
}
