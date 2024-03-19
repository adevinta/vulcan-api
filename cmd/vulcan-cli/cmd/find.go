/*
Copyright 2021 Adevinta
*/

package cmd

import (
	"fmt"
	"os"

	"github.com/adevinta/vulcan-api/cmd/vulcan-cli/cli"
	"github.com/spf13/cobra"
)

var FindTeam = &cobra.Command{
	Use:   `find <teams_dir> <asset1> [asset2 ...]`,
	Short: `Find the teams an asset pertains to in the local teams directory`,
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runFindTeam(args)
	},
}

func init() {
	rootCmd.AddCommand(FindTeam)
}

func runFindTeam(args []string) error {
	path := args[0]
	assets := args[1:]

	teams, err := cli.ReadLocalTeams(path)
	if err != nil {
		return err
	}

	for _, a := range assets {
		found := false
		for _, t := range teams {
			_, ok := t.Assets.FindByTarget(a)
			if ok {
				found = true
				fmt.Fprintf(os.Stderr, "[*] Found %s: '%s'\n", a, t.Name)
				fmt.Printf("%s;%s\n", a, t.Name)
			}
		}
		if !found {
			fmt.Fprintf(os.Stderr, "[*] NOT Found '%s'\n", a)
		}
	}

	return nil
}
