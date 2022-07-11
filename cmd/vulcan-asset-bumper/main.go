package main

import (
	"flag"
	"fmt"
	"os"

	gokitlog "github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/spf13/viper"

	"github.com/adevinta/vulcan-api/pkg/api/store"
	"github.com/adevinta/vulcan-api/pkg/asyncapi"
	"github.com/adevinta/vulcan-api/pkg/asyncapi/kafka"
)

var description = `Bumps all the assets in the configured vulcan db to the configured kafka topic.
The tool will lock all the teams, assets and annotations of the db, so the vulcan-api
won't be able to work meanwhile this tool is running.`

func main() {
	flag.Usage = buildUsageFunc(flag.PrintDefaults)
	configFile := flag.String("c", "", "path the config file")
	logLevel := flag.String("l", "error", `log level valid values: "error", "warn" "debug"`)
	flag.Parse()
	if *configFile == "" {
		flag.Usage()
		os.Exit(1)
	}
	cfg := &config{}
	viper.SetConfigFile(*configFile)
	if err := viper.ReadInConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "error reading config %s:%+v", *configFile, err)
		os.Exit(1)
	}
	if err := viper.Unmarshal(&cfg); err != nil {
		fmt.Fprintf(os.Stderr, "error reading config file %s: %v", *configFile, err)
		os.Exit(1)
	}

	var logger gokitlog.Logger
	{
		logger = gokitlog.NewLogfmtLogger(os.Stderr)
		logger = level.NewFilter(logger, parseLogLevel(*logLevel))
	}
	store, err := store.NewStore("", cfg.DB.ConnString, logger, false, map[string][]string{})
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
	apiLogger := asyncVulcanLogger{logger}
	v := asyncapi.NewVulcan(&kclient, apiLogger)
	err = bump(v, store, logger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error bumping assets: %v", err)
		os.Exit(1)
	}
}

func bump(v asyncapi.Vulcan, s store.Store, logger gokitlog.Logger) error {
	r, err := s.NewAssetReader(true, 5)
	if err != nil {
		return err
	}
	defer r.Close()
	for err == nil {
		assets, err := r.Read()
		for _, a := range assets {
			var annotations []*asyncapi.Annotation
			for _, an := range a.AssetAnnotations {
				annotations = append(annotations, &asyncapi.Annotation{
					Key:   an.Key,
					Value: an.Value,
				})
			}
			payload := asyncapi.AssetPayload{
				Id:          a.ID,
				Team:        &asyncapi.Team{Id: a.Team.ID, Name: a.Team.Name, Description: a.Team.Description},
				Alias:       a.Alias,
				Rolfp:       a.ROLFP.String(),
				Scannable:   *a.Scannable,
				AssetType:   (*asyncapi.AssetType)(&a.AssetType.Name),
				Identifier:  a.Identifier,
				Annotations: annotations,
			}
			if err = v.PushAsset(payload); err != nil {
				return err
			}
		}
	}
	if err == store.ErrReadAssetsFinished {
		return nil
	}
	return err
}

func buildUsageFunc(defaultUsage func()) func() {
	return func() {
		fmt.Fprintln(os.Stderr, description)
		defaultUsage()
	}
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

type asyncVulcanLogger struct {
	gokitlog.Logger
}

func (a asyncVulcanLogger) ErrorF(s string, params ...any) {
	var v string
	if len(params) == 0 {
		v = s
	} else {
		v = fmt.Sprintf(s, a)
	}
	level.Error(a.Logger).Log(v)
}

func (a asyncVulcanLogger) InfoF(s string, params ...any) {
	var v string
	if len(params) == 0 {
		v = s
	} else {
		v = fmt.Sprintf(s, a)
	}
	level.Info(a.Logger).Log(v)
}

func (a asyncVulcanLogger) DebugF(s string, params ...any) {
	var v string
	if len(params) == 0 {
		v = s
	} else {
		v = fmt.Sprintf(s, a)
	}
	level.Info(a.Logger).Log(v)
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
