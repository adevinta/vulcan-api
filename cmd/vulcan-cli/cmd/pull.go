/*
Copyright 2021 Adevinta
*/

package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/adevinta/vulcan-api/cmd/vulcan-cli/cli"
)

var PullTeam = &cobra.Command{
	Use:   `pull <teams_dir> <team_name>`,
	Short: `Dowloads all the info of a team(s) into files inside a directory, overwriting the local info`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPullTeam(args, apiClient)
	},
}

func init() {
	PullTeam.Flags().BoolVarP(&force, "force", "f", false, "Force overwrite of local team, otherwise pull is aborted if team exists locally")

	rootCmd.AddCommand(PullTeam)
}

func runPullTeam(args []string, apiClient *cli.CLI) error {
	path := args[0]
	name := args[1]

	teams, err := readRemoteTeams(name, apiClient)
	if err != nil {
		return err
	}

	for _, t := range teams {
		teamDir, err := createTeamDirectory(path, t.Name, force) //nolint
		if err != nil {
			return err
		}

		if err := t.WriteLocal(teamDir); err != nil {
			return err
		}

		if len(t.OrphanAssets.Assets) > 0 {
			fmt.Printf("[*] WARNING - The team '%v' has orphan assets\n", t.Name)
		}

		if len(t.ForeignAssets.Assets) > 0 {
			fmt.Printf("[*] WARNING - The team '%v' has foreign assets\n", t.Name)
		}

		if len(t.DuppedAssets.Assets) > 0 {
			fmt.Printf("[*] WARNING - The team '%v' has dupped assets\n", t.Name)
		}
	}

	users, err := apiClient.Users()
	if err != nil {
		return err
	}

	if err := users.WriteLocal(path); err != nil {
		return err
	}

	unassigned, err := apiClient.Unassigned(users, teams)
	if err != nil {
		return err
	}

	if err := unassigned.WriteLocal(path); err != nil {
		return err
	}

	if len(unassigned.Users) > 0 {
		fmt.Printf("[*] WARNING - There are unassigned users\n")
	}

	return nil
}

func readRemoteTeams(name string, apiClient *cli.CLI) ([]*cli.Team, error) {
	var teams []*cli.Team
	if name == allTeamsToken {
		t, err := apiClient.Teams()
		if err != nil {
			return nil, err
		}

		teams = append(teams, t...)
	} else {
		t, err := apiClient.TeamByName(name)
		if err != nil {
			return nil, err
		}

		teams = append(teams, t)
	}

	for i, t := range teams {
		recipients, err := apiClient.Recipients(t.ID)
		if err != nil {
			return nil, err
		}
		t.Recipients = recipients

		members, err := apiClient.Members(t.ID)
		if err != nil {
			return nil, err
		}
		t.Members = members

		groups, err := apiClient.Groups(t.ID)
		if err != nil {
			return nil, err
		}
		t.Groups = groups

		assets, err := apiClient.Assets(t.ID)
		if err != nil {
			return nil, err
		}
		t.Assets = assets

		orphans, err := apiClient.OrphanAssets(assets, groups)
		if err != nil {
			return nil, err
		}
		t.OrphanAssets = orphans

		foreigns, err := apiClient.ForeignAssets(assets, groups)
		if err != nil {
			return nil, err
		}
		t.ForeignAssets = foreigns

		dupped, err := apiClient.DuppedAssets(assets)
		if err != nil {
			return nil, err
		}
		t.DuppedAssets = dupped

		policies, err := apiClient.Policies(t.ID)
		if err != nil {
			return nil, err
		}
		t.Policies = policies

		programs, err := apiClient.Programs(t.ID)
		if err != nil {
			return nil, err
		}
		t.Programs = programs

		apiClient.AddProgramsToPolicies(programs, policies)

		coverage, err := apiClient.Coverage(t.ID)
		if err != nil {
			return nil, err
		}
		t.Coverage = coverage

		teams[i] = t
	}

	return teams, nil
}

func createTeamDirectory(path, name string, force bool) (string, error) {
	path = filepath.Join(path, name)
	path, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	_, err = os.Stat(path)
	switch {
	case err == nil:
		if !force {
			return "", fmt.Errorf("path %v already exists and force option is false", path)
		}
		if err := os.RemoveAll(path); err != nil { //nolint
			return "", err
		}
		fallthrough
	case os.IsNotExist(err):
		if err := os.Mkdir(path, os.ModePerm); err != nil {
			return "", err
		}
	case err != nil:
		return "", err
	}

	return path, nil
}
