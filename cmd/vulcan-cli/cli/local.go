/*
Copyright 2021 Adevinta
*/

package cli

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func (t *Team) WriteLocal(teamDirectory string) error {
	groupsDirectory := filepath.Join(teamDirectory, "groups")
	if err := os.MkdirAll(groupsDirectory, os.ModeDir|0755); err != nil {
		return err
	}

	policiesDirectory := filepath.Join(teamDirectory, "policies")
	if err := os.MkdirAll(policiesDirectory, os.ModeDir|0755); err != nil {
		return err
	}

	content := fmt.Sprintf("%s\n", t.String())
	if err := ioutil.WriteFile(filepath.Join(teamDirectory, "team.txt"), []byte(content), os.ModePerm); err != nil {
		return err
	}

	if err := t.Recipients.WriteLocal(teamDirectory); err != nil {
		return err
	}

	if err := t.Members.WriteLocal(teamDirectory); err != nil {
		return err
	}

	if err := t.Groups.WriteLocal(groupsDirectory); err != nil {
		return err
	}

	if err := t.Assets.WriteLocal(teamDirectory); err != nil {
		return err
	}

	if err := t.OrphanAssets.WriteLocal(teamDirectory); err != nil {
		return err
	}

	if err := t.DuppedAssets.WriteLocal(teamDirectory); err != nil {
		return err
	}

	if err := t.ForeignAssets.WriteLocal(teamDirectory); err != nil {
		return err
	}

	if err := t.Programs.WriteLocal(teamDirectory); err != nil {
		return err
	}

	if err := t.Policies.WriteLocal(policiesDirectory); err != nil {
		return err
	}

	if err := t.Coverage.WriteLocal(teamDirectory); err != nil {
		return err
	}

	return nil
}

func ReadLocalTeams(localTeamsRootDir string) ([]*Team, error) {
	items, err := ioutil.ReadDir(localTeamsRootDir)
	if err != nil {
		return nil, err
	}

	var teams []*Team
	for _, i := range items {
		if i.IsDir() {
			t, err := ReadLocalTeam(filepath.Join(localTeamsRootDir, i.Name()))
			if err != nil {
				return nil, err
			}

			teams = append(teams, t)
		}
	}

	return teams, nil
}

func ReadLocalTeam(teamDirectory string) (*Team, error) {
	t, err := readTeamInfo(teamDirectory)
	if err != nil {
		return nil, err
	}

	recipients, err := ReadLocalRecipients(teamDirectory)
	if err != nil {
		return nil, err
	}
	t.Recipients = recipients

	members, err := ReadLocalMembers(teamDirectory)
	if err != nil {
		return nil, err
	}
	t.Members = members

	groups, err := ReadLocalGroups(filepath.Join(teamDirectory, "groups"))
	if err != nil {
		return nil, err
	}
	t.Groups = groups

	assets, err := ReadLocalAssets(teamDirectory)
	if err != nil {
		return nil, err
	}
	t.Assets = assets

	orphanAssets, err := ReadLocalOrphanAssets(teamDirectory)
	if err != nil {
		return nil, err
	}
	t.OrphanAssets = orphanAssets

	foreignAssets, err := ReadLocalForeignAssets(teamDirectory)
	if err != nil {
		return nil, err
	}
	t.ForeignAssets = foreignAssets

	duppedAssets, err := ReadLocalDuppedAssets(teamDirectory)
	if err != nil {
		return nil, err
	}
	t.DuppedAssets = duppedAssets

	policies, err := ReadLocalPolicies(filepath.Join(teamDirectory, "policies"))
	if err != nil {
		return nil, err
	}
	t.Policies = policies

	programs, err := ReadLocalPrograms(teamDirectory)
	if err != nil {
		return nil, err
	}
	t.Programs = programs

	return t, nil
}

func readTeamInfo(teamDirectory string) (*Team, error) {
	f := filepath.Join(teamDirectory, "team.txt")
	lines, err := ReadLines(f)
	if err != nil {
		return nil, err
	}
	if len(lines) != 1 {
		return nil, fmt.Errorf("invalid number of lines %v in %v", len(lines), f)
	}

	var t Team
	parts := strings.Split(lines[0], ";")
	switch len(parts) {
	case 4:
		t.Tag = parts[3]
		fallthrough
	case 3:
		t.ID = parts[2]
		fallthrough
	case 2:
		t.Name = parts[0]
		t.Description = parts[1]

		if t.Name == "" {
			return nil, fmt.Errorf("invalid line: %s", lines[0])
		}
	default:
		return nil, fmt.Errorf("invalid line: %s", lines[0])
	}

	return &t, nil
}

