/*
Copyright 2021 Adevinta
*/

package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/adevinta/vulcan-api/cmd/vulcan-cli/cli"
	"github.com/spf13/cobra"
)

var (
	Assets = &cobra.Command{
		Use:   `assets <output_file>`,
		Short: `Downloads all identifiers of the assets in vulcan to a specified text file.`,
		Long: `Downloads all the assets in vulcan optionally filtering by asset type.
	The output file will composed by lines with the following format: identifier;asset_type`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAssets(args, apiClient)
		},
	}
	assetTypes             []string
	ErrOutputAlreadyExists = errors.New("output file already exists")
)

func init() {
	Assets.Flags().BoolVarP(&force, "force", "f", false, "Force overwrite the output file if it exits")
	Assets.PersistentFlags().StringArrayVarP(&assetTypes, "type", "", []string{}, "asset types to get info for, it can be used multiple times to specify more than one asset type")
	rootCmd.AddCommand(Assets)
}

func runAssets(args []string, apiClient *cli.CLI) error {
	path := args[0]
	if !force {
		_, err := os.Stat(path)
		if err != nil && errors.Is(err, fs.ErrExist) {
			return ErrOutputAlreadyExists
		}
	}

	assets, err := allAssets(apiClient, assetTypes)
	if err != nil {
		return err
	}
	fs, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer fs.Close()
	for _, a := range assets {
		line := fmt.Sprintf("%s;%s\n", a.Identifier, a.AssetType)
		fs.WriteString(line)
	}
	return nil
}

type assetInfo struct {
	Identifier string
	AssetType  string
}

func allAssets(apiClient *cli.CLI, types []string) ([]assetInfo, error) {
	var (
		teams []string
		infos = make(map[string]assetInfo)
	)

	teamsData, err := apiClient.Teams()
	if err != nil {
		return nil, err
	}
	for _, t := range teamsData {
		teams = append(teams, t.ID)
	}
	for _, t := range teams {
		assets, err := apiClient.Assets(t)
		if err != nil {
			return nil, err
		}
		for _, a := range assets {
			if len(assetTypes) > 0 && !strSliceExist(a.AssetType, assetTypes) {
				continue
			}
			id := fmt.Sprintf("%s:%s", a.AssetType, a.Target)
			if _, exists := infos[id]; !exists {
				infos[id] = assetInfo{
					Identifier: a.Target,
					AssetType:  a.AssetType,
				}
			}
		}
	}
	var assets = make([]assetInfo, 0, len(infos))
	for _, a := range infos {
		assets = append(assets, a)
	}
	return assets, nil
}

func strSliceExist(value string, in []string) bool {
	for _, v := range in {
		if v == value {
			return true
		}
	}
	return false
}
