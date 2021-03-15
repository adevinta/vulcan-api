/*
Copyright 2021 Adevinta
*/

package cli

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Info struct {
	Description string
	Tag         string
}

type Team struct {
	Info

	ID            string
	Name          string
	Recipients    Recipients
	Members       Members
	Groups        Groups
	Collections   []AssetsByType
	OrphanAssets  OrphanAssets
	ForeignAssets ForeignAssets
	Assets        Assets
	DuppedAssets  DuppedAssets
	Policies      Policies
	Programs      Programs
	Coverage      Coverage
}

func (t *Team) String() string {
	return fmt.Sprintf("%s;%s;%s;%s", t.Name, t.Description, t.ID, t.Tag)
}

type Recipient struct {
	Email string
}

type Recipients []Recipient

func (r Recipients) String() string {
	var b strings.Builder
	for _, recipient := range r {
		fmt.Fprintf(&b, "%s\n", recipient.Email) //nolint
	}
	return b.String()
}

type User struct {
	ID        string
	Firstname string
	Lastname  string
	Email     string
	Admin     bool
	Observer  bool
	Active    bool
}

func (u *User) String() string {
	return fmt.Sprintf("%s;%t;%t;%t;%s;%s;%s", u.Email, u.Admin, u.Active, u.Observer, u.Firstname, u.Lastname, u.ID)
}

type Users []*User

func (u Users) String() string {
	var b strings.Builder
	for _, user := range u {
		fmt.Fprintf(&b, "%s\n", user.String())
	}
	return b.String()
}

type Unassigned struct {
	Users
}

type Member struct {
	User

	Role string
}

func (m *Member) String() string {
	return fmt.Sprintf("%s;%s;%s", m.Email, m.Role, m.ID)
}

type Members []*Member

func (m Members) String() string {
	var b strings.Builder
	for _, member := range m {
		fmt.Fprintf(&b, "%s\n", member.String()) //nolint
	}
	return b.String()
}

func (ms Members) Find(email string) (*Member, bool) {
	for _, m := range ms {
		if m.Email == email {
			return m, true
		}
	}

	return nil, false
}

type Asset struct {
	ID        string
	Target    string
	AssetType string
	Sensitive bool
	Rolfp     string
	Alias     string
}

func (a *Asset) String() string {
	return fmt.Sprintf("%s;%s;%s;%s;%s", a.Target, a.AssetType, a.Rolfp, a.Alias, a.ID)
}

type Assets []*Asset

func (a Assets) String() string {
	var b strings.Builder
	for _, asset := range a {
		fmt.Fprintf(&b, "%s\n", asset.String()) //nolint
	}
	return b.String()
}

func (as Assets) Find(ID string) (*Asset, bool) {
	for _, a := range as {
		if a.ID == ID {
			return a, true
		}
	}

	return nil, false
}

func (as Assets) FindByTarget(target string) (*Asset, bool) {
	for _, a := range as {
		if a.Target == target {
			return a, true
		}
	}

	return nil, false
}

func (as Assets) IsDupped(target, assetType, id string) bool {
	for _, a := range as {
		if a.Target == target && a.AssetType == assetType && a.ID != id {
			return true
		}
	}

	return false
}

type OrphanAssets struct {
	Assets
}

type ForeignAssets struct {
	Assets
}

type DuppedAssets struct {
	Assets
}

type Group struct {
	ID     string
	Name   string
	Assets Assets
}

func (g *Group) FindAssetByID(ID string) (*Asset, bool) {
	for _, a := range g.Assets {
		if a.ID == ID {
			return a, true
		}
	}

	return nil, false
}

type Groups []*Group

func (gs Groups) FindByName(name string) (*Group, bool) {
	for _, g := range gs {
		if g.Name == name {
			return g, true
		}
	}

	return nil, false
}

type AssetsByType struct {
	AssetType string
	Assets    []*Asset
}

type Policy struct {
	ID       string
	Name     string
	Settings SettingsCollection
	Programs Programs
}

type Policies []*Policy

func (p Policies) Find(ID string) (*Policy, bool) {
	for _, policy := range p {
		if policy.ID == ID {
			return policy, true
		}
	}

	return nil, false
}

type Settings struct {
	ID      string
	Name    string
	Options string
}

type SettingsCollection []*Settings

type Program struct {
	ID           string
	Name         string
	Cron         string
	Autosend     bool
	PolicyGroups []PolicyGroup
}

type PolicyGroup struct {
	GroupID  string `json:"group_id,omitempty"`
	PolicyID string `json:"policy_id,omitempty"`
}

func (p *Program) String() string {
	pgroups, err := json.Marshal(p.PolicyGroups)
	if err != nil {
		panic(fmt.Sprintf("printing program as string, error:%+v", err))
	}
	return fmt.Sprintf("%s;%s;%t;%s;%s", p.Name, pgroups, p.Autosend, p.Cron, p.ID)
}

type Programs []*Program

func (p Programs) String() string {
	var b strings.Builder
	for _, program := range p {
		fmt.Fprintf(&b, "%s\n", program.String()) //nolint
	}
	return b.String()
}

func (p Programs) Find(ID string) (*Program, bool) {
	for _, program := range p {
		program := program
		if program.ID == ID {
			return program, true
		}
	}
	return nil, false
}

type Coverage float64

func (c Coverage) String() string {
	return fmt.Sprintf("%.2f", c)
}

type Scan struct {
	ID      string
	Status  string
	Team    string
	Program string
}

func ParseScan(s string) (*Scan, error) {
	parts := strings.Split(s, ";")

	if len(parts) == 3 && strings.Contains(parts[1], "ERROR") {
		return nil, fmt.Errorf("scan was not created correctly: %s", s)
	}

	if len(parts) != 4 {
		return nil, fmt.Errorf("malformed line: %s", s)
	}

	return &Scan{
		Program: parts[0],
		Status:  parts[1],
		Team:    parts[2],
		ID:      parts[3],
	}, nil
}

func ParseScans(path string) (Scans, error) {
	lines, err := ReadLines(path)
	if err != nil {
		return nil, err
	}

	var scans Scans
	for _, l := range lines {
		scan, err := ParseScan(l)
		if err != nil {
			return nil, err
		}

		scans = append(scans, scan)
	}

	return scans, nil
}

func (s *Scan) String() string {
	return fmt.Sprintf("%s;%s;%s;%s", s.Program, s.Status, s.Team, s.ID)
}

type Scans []*Scan

func (s Scans) String() string {
	var b strings.Builder
	for _, scan := range s {
		fmt.Fprintf(&b, "%s\n", scan.String()) //nolint
	}
	return b.String()
}

type Finding struct {
	Summary   string
	Score     float64
	Target    string
	Checktype string
}