func (r Recipients) WriteLocal(teamDirectory string) error {
	content := r.String()
	return ioutil.WriteFile(filepath.Join(teamDirectory, "emails.txt"), []byte(content), os.ModePerm)
}

func ReadLocalRecipients(teamDirectory string) (Recipients, error) {
	lines, err := ReadLines(filepath.Join(teamDirectory, "emails.txt"))
	if err != nil {
		return nil, err
	}

	var r Recipients

	for _, line := range lines {
		r = append(r, Recipient{
			Email: line,
		})
	}

	return r, nil
}

func (u Users) WriteLocal(teamsDirectory string) error {
	content := u.String()
	return ioutil.WriteFile(filepath.Join(teamsDirectory, "users.txt"), []byte(content), os.ModePerm)
}

func (u Unassigned) WriteLocal(teamsDirectory string) error {
	content := u.String()
	return ioutil.WriteFile(filepath.Join(teamsDirectory, "unassigned.txt"), []byte(content), os.ModePerm)
}

func (m Members) WriteLocal(teamDirectory string) error {
	content := m.String()
	return ioutil.WriteFile(filepath.Join(teamDirectory, "members.txt"), []byte(content), os.ModePerm)
}

func ReadLocalMembers(teamDirectory string) (Members, error) {
	lines, err := ReadLines(filepath.Join(teamDirectory, "members.txt"))
	if err != nil {
		return nil, err
	}

	var members Members
	for _, line := range lines {
		m, err := parseMember(line)
		if err != nil {
			return Members{}, err
		}

		members = append(members, m)
	}

	return members, nil
}

func (a Assets) WriteLocal(teamDirectory string) error {
	content := a.String()
	return ioutil.WriteFile(filepath.Join(teamDirectory, "assets.txt"), []byte(content), os.ModePerm)
}

func ReadLocalAssets(teamDirectory string) (Assets, error) {
	lines, err := ReadLines(filepath.Join(teamDirectory, "assets.txt"))
	if err != nil {
		return nil, err
	}

	var assets Assets
	for _, line := range lines {
		a, err := parseAsset(line)
		if err != nil {
			return Assets{}, err
		}

		assets = append(assets, a)
	}

	return assets, nil
}

func (o OrphanAssets) WriteLocal(teamDirectory string) error {
	content := o.String()
	return ioutil.WriteFile(filepath.Join(teamDirectory, "orphan.txt"), []byte(content), os.ModePerm)
}

func ReadLocalOrphanAssets(teamDirectory string) (OrphanAssets, error) {
	lines, err := ReadLines(filepath.Join(teamDirectory, "orphan.txt"))
	if err != nil {
		return OrphanAssets{}, err
	}

	var orphans OrphanAssets
	for _, line := range lines {
		a, err := parseAsset(line)
		if err != nil {
			return OrphanAssets{}, err
		}

		orphans.Assets = append(orphans.Assets, a)
	}

	return orphans, nil
}

func (o ForeignAssets) WriteLocal(teamDirectory string) error {
	content := o.String()
	return ioutil.WriteFile(filepath.Join(teamDirectory, "foreign.txt"), []byte(content), os.ModePerm)
}

func ReadLocalForeignAssets(teamDirectory string) (ForeignAssets, error) {
	lines, err := ReadLines(filepath.Join(teamDirectory, "foreign.txt"))
	if err != nil {
		return ForeignAssets{}, err
	}

	var foreigns ForeignAssets
	for _, line := range lines {
		a, err := parseAsset(line)
		if err != nil {
			return ForeignAssets{}, err
		}

		foreigns.Assets = append(foreigns.Assets, a)
	}

	return foreigns, nil
}

func (d DuppedAssets) WriteLocal(teamDirectory string) error {
	content := d.String()
	return ioutil.WriteFile(filepath.Join(teamDirectory, "dupped.txt"), []byte(content), os.ModePerm)
}

func ReadLocalDuppedAssets(teamDirectory string) (DuppedAssets, error) {
	lines, err := ReadLines(filepath.Join(teamDirectory, "dupped.txt"))
	if err != nil {
		return DuppedAssets{}, err
	}

	var dupped DuppedAssets
	for _, line := range lines {
		a, err := parseAsset(line)
		if err != nil {
			return DuppedAssets{}, err
		}

		dupped.Assets = append(dupped.Assets, a)
	}

	return dupped, nil
}

func (g *Group) WriteLocal(groupsDirectory string) error {
	content := g.Assets.String()
	return ioutil.WriteFile(filepath.Join(groupsDirectory, fmt.Sprintf("%s;%s", g.Name, g.ID)), []byte(content), os.ModePerm)
}

