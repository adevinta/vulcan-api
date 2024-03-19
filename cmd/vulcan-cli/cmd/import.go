/*
Copyright 2021 Adevinta
*/

package cmd

import (
	"github.com/adevinta/vulcan-api/cmd/vulcan-cli/cli"
	"github.com/spf13/cobra"
)

var ImportCmd = &cobra.Command{
	Use:   `import <vulcanito_assets_dir>`,
	Short: `Import assets in Vulcanito format into vulcan-api`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runImport(args, apiClient)
	},
}

func init() {
	rootCmd.AddCommand(ImportCmd)
}

func runImport(args []string, apiClient *cli.CLI) error {
	teams, err := cli.ReadVulcanitoTeams(args[0])
	if err != nil {
		return err
	}

	for _, t := range teams {
		id, err := apiClient.CreateTeam(t)
		if err != nil {
			return err
		}

		if err := apiClient.AddRecipients(id, t.Recipients); err != nil { // nolint
			return err
		}

		defaultGroupID, err := apiClient.CreateGroup(id, "Default")
		if err != nil {
			return err
		}

		sensitiveGroupID, err := apiClient.CreateGroup(id, "Sensitive")
		if err != nil {
			return err
		}

		for _, ac := range t.Collections {
			for _, asset := range ac.Assets {
				createdAssets, err := apiClient.CreateAsset(id, asset.Target, ac.AssetType, asset.Rolfp, asset.Alias)
				if err != nil {
					return err
				}

				group := defaultGroupID
				if asset.Sensitive {
					group = sensitiveGroupID
				}
				for _, a := range createdAssets {
					err = apiClient.AssociateAsset(id, group, a.ID)
				}
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
