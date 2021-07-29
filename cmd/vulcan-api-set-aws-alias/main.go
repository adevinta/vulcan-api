/*
Copyright 2021 Adevinta
*/

package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/user"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/spf13/viper"

	"github.com/adevinta/vulcan-api/pkg/api"
	"github.com/adevinta/vulcan-api/pkg/api/store"
	"github.com/adevinta/vulcan-api/pkg/awscatalogue"
	awscatalogueclient "github.com/adevinta/vulcan-api/pkg/awscatalogue/client"
)

type dbConfig struct {
	ConnString string `mapstructure:"connection_string"`
	LogMode    bool   `mapstructure:"log_mode"`
}

type awsCatalogueConfig struct {
	Kind          string `mapstructure:"kind"`
	URL           string `mapstructure:"url"`
	Key           string `mapstructure:"key"`
	Retries       int    `mapstructure:"retries"`
	RetryInterval int    `mapstructure:"retry_interval"`
}

type config struct {
	DB           dbConfig
	AWSCatalogue awsCatalogueConfig
}

var cfgFile string

func main() {
	flag.StringVar(&cfgFile, "config", "c", "path to a config file")
	flag.Parse()
	cfg := mustInitConfig()
	var l = log.NewLogfmtLogger(os.Stderr)
	db, err := store.NewDB("postgres", cfg.DB.ConnString, l, cfg.DB.LogMode, map[string][]string{})
	if err != nil {
		err = fmt.Errorf("opening DB connection: %v", err)
		l.Log("error", err)
		os.Exit(1)
	}

	catalogueCfg := awscatalogueclient.AWSCatalogueAPIConfig{
		URL: cfg.AWSCatalogue.URL, Key: cfg.AWSCatalogue.Key,
		Retries: cfg.AWSCatalogue.Retries, RetryInterval: cfg.AWSCatalogue.RetryInterval,
	}
	cgAPI, err := awscatalogueclient.NewClient(cfg.AWSCatalogue.Kind, catalogueCfg)
	if err != nil {
		err = fmt.Errorf("creating AWS catalogue client: %v", err)
		l.Log("error", err)
		os.Exit(1)
	}
	awsAccounts := awscatalogue.NewAWSAccounts(cgAPI, l)
	err = awsAccounts.RefreshCache()
	if err != nil {
		err = fmt.Errorf("loading the aws account names cache: %w", err)
		l.Log("error", err)
		os.Exit(1)
	}

	teams, err := db.ListTeams()
	if err != nil {
		l.Log("error", err)
		os.Exit(1)
	}
	uAssets := []*api.Asset{}
	var noAliasAccounts, totalAccouts int
	for _, t := range teams {
		t := t
		assets, err := db.ListAssets(t.ID, api.Asset{})
		if err != nil {
			err = fmt.Errorf("opening DB connection: %v", err)
			l.Log("error", err)
			os.Exit(1)
		}
		for _, a := range assets {
			a := a
			if a.AssetType.Name == "AWSAccount" && a.Alias == "" {
				totalAccouts++
				id := strings.Replace(a.Identifier, "arn:aws:iam::", "", -1)
				id = strings.Replace(id, ":root", "", -1)
				id = strings.Trim(id, " ")
				alias, err := awsAccounts.Name(id)
				if err != nil {
					if errors.Is(err, awscatalogue.ErrAccountNotFound) {
						noAliasAccounts++
						continue
					}
					err = fmt.Errorf("getting account alias: %v", err)
					l.Log("error", err)
					os.Exit(1)
				}
				a.Alias = alias
				uAssets = append(uAssets, a)
			}
		}
	}

	fmt.Printf("total accounts: %d, accounts to add alias %d, accounts alias not found %d\n", totalAccouts, len(uAssets), noAliasAccounts)
	var updatedAccounts int
	for _, a := range uAssets {
		_, err = db.UpdateAsset(*a)
		if err != nil {
			err = fmt.Errorf("updating account: %v", err)
			l.Log("error", err)
			os.Exit(1)
		}
		updatedAccounts++
	}

	fmt.Printf("updated accounts %d\n", updatedAccounts)
}

func mustInitConfig() config {
	var cfg config
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		usr, err := user.Current()
		if err != nil {
			fmt.Println("Can't get current user:", err)
			os.Exit(1)
		}

		// Search config in home directory with name ".vulcan-api" (without extension).
		viper.AddConfigPath(usr.HomeDir)
		viper.SetConfigName(".vulcan-api")
	}

	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Can't read config:", err)
		os.Exit(1)
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		fmt.Println("Can't decode config:", err)
		os.Exit(1)
	}
	return cfg
}