func (g Groups) WriteLocal(groupsDirectory string) error {
	for _, group := range g {
		if err := group.WriteLocal(groupsDirectory); err != nil {
			return err
		}
	}
	return nil
}

func ReadLocalGroups(groupsDirectory string) (Groups, error) {
	files, err := ioutil.ReadDir(groupsDirectory)
	if err != nil {
		return Groups{}, err
	}

	var groups Groups
	for _, f := range files {
		name := f.Name()

		g, err := ReadLocalGroup(filepath.Join(groupsDirectory, name))
		if err != nil {
			return Groups{}, err
		}

		groups = append(groups, g)
	}

	return groups, nil
}

func ReadLocalGroup(path string) (*Group, error) {
	name := filepath.Base(path)
	parts := strings.Split(name, ";")
	var g Group
	switch len(parts) {
	case 2:
		g.ID = parts[1]
		fallthrough
	case 1:
		g.Name = parts[0]
		if g.Name == "" {
			return nil, fmt.Errorf("invalid group name: %s", name)
		}
	default:
		return nil, fmt.Errorf("invalid group name: %s", name)
	}

	lines, err := ReadLines(path)
	if err != nil {
		return nil, err
	}

	for _, line := range lines {
		// Skip empty lines.
		if line == "" {
			continue
		}
		a, err := parseAsset(line)
		if err != nil {
			return nil, err
		}
		g.Assets = append(g.Assets, a)
	}

	return &g, nil
}

func parseAsset(line string) (*Asset, error) {
	parts := strings.Split(line, ";")
	// if line == "arn:aws:iam::344038293544:root;AWSAccount;R:1/O:1/L:1/F:1/P:1+S:2;;" {
	// 	fmt.Printf("PARTS %+v", parts)
	// }
	fmt.Printf("PARTS %+v\n", parts)
	var a Asset
	switch len(parts) {
	case 5:
		// Existing asset with or without rolfp.
		// e.g.: example3.vulcan.example.com;Hostname;R:0/O:1/L:0/F:1/P:0+S:2;2abfe9a3-78ad-4e6b-a82e-8de7e40a3c58
		// e.g.: arn:aws:iam::239557989611:root;AWSAccount;;fe004598-ba75-4be6-89...
		// arn:aws:iam::581382904508:root;AWSAccount;R:1/O:1/L:1/F:1/P:1+S:2;SPT Infrastructure Public (DEV);308262f3-f681-49b8-8210-321726273957
		a.ID = parts[4]
		a.Alias = parts[3]
		// If the assets does not have rolfp parts[2] will empty.
		a.Rolfp = parts[2]
		a.AssetType = parts[1]
		a.Target = parts[0]
	case 4:
		// New asset with rolfp and alias e.g. example3.vulcan.example.com;Hostname;R:0/O:1/L:0/F:1/P:0+S:2;alias
		a.Alias = parts[3]
		a.Rolfp = parts[2]
		a.AssetType = parts[1]
		a.Target = parts[0]
	case 3:
		// New asset with rolfp but without alias.
		// exmpale3.vulcan.example.com;Hostname
		a.Target = parts[0]
		a.Rolfp = parts[2]
		a.AssetType = parts[1]
		if a.Target == "" || a.AssetType == "" {
			return nil, fmt.Errorf("invalid line: %s", line)
		}
	case 2:
		// New asset without rolfp and without alias
		// exmpale3.vulcan.example.com;Hostname
		a.Target = parts[0]
		a.AssetType = parts[1]
		if a.Target == "" || a.AssetType == "" {
			return nil, fmt.Errorf("invalid line: %s", line)
		}
	case 1:
		a.Target = parts[0]
		if a.Target == "" {
			return nil, fmt.Errorf("invalid line: %s", line)
		}
	default:
		return nil, fmt.Errorf("invalid line: %s", line)
	}
	if a.ID == "" {
		fmt.Printf("NEW ASSET TO CREATE READ %+v\n", a)
	}
	return &a, nil
}

func parseMember(line string) (*Member, error) {
	parts := strings.Split(line, ";")
	var m Member
	switch len(parts) {
	case 3:
		m.ID = parts[2]
		fallthrough
	case 2:
		m.Role = parts[1]
		fallthrough
	case 1:
		m.Email = parts[0]
	default:
		return nil, fmt.Errorf("invalid line: %s", line)
	}

	if m.Email == "" {
		return nil, fmt.Errorf("invalid line: %s", line)
	}
	if m.Role == "" {
		m.Role = "member"
	}

	return &m, nil
}

