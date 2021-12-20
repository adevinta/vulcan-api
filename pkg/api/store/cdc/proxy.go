/*
Copyright 2021 Adevinta
*/

package cdc

import (
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"github.com/adevinta/vulcan-api/pkg/api"
	"github.com/adevinta/vulcan-api/pkg/saml"
)

const (
	// default advisory lock ID so we can handle sync across instances.
	defLockID uint32 = 1869877622
	// CDCLogTag is a tag to use for logging.
	CDCLogTag = "CDC"
)

var (
	defStartAwakePeriod = 10 * time.Second
	defErrAwakePeriod   = 30 * time.Second
)

// BrokerProxy is a proxy applied to the
// storage component which acts as a broker
// following Change Data Capture pattern.
type BrokerProxy struct {
	logger log.Logger
	db     DB
	store  api.VulcanitoStore
	parser Parser
	cond   *sync.Cond
}

// NewBrokerProxy builds a new CDC broker proxy around VulcanitoStore.
func NewBrokerProxy(logger log.Logger, db DB, store api.VulcanitoStore,
	parser Parser) *BrokerProxy {

	bp := &BrokerProxy{
		logger: logger,
		db:     db,
		store:  store,
		parser: parser,
		cond:   sync.NewCond(&sync.Mutex{}),
	}

	go bp.start()
	// Awake broker initially to check
	// for remaining outbox log entries
	go bp.awakeAfter(defStartAwakePeriod)

	return bp
}

// awakeBroker takes the CDC broker lock so we
// let it finish its work in case it's processing
// messages, then releases the lock and signals for
// new changes.
func (b *BrokerProxy) awakeBroker() {
	b.cond.L.Lock()
	b.cond.L.Unlock()
	b.cond.Signal()
}

// awakeAfter awakes the CDC broker after d duration.
func (b *BrokerProxy) awakeAfter(d time.Duration) {
	time.Sleep(d)
	b.awakeBroker()
}

func (b *BrokerProxy) start() {
LOOP:
	for {
		// Wait for signal
		b.cond.L.Lock()
		b.cond.Wait()

		// Try to get lock.
		// As we are using the same lock ID,
		// concurrent running instances of API
		// might be locking it.
		lock, err := b.db.TryGetLock(defLockID)
		if err != nil {
			b.logErr(err)
			b.cond.L.Unlock()
			continue
		}

		if !lock.Acquired {
			b.db.ReleaseLock(lock) // nolint
			b.cond.L.Unlock()
			// Force awake to ensure new log events
			// are processed, because the other API
			// instance that currently has got the
			// advisory lock, might have not retrieved
			// the latest event from DB
			go b.awakeAfter(defErrAwakePeriod)
			continue
		}

		// Get log
		log, err := b.db.GetLog()
		if err != nil {
			b.logErr(err)
			b.db.ReleaseLock(lock) // nolint
			b.cond.L.Unlock()
			continue
		}

		// Process events
		for _, e := range log {
			nParsed := b.parser.Parse([]Event{e})
			if nParsed == 0 {
				err = b.db.FailedEvent(e)
				if err != nil {
					b.logErr(err)
				}

				b.db.ReleaseLock(lock) // nolint
				b.cond.L.Unlock()

				// If there was an errored event, do not
				// wait until next proxied method is called,
				// instead wake up broker after def duration
				go b.awakeAfter(defErrAwakePeriod)

				continue LOOP
			}

			err = b.db.CleanEvent(e)
			if err != nil {
				b.logErr(err)
				b.db.ReleaseLock(lock) // nolint
				b.cond.L.Unlock()

				// If there was an error cleaning event,
				// abort processing next events and force
				// broker awake after def duration
				go b.awakeAfter(defErrAwakePeriod)

				continue LOOP
			}
		}

		b.db.ReleaseLock(lock) // nolint
		b.cond.L.Unlock()
	}
}

func (b *BrokerProxy) logErr(err error) {
	_ = level.Error(b.logger).Log(
		"component", CDCLogTag, "error", err,
	)
}

