/*
Copyright 2022 Adevinta
*/

package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path"

	gokitlog "github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/spf13/viper"

	"github.com/adevinta/vulcan-api/pkg/api"
	"github.com/adevinta/vulcan-api/pkg/api/store"
	"github.com/adevinta/vulcan-api/pkg/asyncapi"
	"github.com/adevinta/vulcan-api/pkg/asyncapi/kafka"
)

var description = `Bumps all the assets in the configured vulcan db to the configured kafka topic.
It locks for writing the tables: teams, assets and annotations of the vulcan-api database.`

func main() {
	flag.Usage = usage
	configFile := flag.String("c", "", "config file (default is $HOME/.vulcan-asset-bumper.toml)")
	logLevel := flag.String("l", "info", `log level (valid values: "info", "error", "warn", "debug")`)
	pageSize := flag.Int("p", 10, `page size`)
	flag.Parse()
	if *configFile == "" {
		homedir, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error getting user home directory: %v", err)
			os.Exit(1)
		}
		*configFile = path.Join(homedir, ".vulcan-asset-bumper.toml")
	}
	cfg := config{}
	viper.SetConfigFile(*configFile)
	if err := viper.ReadInConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "error reading config %s: %v", *configFile, err)
		os.Exit(1)
	}
	if err := viper.Unmarshal(&cfg); err != nil {
		fmt.Fprintf(os.Stderr, "error parsing config file %s: %v", *configFile, err)
		os.Exit(1)
	}

	storeLogger := gokitlog.NewLogfmtLogger(os.Stderr)
	optionLevel, err := parseLogLevel(*logLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing log level: %v", err)
		os.Exit(1)
	}
	storeLogger = level.NewFilter(storeLogger, optionLevel)

	l := asyncapi.LevelLogger{storeLogger}
	store, err := store.NewStore("", cfg.DB.ConnString, storeLogger, false, map[string][]string{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating the store: %v", err)
		os.Exit(1)
	}
	kcfg := cfg.Kafka
	kclient, err := kafka.NewClient(kcfg.User, kcfg.Pass, kcfg.Broker, kcfg.Topics)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating the kafka client: %v", err)
		os.Exit(1)
	}
	v := asyncapi.NewVulcan(&kclient, l)
	if err = bump(v, store, int(*pageSize), l); err != nil {
		fmt.Fprintf(os.Stderr, "error bumping assets: %v", err)
		os.Exit(1)
	}
}

func bump(v *asyncapi.Vulcan, s store.Store, psize int, logger asyncapi.LevelLogger) error {
	r, err := s.NewAssetReader(true, psize)
	if err != nil {
		return err
	}
	defer r.Close()
	logger.Infof("sending assets in batches of %d", psize)
	page := 0
	for r.Read() {
		assets := r.Assets()
		from := page*int(psize) + 1
		to := from + len(assets) - 1
		logger.Infof("sending batch of assets from %d to %d", from, to)
		for _, a := range assets {
			payload := assetToAsyncAsset(*a)
			err = v.PushAsset(payload)
			if err != nil {
				return err
			}
		}
		page++
	}
	return r.Err()
}

func usage() {
	fmt.Fprintln(os.Stderr, description)
	fmt.Fprintf(os.Stderr, "usage: %s [flags]\n", os.Args[0])
	flag.PrintDefaults()
}

type config struct {
	DB    dbConfig
	Kafka kafkaConfig `mapstructure:"kafka"`
}

// dbConfig stores the data defined in the db config section.
type dbConfig struct {
	ConnString string `mapstructure:"connection_string"`
	LogMode    bool   `mapstructure:"log_mode"`
}

// kafkaConfig stores the configuration needed to connect to a kafka topic.
type kafkaConfig struct {
	User   string
	Pass   string
	Broker string
	Topics map[string]string
}

func parseLogLevel(l string) (level.Option, error) {
	switch l {
	case "error":
		return level.AllowError(), nil
	case "warn":
		return level.AllowWarn(), nil
	case "debug":
		return level.AllowDebug(), nil
	case "info":
		return level.AllowInfo(), nil
	default:
		err := errors.New("invalid level, the valid levels are: info, error, warn, debug")
		return nil, err
	}
}

// TODO: This function is duplicated here: pkg/api/store/cdc/parser.go, we
// should find a proper package to move it so we have only one function doing
// the same thing.
func assetToAsyncAsset(a api.Asset) asyncapi.AssetPayload {
	var annotations []*asyncapi.Annotation
	for _, asset := range a.AssetAnnotations {
		annotations = append(annotations, &asyncapi.Annotation{
			Key:   asset.Key,
			Value: asset.Value,
		})
	}
	ROLFP := ""
	if a.ROLFP != nil {
		ROLFP = a.ROLFP.String()
	}
	scannable := false
	if a.Scannable != nil {
		scannable = *a.Scannable
	}
	assetType := ""
	if a.AssetType != nil {
		assetType = a.AssetType.Name
	}
	asyncAsset := asyncapi.AssetPayload{
		Id: a.ID,
		Team: &asyncapi.Team{
			Id:          a.Team.ID,
			Name:        a.Team.Name,
			Description: a.Team.Description,
			Tag:         a.Team.Tag,
		},
		Alias:       a.Alias,
		Rolfp:       ROLFP,
		Scannable:   scannable,
		AssetType:   (*asyncapi.AssetType)(&assetType),
		Identifier:  a.Identifier,
		Annotations: annotations,
	}
	return asyncAsset
}
