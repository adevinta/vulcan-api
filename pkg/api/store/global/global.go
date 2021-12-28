/*
Copyright 2021 Adevinta
*/

package global

import (
	"context"

	"github.com/adevinta/errors"
	"github.com/adevinta/vulcan-api/pkg/api"
)

var (
	groups   = map[string]Group{}
	policies = map[string]Policy{}
	programs = map[string]Program{}
	reports  = map[string]Report{}
)

func registerGroup(g Group) {
	groups[g.Name()] = g
}

func registerPolicy(p Policy) {
	policies[p.Name()] = p
}

func registerProgram(p Program) {
	programs[p.ID] = p
}

func registerReport(r Report) {
	reports[r.ID] = r
}

// ChecktypesInformer defines the functions that the component providing
// checktypes info to the global policies must implement.
type ChecktypesInformer interface {
	ByAssettype(ctx context.Context) (map[string][]string, error)
}

// Entities shapes the interface exposed to other packages
// to interact with global entities.
type Entities struct {
	store    api.VulcanitoStore
	informer ChecktypesInformer
	groups   map[string]Group
	policies map[string]Policy
	programs map[string]Program
	reports  map[string]Report
}

// NewEntities returns a struct that exposes the current defined global
// entities.
func NewEntities(store api.VulcanitoStore, informer ChecktypesInformer) (*Entities, error) {
	// Initialize global entities with the required dependencies.
	globals := &Entities{
		groups:   groups,
		policies: policies,
		programs: programs,
		reports:  reports,
	}
	// Inject required dependencies in each global entity.
	for k, g := range globals.groups {
		err := g.Init(store)
		if err != nil {
			return nil, errors.Default(err)
		}
		globals.groups[k] = g
	}

	for k, p := range globals.policies {
		err := p.Init(informer)
		if err != nil {
			return nil, errors.Default(err)
		}
		globals.policies[k] = p
	}
	// By now no need to initialize scans.
	return globals, nil

}

// Groups returns the current defined current groups.
func (c *Entities) Groups() map[string]Group {
	return groups
}

// Policies returns current defined global policies.
func (c *Entities) Policies() map[string]Policy {
	return policies
}

// Programs returns current defined global programs.
func (c *Entities) Programs() map[string]Program {
	return programs
}

// Reports returns current defined global reports.
func (c *Entities) Reports() map[string]Report {
	return reports
}

// Group defines the methods all the global groups must implement.
type Group interface {
	Init(api.VulcanitoStore) error
	Name() string
	Options() string
	Description() string
	// ShadowTeamGroup must return a group name if the global group is shadowing
	// a "normal" group of a team. A shadowed group is a global group that can
	// be referenced by a global program but acts as it is effectively the real
	// group of the team.
	ShadowTeamGroup() string
	Eval(teamID string) ([]*api.Asset, error)
}

// Policy defines the shape of a global policy.
type Policy interface {
	Init(ChecktypesInformer) error
	Description() string
	Name() string
	Eval(context.Context, GlobalPolicyConfig) ([]*api.ChecktypeSetting, error)
}

// Program defines the information required to define a global
// program.
type Program struct {
	ID              string
	Name            string
	Policies        []PolicyGroup
	DefaultMetadata api.GlobalProgramsMetadata
}

type PolicyGroup struct {
	Group  string
	Policy string
}

type ChecksByAssetType struct {
	Assettype string
	Name      []string
}

// Report defines the information required to define a global
// report.
type Report struct {
	ID              string
	Name            string
	DefaultSchedule string
}

// globalGroup provides a template for implementing global groups.
// Can be used by embeding it in the struct implementing a concrete
// global group and overriding the functions as needed.
type globalGroup struct {
	Store api.VulcanitoStore
}

func (g *globalGroup) Init(store api.VulcanitoStore) error {
	g.Store = store
	return nil
}

func (g *globalGroup) Eval(teamID string) ([]*api.Asset, error) {
	return nil, errors.Default("not implemented")
}

func (g *globalGroup) ShadowTeamGroup() string {
	return ""
}

func (g *globalGroup) Description() string {
	return ""
}

func (d *globalGroup) Options() string {
	return ""
}
