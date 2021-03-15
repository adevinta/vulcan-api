/*
Copyright 2021 Adevinta
*/

package cli

import (
	"fmt"
	"path/filepath"
	"strings"
)

func ReadVulcanitoTeams(vulcanitoTeamsRootDir string) ([]*Team, error) {
	lines, err := ReadLines(filepath.Join(vulcanitoTeamsRootDir, "index.txt"))
	if err != nil {
		return nil, err
	}

	var teams []*Team
	for _, l := range lines {
		if l == "" {
			continue
		}

		s := strings.Split(l, ";")
		if len(s) != 2 {
			return nil, fmt.Errorf("invalid line: %v", l)
		}

		collections, err := ReadVulcanitoCollections(filepath.Join(vulcanitoTeamsRootDir, s[1]))
		if err != nil {
			return nil, err
		}

		recipients, err := ReadVulcanitoRecipients(filepath.Join(vulcanitoTeamsRootDir, s[1]))
		if err != nil {
			return nil, err
		}

		t := &Team{
			Name:        s[0],
			Collections: collections,
			Recipients:  recipients,
		}
		teams = append(teams, t)
	}

	return teams, nil
}

func ReadVulcanitoCollections(vulcanitoTeamDir string) ([]AssetsByType, error) {
	m := map[string]string{
		"IP":         "ips.txt",
		"DomainName": "domains.txt",
		"Hostname":   "hostnames.txt",
		"AWSAccount": "aws.txt",
	}

	nessus, err := ReadLines(filepath.Join(vulcanitoTeamDir, "nessus.txt"))
	if err != nil {
		return nil, err
	}

	n := make(map[string]bool)
	for _, h := range nessus {
		n[h] = true
	}

	var collections []AssetsByType
	for k, v := range m {
		targets, err := ReadLines(filepath.Join(vulcanitoTeamDir, v))
		if err != nil {
			return nil, err
		}

		var assets []*Asset
		for _, t := range targets {
			if t == "" {
				continue
			}

			a := &Asset{Target: t}

			if k == "Hostname" {
				a.Sensitive = !n[t]
			}

			assets = append(assets, a)
		}

		ac := AssetsByType{
			AssetType: k,
			Assets:    assets,
		}
		collections = append(collections, ac)
	}

	return collections, nil
}

func ReadVulcanitoRecipients(vulcanitoTeamDir string) (Recipients, error) {
	var r Recipients
	lines, err := ReadLines(filepath.Join(vulcanitoTeamDir, "emails.txt"))
	if err != nil {
		return Recipients{}, err
	}

	for _, line := range lines {
		r = append(r, Recipient{
			Email: line,
		})
	}

	return r, nil
}
