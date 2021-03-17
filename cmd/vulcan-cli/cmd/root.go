/*
Copyright 2021 Adevinta
*/

package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/adevinta/vulcan-api/cmd/vulcan-cli/cli"
	"github.com/spf13/cobra"
)

const (
	allTeamsToken  = "all"
	defaultKeyFile = ".vulcan-api-token"
)

var (
	apiClient *cli.CLI
	cliCfg    cli.Config
	force     bool
)

var rootCmd = &cobra.Command{
	Use:   "vulcan-cli",
	Short: `Rich CLI to interact with vulcan-api`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		logger := log.New(os.Stderr, "", log.LstdFlags)

		if cliCfg.Key == "" {
			var err error
			cliCfg.Key, err = readDefaultKey()
			if err != nil {
				return fmt.Errorf("key not provided and can not be read from default location: %v", err)
			}
		}

		apiClient = cli.NewCLI(context.Background(), cliCfg, logger)

		return nil
	},
}

func init() {
	// Register signer flags
	rootCmd.PersistentFlags().StringVarP(&cliCfg.Key, "key", "k", "", fmt.Sprintf("API key used for authentication (default is read from $HOME/%s)", defaultKeyFile))
	rootCmd.PersistentFlags().StringVar(&cliCfg.Format, "format", "Bearer %s", "Format used to create auth header or query from key")
	// Register global flags
	rootCmd.PersistentFlags().StringVarP(&cliCfg.Scheme, "scheme", "s", "https", "Set the requests scheme")
	rootCmd.PersistentFlags().StringVarP(&cliCfg.Host, "host", "H", "www.vulcan.example.com", "API hostname")
	rootCmd.PersistentFlags().DurationVarP(&cliCfg.Timeout, "timeout", "t", time.Duration(20)*time.Second, "Set the request timeout")
	rootCmd.PersistentFlags().BoolVar(&cliCfg.Dump, "dump", false, "Dump HTTP request and response.")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func readDefaultKey() (string, error) {
	// Find home directory.
	usr, err := user.Current()
	if err != nil {
		return "", err
	}

	path := filepath.Join(usr.HomeDir, defaultKeyFile)

	k, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(k)), nil
}
