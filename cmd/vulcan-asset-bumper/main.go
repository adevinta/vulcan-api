/*
Copyright 2022 Adevinta
*/

package main

import (
	"flag"
	"fmt"
	"os"

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
	configFile := flag.String("c", "", "path the config file")
	logLevel := flag.String("l", "info", `log level valid values: "info", "error", "warn" "debug"`)
	pageSize := flag.Uint("p", 10, `page size`)
	flag.Parse()
	if *configFile == "" {
		flag.Usage()
		os.Exit(1)
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
	storeLogger = level.NewFilter(storeLogger, parseLogLevel(*logLevel))

	l := levelLogger{storeLogger}
	store, err := store.NewStore("", cfg.DB.ConnString, storeLogger, false, map[string][]string{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating the store: %v", err)
		os.Exit(1)
	}
	kcfg := cfg.Kafka
	kclient, err := kafka.NewClient(kcfg.User, kcfg.Pass, kcfg.Broker, kcfg.Topics)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating the store: %v", err)
		os.Exit(1)
	}
	v := asyncapi.NewVulcan(&kclient, l)
	err = bump(v, store, *pageSize, l)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error bumping assets: %v", err)
		os.Exit(1)
	}
}

func bump(v asyncapi.Vulcan, s store.Store, psize uint, logger levelLogger) error {
	r, err := s.NewAssetReader(true, int(psize))
	if err != nil {
		return err
	}
	defer r.Close()
	logger.Infof("sending assets in bacthes of %d", psize)
	page := 0
	for err == nil {
		var assets []*api.Asset
		assets, err = r.Read()
		if len(assets) > 0 {
			from := page*int(psize) + 1
			to := from + len(assets) - 1
			logger.Infof("sending batch of assets from %d to %d", from, to)
		}
		for _, a := range assets {
			annotations := []*asyncapi.Annotation{}
			for _, an := range a.AssetAnnotations {
				annotations = append(annotations, &asyncapi.Annotation{
					Key:   an.Key,
					Value: an.Value,
				})
			}
			payload := asyncapi.AssetPayload{
				Id: a.ID,
				Team: &asyncapi.Team{
					Id:          a.Team.ID,
					Name:        a.Team.Name,
					Description: a.Team.Description,
					Tag:         a.Team.Tag,
				},
				Alias:      a.Alias,
				Rolfp:      a.ROLFP.String(),
				Scannable:  *a.Scannable,
				AssetType:  (*asyncapi.AssetType)(&a.AssetType.Name),
				Identifier: a.Identifier,

				Annotations: annotations,
			}
			err = v.PushAsset(payload)
		}
		page++
	}
	if err == store.ErrReadAssetsFinished {
		return nil
	}
	return err
}

func usage() {
	fmt.Fprintln(os.Stderr, description)
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

type levelLogger struct {
	gokitlog.Logger
}

func (a levelLogger) Errorf(s string, params ...any) {
	var v string
	if len(params) == 0 {
		v = s
	} else {
		v = fmt.Sprintf(s, params...)
	}
	level.Error(a.Logger).Log("log", v)
}

func (a levelLogger) Infof(s string, params ...any) {
	var v string
	if len(params) == 0 {
		v = s
	} else {
		v = fmt.Sprintf(s, params...)
	}
	level.Info(a.Logger).Log("log", v)
}

func (a levelLogger) Debugf(s string, params ...any) {
	var v string
	if len(params) == 0 {
		v = s
	} else {
		v = fmt.Sprintf(s, params...)
	}
	level.Debug(a.Logger).Log("log", v)
}

func parseLogLevel(l string) level.Option {
	switch l {
	case "error":
		return level.AllowError()
	case "warn":
		return level.AllowWarn()
	case "debug":
		return level.AllowDebug()
	default:
		return level.AllowInfo()
	}
}