func (p *Policy) WriteLocal(policiesDirectory string) error {
	dir := filepath.Join(policiesDirectory, fmt.Sprintf("%s;%s", p.Name, p.ID))
	if err := os.MkdirAll(dir, os.ModeDir|0755); err != nil {
		return err
	}

	if err := p.Settings.WriteLocal(filepath.Join(dir, "settings")); err != nil {
		return err
	}

	return p.Programs.WriteLocal(dir)
}

func (p Policies) WriteLocal(policiesDirectory string) error {
	for _, policy := range p {
		if err := policy.WriteLocal(policiesDirectory); err != nil {
			return err
		}
	}
	return nil
}

func ReadLocalPolicies(policiesDirectory string) (Policies, error) {
	files, err := ioutil.ReadDir(policiesDirectory)
	if err != nil {
		return Policies{}, err
	}

	var policies Policies
	for _, f := range files {
		name := f.Name()

		p, err := ReadLocalPolicy(filepath.Join(policiesDirectory, name))
		if err != nil {
			return Policies{}, err
		}

		policies = append(policies, p)
	}

	return policies, nil
}

func ReadLocalPrograms(teamDirectory string) (Programs, error) {
	path := filepath.Join(teamDirectory, "programs.txt")
	var programs Programs
	lines, err := ReadLines(path)
	if err != nil {
		return nil, err
	}
	for _, l := range lines {
		var p Program
		parts := strings.Split(l, ";")
		switch len(parts) {
		case 5:
			p.ID = parts[4]
			fallthrough
		case 4:
			cron := parts[3]
			p.Cron = cron
			autosend, err := strconv.ParseBool(parts[2])
			if err != nil {
				return nil, err
			}
			p.Autosend = autosend
			// TODO load policy groups.
			p.Name = parts[0]
		}
		programs = append(programs, &p)
	}

	return programs, nil
}

func ReadLocalPolicy(path string) (*Policy, error) {
	name := filepath.Base(path)
	parts := strings.Split(name, ";")
	var p Policy
	switch len(parts) {
	case 2:
		p.ID = parts[1]
		fallthrough
	case 1:
		p.Name = parts[0]
		if p.Name == "" {
			return nil, fmt.Errorf("invalid policy name: %s", name)
		}
	default:
		return nil, fmt.Errorf("invalid policy name: %s", name)
	}

	settings, err := ReadLocalSettingsCollection(filepath.Join(path, "settings"))
	if err != nil {
		return nil, err
	}
	p.Settings = settings

	return &p, nil
}

func ReadLocalSettingsCollection(settingsDirectory string) (SettingsCollection, error) {
	files, err := ioutil.ReadDir(settingsDirectory)
	if err != nil {
		return SettingsCollection{}, err
	}

	var settings SettingsCollection
	for _, f := range files {
		name := f.Name()

		s, err := ReadLocalSettings(filepath.Join(settingsDirectory, name))
		if err != nil {
			return SettingsCollection{}, err
		}

		settings = append(settings, s)
	}

	return settings, nil
}

func ReadLocalSettings(path string) (*Settings, error) {
	name := filepath.Base(path)
	parts := strings.Split(name, ";")
	var s Settings
	switch len(parts) {
	case 2:
		s.ID = parts[1]
		fallthrough
	case 1:
		s.Name = parts[0]
		if s.Name == "" {
			return nil, fmt.Errorf("invalid settings name: %s", name)
		}
	default:
		return nil, fmt.Errorf("invalid settings name: %s", name)
	}

	b, err := ioutil.ReadFile(path) //nolint
	if err != nil {
		return nil, err
	}
	s.Options = string(b)

	return &s, nil
}

func (s *Settings) WriteLocal(settingsDirectory string) error {
	file := filepath.Join(settingsDirectory, fmt.Sprintf("%s;%s", s.Name, s.ID))
	content := s.Options
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}

	return ioutil.WriteFile(file, []byte(content), os.ModePerm)
}

func (s SettingsCollection) WriteLocal(settingsDirectory string) error {
	if err := os.MkdirAll(settingsDirectory, os.ModeDir|0755); err != nil {
		return err
	}
	for _, settings := range s {
		if err := settings.WriteLocal(settingsDirectory); err != nil {
			return err
		}
	}
	return nil
}

func (p Programs) WriteLocal(directory string) error {
	content := p.String()
	return ioutil.WriteFile(filepath.Join(directory, "programs.txt"), []byte(content), os.ModePerm)
}

func (c Coverage) WriteLocal(directory string) error {
	content := fmt.Sprintf("%s\n", c.String())
	return ioutil.WriteFile(filepath.Join(directory, "coverage.txt"), []byte(content), os.ModePerm)
}