/***
* Proxied Store Methods:
*
* To make CDC broker process generated data for a proxied
* method execution, call broker's 'awakeBroker' method inside
* a goroutine so we don't block the processing of main request.
* E.g.:
*	err := b.proxiedMethod()
*	go awakeBroker()
*	return err
***/

func (b *BrokerProxy) Close() error {
	return b.store.Close()
}

func (b *BrokerProxy) NotFoundError(err error) bool {
	return b.store.NotFoundError(err)
}

func (b *BrokerProxy) Healthcheck() error {
	return b.store.Healthcheck()
}

func (b *BrokerProxy) FindJob(jobID string) (*api.Job, error) {
	return b.store.FindJob(jobID)
}

func (b *BrokerProxy) CreateUserIfNotExists(userData saml.UserData) error {
	return b.store.CreateUserIfNotExists(userData)
}

func (b *BrokerProxy) ListUsers() ([]*api.User, error) {
	return b.store.ListUsers()
}
func (b *BrokerProxy) CreateUser(user api.User) (*api.User, error) {
	return b.store.CreateUser(user)
}
func (b *BrokerProxy) UpdateUser(user api.User) (*api.User, error) {
	return b.store.UpdateUser(user)
}
func (b *BrokerProxy) FindUserByID(userID string) (*api.User, error) {
	return b.store.FindUserByID(userID)
}
func (b *BrokerProxy) FindUserByEmail(email string) (*api.User, error) {
	return b.store.FindUserByEmail(email)
}
func (b *BrokerProxy) DeleteUserByID(userID string) error {
	return b.store.DeleteUserByID(userID)
}

func (b *BrokerProxy) CreateTeam(team api.Team, ownerEmail string) (*api.Team, error) {
	return b.store.CreateTeam(team, ownerEmail)
}
func (b *BrokerProxy) UpdateTeam(team api.Team) (*api.Team, error) {
	return b.store.UpdateTeam(team)
}
func (b *BrokerProxy) FindTeam(teamID string) (*api.Team, error) {
	return b.store.FindTeam(teamID)
}
func (b *BrokerProxy) FindTeamByIDForUser(ID, userID string) (*api.UserTeam, error) {
	return b.store.FindTeamByIDForUser(ID, userID)
}
func (b *BrokerProxy) FindTeamsByUser(userID string) ([]*api.Team, error) {
	return b.store.FindTeamsByUser(userID)
}
func (b *BrokerProxy) FindTeamByName(name string) (*api.Team, error) {
	return b.store.FindTeamByName(name)
}
func (b *BrokerProxy) FindTeamByTag(tag string) (*api.Team, error) {
	return b.store.FindTeamByTag(tag)
}
func (b *BrokerProxy) FindTeamByProgram(programID string) (*api.Team, error) {
	return b.store.FindTeamByProgram(programID)
}
func (b *BrokerProxy) DeleteTeam(teamID string) error {
	err := b.store.DeleteTeam(teamID)
	go b.awakeBroker()
	return err
}
func (b *BrokerProxy) ListTeams() ([]*api.Team, error) {
	return b.store.ListTeams()
}

func (b *BrokerProxy) CreateTeamMember(teamMember api.UserTeam) (*api.UserTeam, error) {
	return b.store.CreateTeamMember(teamMember)
}
func (b *BrokerProxy) DeleteTeamMember(teamID string, userID string) error {
	return b.store.DeleteTeamMember(teamID, userID)
}
func (b *BrokerProxy) FindTeamMember(teamID string, userID string) (*api.UserTeam, error) {
	return b.store.FindTeamMember(teamID, userID)
}
func (b *BrokerProxy) UpdateTeamMember(teamMember api.UserTeam) (*api.UserTeam, error) {
	return b.store.UpdateTeamMember(teamMember)
}

func (b *BrokerProxy) UpdateRecipients(teamID string, emails []string) error {
	return b.store.UpdateRecipients(teamID, emails)
}
func (b *BrokerProxy) ListRecipients(teamID string) ([]*api.Recipient, error) {
	return b.store.ListRecipients(teamID)
}

