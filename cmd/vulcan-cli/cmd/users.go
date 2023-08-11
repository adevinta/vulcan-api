/*
Copyright 2023 Adevinta
*/

package cmd

import (
	"fmt"
	"os"
	"slices"

	"github.com/spf13/cobra"

	"github.com/adevinta/vulcan-api/cmd/vulcan-cli/cli"
)

var users = &cobra.Command{
	Use:   "users <output_file>",
	Short: "Writes the list of the emails corresponding to all the users and recipients in Vulcan to the specified output file.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runUsers(args[0], apiClient)
	},
}

func init() {
	users.Flags().BoolVarP(&force, "force", "f", false, "overwrites the output file if it exits")
	rootCmd.AddCommand(users)
}

func runUsers(outputFile string, apiCLI *cli.CLI) error {
	if !force {
		// Check for the existence of the file before performing the calls to
		// the vulcan api to avoid the user to wait just to see the command fail.
		exists, err := fileExists(outputFile)
		if err != nil {
			return err
		}
		if exists {
			return ErrOutputAlreadyExists
		}
	}

	emails := map[string]struct{}{}
	teams, err := apiCLI.Teams()
	if err != nil {
		return fmt.Errorf("error retrieving teams: %w", err)
	}
	for _, t := range teams {
		members, err := apiCLI.Members(t.ID)
		if err != nil {
			return fmt.Errorf("error retrieving members of the team %s: %w", t.ID, err)
		}
		for _, m := range members {
			emails[m.Email] = struct{}{}
		}
		recipients, err := apiCLI.Recipients(t.ID)
		if err != nil {
			return fmt.Errorf("error retrieving recipients of the team %s: %w", t.ID, err)
		}
		for _, r := range recipients {
			emails[r.Email] = struct{}{}
		}
	}
	var list []string
	for email := range emails {
		list = append(list, email)
	}
	// Sort the output so its easier to visually compare results between
	// executions.
	slices.Sort(list)
	// We accept that if the output file was created after we checked for
	// its existence above and now the file will be overwritten.
	fs, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer fs.Close()
	for _, email := range list {
		if _, err := fmt.Fprintln(fs, email); err != nil {
			return fmt.Errorf("error writing to file %w", err)
		}
	}
	return nil
}
