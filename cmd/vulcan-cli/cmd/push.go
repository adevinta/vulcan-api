/*
Copyright 2021 Adevinta
*/

package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/adevinta/vulcan-api/cmd/vulcan-cli/cli"
)

var PushTeam = &cobra.Command{
	Use:   `push <teams_dir> <team_name>`,
	Short: `Uploads the info of a team with the one that is defined in files, overwriting the remote info`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPushTeam(args, apiClient)
	},
}

func init() {
	PushTeam.Flags().BoolVarP(&force, "force", "f", false, "Push changes to remote, otherwise performs a dry run only")

	rootCmd.AddCommand(PushTeam)
}

func runPushTeam(args []string, apiClient *cli.CLI) error {
	path := args[0]
	name := args[1]

	localTeams, err := readLocalTeams(path, name)
	if err != nil {
		return err
	}

	remoteTeams, err := readRemoteTeams(name, apiClient)
	if err != nil {
		return err
	}

	journal, err := cli.NewJournal(localTeams, remoteTeams, apiClient)
	if err != nil {
		return err
	}
	err = journal.BuildModifications()
	if err != nil {
		return err
	}
	fmt.Println(journal)

	if force {
		fmt.Println("[*] Applying")
		if err := journal.Apply(); err != nil {
			return err
		}

		fmt.Println("[*] Done. Updating local files")
		return updateLocalTeams(path, localTeams)
	}

	fmt.Println("[*] Use force option to execute it")

	return nil
}

func readLocalTeams(path, name string) ([]*cli.Team, error) {
	var teams []*cli.Team
	if name == allTeamsToken {
		t, err := cli.ReadLocalTeams(path)
		if err != nil {
			return nil, err
		}

		teams = append(teams, t...)
	} else {
		t, err := cli.ReadLocalTeam(filepath.Join(path, name))
		if err != nil {
			return nil, err
		}

		teams = append(teams, t)
	}

	return teams, nil
}

func updateLocalTeams(path string, teams []*cli.Team) error {
	for _, t := range teams {
		assets, err := apiClient.Assets(t.ID)
		if err != nil {
			return err
		}
		t.Assets = assets

		groups, err := apiClient.Groups(t.ID)
		if err != nil {
			return err
		}
		t.Groups = groups

		orphans, err := apiClient.OrphanAssets(assets, t.Groups)
		if err != nil {
			return err
		}
		t.OrphanAssets = orphans

		foreigns, err := apiClient.ForeignAssets(assets, t.Groups)
		if err != nil {
			return err
		}
		t.ForeignAssets = foreigns

		members, err := apiClient.Members(t.ID)
		if err != nil {
			return err
		}
		t.Members = members

		if err := t.WriteLocal(filepath.Join(path, t.Name)); err != nil {
			return err
		}
	}

	return nil
}