func (b *BrokerProxy) ListAssets(teamID string, asset api.Asset) ([]*api.Asset, error) {
	return b.store.ListAssets(teamID, asset)
}
func (b *BrokerProxy) FindAsset(teamID, assetID string) (*api.Asset, error) {
	return b.store.FindAsset(teamID, assetID)
}
func (b *BrokerProxy) CreateAsset(asset api.Asset, groups []api.Group) (*api.Asset, error) {
	a, err := b.store.CreateAsset(asset, groups)
	go b.awakeBroker()
	return a, err
}
func (b *BrokerProxy) CreateAssets(assets []api.Asset, groups []api.Group, annotations []*api.AssetAnnotation) ([]api.Asset, error) {
	aa, err := b.store.CreateAssets(assets, groups, annotations)
	go b.awakeBroker()
	return aa, err
}
func (b *BrokerProxy) DeleteAsset(asset api.Asset) error {
	err := b.store.DeleteAsset(asset)
	go b.awakeBroker()
	return err

}
func (b *BrokerProxy) DeleteAllAssets(teamID string) error {
	err := b.store.DeleteAllAssets(teamID)
	go b.awakeBroker()
	return err
}
func (b *BrokerProxy) UpdateAsset(asset api.Asset) (*api.Asset, error) {
	a, err := b.store.UpdateAsset(asset)
	go b.awakeBroker()
	return a, err
}
func (b *BrokerProxy) MergeAssets(mergeOps api.AssetMergeOperations) error {
	return b.store.MergeAssets(mergeOps)
}

// Asset Annotations
func (b *BrokerProxy) ListAssetAnnotations(teamID string, assetID string) ([]*api.AssetAnnotation, error) {
	return b.store.ListAssetAnnotations(teamID, assetID)
}

func (b *BrokerProxy) CreateAssetAnnotations(teamID string, assetID string, annotations []*api.AssetAnnotation) ([]*api.AssetAnnotation, error) {
	return b.store.CreateAssetAnnotations(teamID, assetID, annotations)
}

func (b *BrokerProxy) UpdateAssetAnnotations(teamID string, assetID string, annotations []*api.AssetAnnotation) ([]*api.AssetAnnotation, error) {
	return b.store.UpdateAssetAnnotations(teamID, assetID, annotations)
}

func (b *BrokerProxy) PutAssetAnnotations(teamID string, assetID string, annotations []*api.AssetAnnotation) ([]*api.AssetAnnotation, error) {
	return b.store.PutAssetAnnotations(teamID, assetID, annotations)
}

func (b *BrokerProxy) DeleteAssetAnnotations(teamID string, assetID string, annotations []*api.AssetAnnotation) error {
	return b.store.DeleteAssetAnnotations(teamID, assetID, annotations)
}

func (b *BrokerProxy) GetAssetType(assetTypeName string) (*api.AssetType, error) {
	return b.store.GetAssetType(assetTypeName)
}

func (b *BrokerProxy) CreateGroup(group api.Group) (*api.Group, error) {
	return b.store.CreateGroup(group)
}
func (b *BrokerProxy) ListGroups(teamID, groupName string) ([]*api.Group, error) {
	return b.store.ListGroups(teamID, groupName)
}
func (b *BrokerProxy) UpdateGroup(group api.Group) (*api.Group, error) {
	return b.store.UpdateGroup(group)
}
func (b *BrokerProxy) DeleteGroup(group api.Group) error {
	return b.store.DeleteGroup(group)
}
func (b *BrokerProxy) FindGroup(group api.Group) (*api.Group, error) {
	return b.store.FindGroup(group)
}
func (b *BrokerProxy) FindGroupInfo(group api.Group) (*api.Group, error) {
	return b.store.FindGroupInfo(group)
}
func (b *BrokerProxy) DisjoinAssetsInGroups(teamID, inGroupID string, notInGroupIDs []string) ([]*api.Asset, error) {
	return b.store.DisjoinAssetsInGroups(teamID, inGroupID, notInGroupIDs)
}

func (b *BrokerProxy) CountAssetsInGroups(teamID string, groupIDs []string) (int, error) {
	return b.store.CountAssetsInGroups(teamID, groupIDs)
}

