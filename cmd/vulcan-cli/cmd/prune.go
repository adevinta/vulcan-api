/*
Copyright 2021 Adevinta
*/

package cmd

import (
	"fmt"

	"github.com/adevinta/vulcan-api/cmd/vulcan-cli/cli"
	"github.com/spf13/cobra"
)

var PruneCmd = &cobra.Command{
	Use:   `prune <teams_dir> <team_name>`,
	Short: `Deletes the orphan assets defined in the local file orphan.txt for a team`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPrune(args, apiClient)
	},
}

func init() {
	PruneCmd.Flags().BoolVarP(&force, "force", "f", false, "Applies the deletes operations to remote, otherwise performs a dry run only")

	rootCmd.AddCommand(PruneCmd)
}

func runPrune(args []string, apiClient *cli.CLI) error {
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

	journal.BuildPruneModifications()
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
