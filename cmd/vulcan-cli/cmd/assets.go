/*
Copyright 2023 Adevinta
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
	// Assets command downloads all/some assets from vulcan into a specified file.
	Assets = &cobra.Command{
		Use:   `assets <output_file>`,
		Short: `Downloads all the assets in vulcan to a specified text file.`,
		Long: `Downloads all the assets in vulcan to a file optionally filtering by asset type.
The output file has one asset per line. Line format: identifier;asset_type`,
		Example: `vulcan-cli assets assets.txt --type Hostname --type WebAddress -H vulcan.example.com -k a_token`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAssets(args, apiClient)
		},
	}
	assetTypes             []string
	ErrOutputAlreadyExists = errors.New("output file already exists")
	assetsHelp             = "asset types to get info for, it can be used multiple times to specify filters for more than one asset type"
)

func init() {
	Assets.Flags().BoolVarP(&force, "force", "f", false, "overwrites the output file if it exits")
	Assets.PersistentFlags().StringArrayVarP(&assetTypes, "type", "", []string{}, assetsHelp)
	rootCmd.AddCommand(Assets)
}

func runAssets(args []string, apiClient *cli.CLI) error {
	path := args[0]
	// We check for the existence of the file before performing the calls to
	// the vulcan api to avoid the user to wait just to see the command fail.
	exists, err := fileExists(path)
	if err != nil {
		return err
	}
	if exists && !force {
		return ErrOutputAlreadyExists
	}
	assets, err := getAssets(apiClient, assetTypes)
	if err != nil {
		return err
	}
	// We accept that, if the destination file was created after we checked for
	// its existence above and now, the file will be overwritten.
	fs, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer fs.Close()
	for _, a := range assets {
		fmt.Fprintf(fs, "%s;%s\n", a.Identifier, a.AssetType)
	}
	return nil
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

type assetInfo struct {
	Identifier string
	AssetType  string
}

func getAssets(apiClient *cli.CLI, types []string) ([]assetInfo, error) {
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