func (b *BrokerProxy) GroupAsset(assetsGroup api.AssetGroup, teamID string) (*api.AssetGroup, error) {
	return b.store.GroupAsset(assetsGroup, teamID)
}
func (b *BrokerProxy) ListAssetGroup(assetGroup api.AssetGroup, teamID string) ([]*api.AssetGroup, error) {
	return b.store.ListAssetGroup(assetGroup, teamID)
}
func (b *BrokerProxy) UngroupAssets(assetGroup api.AssetGroup, teamID string) error {
	return b.store.UngroupAssets(assetGroup, teamID)
}

func (b *BrokerProxy) ListPrograms(teamID string) ([]*api.Program, error) {
	return b.store.ListPrograms(teamID)
}
func (b *BrokerProxy) CreateProgram(program api.Program, teamID string) (*api.Program, error) {
	return b.store.CreateProgram(program, teamID)
}
func (b *BrokerProxy) FindProgram(programID string, teamID string) (*api.Program, error) {
	return b.store.FindProgram(programID, teamID)
}
func (b *BrokerProxy) UpdateProgram(program api.Program, teamID string) (*api.Program, error) {
	return b.store.UpdateProgram(program, teamID)
}
func (b *BrokerProxy) DeleteProgram(program api.Program, teamID string) error {
	return b.store.DeleteProgram(program, teamID)
}

func (b *BrokerProxy) ListPolicies(teamID string) ([]*api.Policy, error) {
	return b.store.ListPolicies(teamID)
}
func (b *BrokerProxy) CreatePolicy(policy api.Policy) (*api.Policy, error) {
	return b.store.CreatePolicy(policy)
}
func (b *BrokerProxy) FindPolicy(policyID string) (*api.Policy, error) {
	return b.store.FindPolicy(policyID)
}
func (b *BrokerProxy) UpdatePolicy(policy api.Policy) (*api.Policy, error) {
	return b.store.UpdatePolicy(policy)
}
func (b *BrokerProxy) DeletePolicy(policy api.Policy) error {
	return b.store.DeletePolicy(policy)
}

func (b *BrokerProxy) ListChecktypeSetting(policyID string) ([]*api.ChecktypeSetting, error) {
	return b.store.ListChecktypeSetting(policyID)
}
func (b *BrokerProxy) CreateChecktypeSetting(setting api.ChecktypeSetting) (*api.ChecktypeSetting, error) {
	return b.store.CreateChecktypeSetting(setting)
}
func (b *BrokerProxy) FindChecktypeSetting(checktypeSettingID string) (*api.ChecktypeSetting, error) {
	return b.store.FindChecktypeSetting(checktypeSettingID)
}
func (b *BrokerProxy) UpdateChecktypeSetting(checktypeSetting api.ChecktypeSetting) (*api.ChecktypeSetting, error) {
	return b.store.UpdateChecktypeSetting(checktypeSetting)
}
func (b *BrokerProxy) DeleteChecktypeSetting(checktypeSettingID string) error {
	return b.store.DeleteChecktypeSetting(checktypeSettingID)
}

func (b *BrokerProxy) FindGlobalProgramMetadata(programID string, teamID string) (*api.GlobalProgramsMetadata, error) {
	return b.store.FindGlobalProgramMetadata(programID, teamID)
}
func (b *BrokerProxy) UpsertGlobalProgramMetadata(teamID, program string, defaultAutosend bool, defaultDisabled bool, defaultCron string, autosend *bool, disabled *bool, cron *string) error {
	return b.store.UpsertGlobalProgramMetadata(teamID, program, defaultAutosend, defaultDisabled, defaultCron, autosend, disabled, cron)
}
func (b *BrokerProxy) DeleteProgramMetadata(program string) error {
	return b.store.DeleteProgramMetadata(program)
}

func (b *BrokerProxy) CreateFindingOverwrite(findingOverwrite api.FindingOverwrite) error {
	err := b.store.CreateFindingOverwrite(findingOverwrite)
	go b.awakeBroker()
	return err
}

func (b *BrokerProxy) ListFindingOverwrites(findingID string) ([]*api.FindingOverwrite, error) {
	return b.store.ListFindingOverwrites(findingID)
}
