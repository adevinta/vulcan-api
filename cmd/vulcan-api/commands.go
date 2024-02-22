/*
Copyright 2021 Adevinta
*/

package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"os/user"
	"strings"
	"syscall"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	goaclient "github.com/goadesign/goa/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/adevinta/vulcan-api/pkg/api"
	"github.com/adevinta/vulcan-api/pkg/api/endpoint"
	"github.com/adevinta/vulcan-api/pkg/api/middleware"
	"github.com/adevinta/vulcan-api/pkg/api/service"
	globalmiddleware "github.com/adevinta/vulcan-api/pkg/api/service/middleware/global"
	"github.com/adevinta/vulcan-api/pkg/api/store"
	"github.com/adevinta/vulcan-api/pkg/api/store/cdc"
	"github.com/adevinta/vulcan-api/pkg/api/store/global"
	"github.com/adevinta/vulcan-api/pkg/api/transport"
	"github.com/adevinta/vulcan-api/pkg/asyncapi"
	"github.com/adevinta/vulcan-api/pkg/asyncapi/kafka"
	"github.com/adevinta/vulcan-api/pkg/awscatalogue"
	awscatalogueclient "github.com/adevinta/vulcan-api/pkg/awscatalogue/client"
	"github.com/adevinta/vulcan-api/pkg/checktypes"
	"github.com/adevinta/vulcan-api/pkg/jwt"
	"github.com/adevinta/vulcan-api/pkg/reports"
	saml "github.com/adevinta/vulcan-api/pkg/saml"
	"github.com/adevinta/vulcan-api/pkg/scanengine"
	"github.com/adevinta/vulcan-api/pkg/schedule"
	"github.com/adevinta/vulcan-api/pkg/tickets"
	"github.com/adevinta/vulcan-api/pkg/vulnerabilitydb"
	vulcancore "github.com/adevinta/vulcan-core-cli/vulcan-core/client"
	metrics "github.com/adevinta/vulcan-metrics-client"
)

var (
	cfgFile  string
	httpPort int
	cfg      config
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "vulcan-api",
	Short: "A command to spawn a web server exposing the the Vulcan API",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return startServer()
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.Flags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.vulcan-api.toml)")
	rootCmd.Flags().IntVarP(&httpPort, "port", "p", 0, "web server listening port")
	err := viper.BindPFlag("server.port", rootCmd.Flags().Lookup("port"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

type serverConfig struct {
	Port         string
	SecretKey    string `mapstructure:"secret_key"`
	CookieName   string `mapstructure:"cookie_name"`
	CookieDomain string `mapstructure:"cookie_domain"`
	CookieSecure bool   `mapstructure:"cookie_secure"`
}

type dbConfig struct {
	ConnString string `mapstructure:"connection_string"`
	LogMode    bool   `mapstructure:"log_mode"`
}

// kafkaConfig stores the configuration needed to push the events of the
// async API to Kafka topics.
type kafkaConfig struct {
	User   string            `mapstructure:"user"`
	Pass   string            `mapstructure:"pass"`
	Broker string            `mapstructure:"broker"`
	Topics map[string]string `mapstructure:"topics"`
}

type logConfig struct {
	Level string `mapstructure:"level"`
}

type samlConfig struct {
	Metadata       string   `mapstructure:"saml_metadata"`
	Issuer         string   `mapstructure:"saml_issuer"`
	Callback       string   `mapstructure:"saml_callback"`
	TrustedDomains []string `mapstructure:"saml_trusted_domains"`
}

type vulcanCoreConfig struct {
	Schema string
	Host   string
}

type vulnerabilityDBConfig struct {
	URL         string `mapstructure:"url"`
	InsecureTLS bool   `mapstructure:"insecure_tls"`
}

type vulcantrackerConfig struct {
	URL            string `mapstructure:"url"`
	InsecureTLS    bool   `mapstructure:"insecure_tls"`
	OnboardedTeams string `mapstructure:"onboarded_teams"`
}

type metricsConfig struct {
	Enabled bool
}

type vulcanUIConfig struct {
	URL string `mapstructure:"url"`
}

type awsCatalogueConfig struct {
	Kind          string `mapstructure:"kind"`
	URL           string `mapstructure:"url"`
	Key           string `mapstructure:"key"`
	Retries       int    `mapstructure:"retries"`
	RetryInterval int    `mapstructure:"retry_interval"`
}

type dnsHostnameValidation struct {
	DNSHostnameValidation string `mapstructure:"dns_hostname_validation"`
}

type config struct {
	Server                serverConfig
	DB                    dbConfig
	Log                   logConfig
	SAML                  samlConfig
	Defaults              store.DefaultEntities
	ScanEngine            scanengine.Config
	Scheduler             schedule.Config
	Reports               reports.Config
	VulcanCore            vulcanCoreConfig
	VulnerabilityDB       vulnerabilityDBConfig
	VulcanTracker         vulcantrackerConfig
	Metrics               metricsConfig
	AWSCatalogue          awsCatalogueConfig
	Kafka                 kafkaConfig               `mapstructure:"kafka"`
	GlobalPolicyConfig    global.GlobalPolicyConfig `mapstructure:"globalpolicy"`
	DnsHostnameValidation dnsHostnameValidation
}

func initConfig() {
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
}

func startServer() error {
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = level.NewFilter(logger, parseLogLevel(cfg))
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	jwtSignKey := cfg.Server.SecretKey
	jwtConfig := jwt.NewJWTConfig(jwtSignKey)

	// Build vulndb client.
	vulnerabilityDBClient := vulnerabilitydb.NewClient(nil, cfg.VulnerabilityDB.URL, cfg.VulnerabilityDB.InsecureTLS)

	// Build tickets client.
	var vulcantrackerClient tickets.Client
	if cfg.VulcanTracker.URL != "" { // This is an optional component.
		vulcantrackerClient = tickets.NewClient(nil, cfg.VulcanTracker.URL, cfg.VulcanTracker.InsecureTLS)
	}

	// Build reports client.
	reportsClient, err := reports.NewClient(cfg.Reports)
	if err != nil {
		fmt.Printf("error create reports client: %v", err)
	}
	// Build metrics client.
	metricsClient, err := metrics.NewClient()
	if err != nil {
		fmt.Printf("error creating metrics client: %v", err)
		return err
	}

	// Build the AWS accounts names component.
	catalogueCfg := awscatalogueclient.AWSCatalogueAPIConfig{
		URL: cfg.AWSCatalogue.URL, Key: cfg.AWSCatalogue.Key,
		Retries: cfg.AWSCatalogue.Retries, RetryInterval: cfg.AWSCatalogue.RetryInterval,
	}
	awsCatalogueClient, err := awscatalogueclient.NewClient(cfg.AWSCatalogue.Kind, catalogueCfg)
	if err != nil {
		fmt.Printf("error creating AWS catalogue client: %v", err)
		return err
	}
	awsAccounts := awscatalogue.NewAWSAccounts(awsCatalogueClient, logger)
	go func() {
		err = awsAccounts.RefreshCache()
		// We don't wat to completely fail initializing the Vulcan API just because
		// maybe the AWS Catalogue API is down.
		if err != nil {
			fmt.Printf("error loading the aws account names cache: %+v", err)
		}
	}()

	// The JobsRunner is a dependency used by the CDC parser to execute async
	// API jobs, providing a limited access to the API service layer. But as
	// the service layer depends on the store layer, and the CBC proxies the
	// store layer, we need to perform the initialization in two steps.
	// First, declare an empty JobsRunner and inject it to the CDC parser.
	jobsRunner := &api.JobsRunner{}

	// Build CBC proxied store layer.
	db, schedulerClient, err := createVulcanitoDeps(cfg, logger, vulnerabilityDBClient, jobsRunner)
	if err != nil {
		return err
	}

	// Build service layer.
	onBoardedTeamsVT := strings.Split(cfg.VulcanTracker.OnboardedTeams, ",")
	vulcanitoService := service.New(logger, db, jwtConfig, cfg.ScanEngine, schedulerClient, cfg.Reports,
		vulnerabilityDBClient, vulcantrackerClient, reportsClient, metricsClient, awsAccounts, onBoardedTeamsVT,
		strings.EqualFold(cfg.DnsHostnameValidation.DNSHostnameValidation, "true"))

	// Second, inject the service layer to the CDC parser JobsRunner.
	jobsRunner.Client = vulcanitoService

	// Create the global entities service middleware dependencies.
	coreclient := newVulcanCoreAPIClient(cfg.VulcanCore)
	globalEntities, err := global.NewEntities(db, checktypes.New(coreclient))
	if err != nil {
		fmt.Printf("error creating checktypesinformer: %v", err)
		return err
	}
	globalMiddleware := globalmiddleware.NewEntities(logger, globalEntities, db, schedulerClient, schedulerClient, cfg.ScanEngine, metricsClient, cfg.GlobalPolicyConfig)
	// Add global middleware to the vulcanito service.
	vulcanitoService = globalMiddleware(vulcanitoService)

	endpoints := endpoint.MakeEndpoints(vulcanitoService, vulcantrackerClient != nil, logger)

	endpoints = addAuthorizationMiddleware(endpoints, db, logger)
	endpoints = addWhitelistingMiddleware(endpoints, logger)
	endpoints = addEndpointLoggingMiddleware(endpoints, db, logger)
	endpoints = addAuthenticationMiddleware(endpoints, logger, jwtSignKey, db)
	endpoints = addValidateUUIDsMiddleware(endpoints, db, globalEntities, logger)
	if cfg.Metrics.Enabled {
		endpoints = addMetricsMiddleware(endpoints, metricsClient)
	}

	handlers := transport.AttachRoutes(endpoints, logger)

	mux := http.NewServeMux()

	samlProvider, err := saml.NewProvider(cfg.SAML.Metadata, cfg.SAML.Issuer, cfg.SAML.Callback, saml.NewRandomKeyStore())
	if err != nil {
		fmt.Printf("error in SSO configuration %v", err.Error())
		return err
	}
	samlHandler := saml.NewHandler(samlProvider, cfg.SAML.TrustedDomains)
	mux.HandleFunc("/api/v1/login/callback", samlHandler.LoginCallbackHandler(saml.CallbackConfig{
		CookieName:       cfg.Server.CookieName,
		CookieDomain:     cfg.Server.CookieDomain,
		CookieSecure:     cfg.Server.CookieSecure,
		UserDataCallback: db.CreateUserIfNotExists,
		TokenGenerator:   jwtConfig.GenerateToken,
	}))
	mux.HandleFunc("/api/v1/login", samlHandler.LoginHandler())

	mux.Handle("/api/v1/", handlers)
	mux.Handle("/api/v1.1/", handlers)

	// Tiny web view authenticated from OKTA to allow users to generate from the browser an API token to interact with the API.
	mux.HandleFunc("/api/v1/home", func(w http.ResponseWriter, r *http.Request) {
		var cookie *http.Cookie
		cookie, err = r.Cookie(cfg.Server.CookieName)
		if err == nil {
			_, _ = w.Write([]byte(`<html><body><center>
			<p>In this page you can generate a Vulcan API token.</p>
			<p>You will be able to use this token to make requests to the Vulcan API.</p>
			<p>Store it safely as it will not be possible to retrieve it ever again.</p>
			<p>Clicking the button will revoke the previous token and generate a new one.</p>`))
			_, _ = w.Write([]byte(`<p><textarea id="token" type="text" placeholder="The token will appear here..." cols="86" rows="3" readonly></textarea></p>
			<p><input id="generate" type="button" value="Generate" onclick="generateToken()"></input></p>`))
			_, _ = w.Write([]byte(fmt.Sprintf(`
			<script src="https://code.jquery.com/jquery-3.5.1.min.js" integrity="sha256-9/aliU8dGd2tb6OSsuzixeV4y/faTqgFtohetphbbj0=" crossorigin="anonymous"></script>
			<script>
			function generateToken() {
				$.ajax({
					url: "/api/v1/profile",
					type: 'get',
					headers: {
						"Authorization": "Bearer %v"
					},
					dataType: 'json',
					success: function(profile) {
						$.ajax({
							url: "/api/v1/users/" + profile.id + "/token",
							type: 'post',
							headers: {
								"Authorization": "Bearer %v"
							},
							dataType: 'json',
							success: function(token) {
								$("#token").val(token.token);
							}
						});
					}
				});
			}
</script>
`, cookie.Value, cookie.Value)))
			_, _ = w.Write([]byte(`<p>For usage and more examples, check out the documentation.</p>
</center></body></html>`))
		} else {
			http.Redirect(w, r, "/api/v1/login?redirect_to=/api/v1/home", http.StatusFound)
		}
	})

	http.Handle("/", mux)

	errs := make(chan error)
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("%s", <-c)
	}()

	httpAddr := fmt.Sprintf(":%v", cfg.Server.Port)
	go func() {
		_ = logger.Log("transport", "HTTP", "addr", httpAddr, "home", fmt.Sprintf("http://127.0.0.1:%v/api/v1/home", cfg.Server.Port))
		errs <- http.ListenAndServe(httpAddr, nil)
	}()

	return logger.Log("exit", <-errs)
}

func addMetricsMiddleware(endpoints endpoint.Endpoints, metricsClient metrics.Client) endpoint.Endpoints {
	metricsMiddleware := middleware.NewMetricsMiddleware(metricsClient)

	exceptions := map[string]bool{
		endpoint.Healthcheck: true,
	}

	for name := range endpoints {
		if exceptions[name] {
			continue
		}
		endpoints[name] = metricsMiddleware.Measure(endpoints[name])
	}

	return endpoints
}

func addEndpointLoggingMiddleware(endpoints endpoint.Endpoints, db api.VulcanitoStore, logger log.Logger) endpoint.Endpoints {
	for name := range endpoints {
		endpoints[name] = middleware.EndpointLogging(logger, name, db)(endpoints[name])
	}

	return endpoints
}

func addValidateUUIDsMiddleware(endpoints endpoint.Endpoints, db api.VulcanitoStore, globalEntities *global.Entities, logger log.Logger) endpoint.Endpoints {
	exceptions := map[string]bool{
		endpoint.Healthcheck: true,
		endpoint.FindProfile: true,
		endpoint.ListUsers:   true,
		endpoint.CreateUser:  true,
		endpoint.CreateTeam:  true,
		endpoint.ListTeams:   true,
	}

	for name := range endpoints {
		if exceptions[name] {
			continue
		}
		endpoints[name] = middleware.ValidateUUIDs(db, globalEntities, logger)(endpoints[name])
	}

	return endpoints
}

func addAuthenticationMiddleware(endpoints endpoint.Endpoints, logger log.Logger, jwtSignKey string, db api.VulcanitoStore) endpoint.Endpoints {
	exceptions := map[string]bool{
		endpoint.Healthcheck: true,
	}

	for name := range endpoints {
		if exceptions[name] {
			continue
		}
		endpoints[name] = middleware.Authentication(logger, jwtSignKey, db)(endpoints[name])
	}

	return endpoints
}

func addAuthorizationMiddleware(endpoints endpoint.Endpoints, db api.VulcanitoStore, logger log.Logger) endpoint.Endpoints {
	authSrv := service.NewAuthorizationService(db)
	authMiddleware := middleware.NewAuthorizationMiddleware(authSrv, logger)

	exceptions := map[string]bool{
		endpoint.Healthcheck:                true,
		endpoint.FindJob:                    true,
		endpoint.CreateUser:                 true,
		endpoint.UpdateUser:                 true,
		endpoint.DeleteUser:                 true,
		endpoint.FindUser:                   true,
		endpoint.FindProfile:                true,
		endpoint.ListTeams:                  true,
		endpoint.CreateTeam:                 true,
		endpoint.ListUsers:                  true,
		endpoint.GenerateAPIToken:           true,
		endpoint.FindTeamsByUser:            true,
		endpoint.GlobalStatsMTTR:            true,
		endpoint.GlobalStatsExposure:        true,
		endpoint.GlobalStatsCurrentExposure: true,
		endpoint.GlobalStatsOpen:            true,
		endpoint.GlobalStatsFixed:           true,
		endpoint.GlobalStatsAssets:          true,
	}

	for name := range endpoints {
		if exceptions[name] {
			continue
		}
		endpoints[name] = authMiddleware.Authorize(endpoints[name])
	}

	return endpoints
}

func addWhitelistingMiddleware(endpoints endpoint.Endpoints, logger log.Logger) endpoint.Endpoints {
	whitelisted := map[string]bool{
		endpoint.Healthcheck: true,
		// Jobs status.
		endpoint.FindJob: true,
		// User management.
		endpoint.ListUsers:        true,
		endpoint.CreateUser:       true,
		endpoint.UpdateUser:       true,
		endpoint.FindUser:         true,
		endpoint.DeleteUser:       true,
		endpoint.FindProfile:      true,
		endpoint.GenerateAPIToken: true,
		// Team management.
		endpoint.CreateTeam: true,
		endpoint.UpdateTeam: true,
		endpoint.FindTeam:   true,
		endpoint.ListTeams:  true,
		endpoint.DeleteTeam: true,
		// Team membership.
		endpoint.FindTeamsByUser:  true,
		endpoint.ListTeamMembers:  true,
		endpoint.FindTeamMember:   true,
		endpoint.CreateTeamMember: true,
		endpoint.UpdateTeamMember: true,
		endpoint.DeleteTeamMember: true,
		// Recipients management.
		endpoint.ListRecipients:   true,
		endpoint.UpdateRecipients: true,
		// Assets management.
		endpoint.ListAssets:             true,
		endpoint.CreateAsset:            true,
		endpoint.CreateAssetMultiStatus: true,
		endpoint.MergeDiscoveredAssets:  true,
		endpoint.FindAsset:              true,
		endpoint.UpdateAsset:            true,
		endpoint.DeleteAsset:            true,
		// Asset Annotations management.
		endpoint.ListAssetAnnotations:   true,
		endpoint.CreateAssetAnnotations: true,
		endpoint.UpdateAssetAnnotations: true,
		endpoint.PutAssetAnnotations:    true,
		endpoint.DeleteAssetAnnotations: true,
		// Group management.
		endpoint.CreateGroup:    true,
		endpoint.ListGroups:     true,
		endpoint.UpdateGroup:    true,
		endpoint.DeleteGroup:    true,
		endpoint.FindGroup:      true,
		endpoint.GroupAsset:     true,
		endpoint.UngroupAsset:   true,
		endpoint.ListAssetGroup: true,

		// List scans.
		endpoint.ListProgramScans: true,
		// List programs.
		endpoint.ListPrograms: true,
		// Findings access.
		endpoint.ListFindings:           true,
		endpoint.ListFindingsIssues:     true,
		endpoint.ListFindingsByIssue:    true,
		endpoint.ListFindingsTargets:    true,
		endpoint.ListFindingsByTarget:   true,
		endpoint.FindFinding:            true,
		endpoint.CreateFindingOverwrite: true,
		endpoint.ListFindingOverwrites:  true,
		endpoint.ListFindingsLabels:     true,
		endpoint.CreateFindingTicket:    true,
		// Metrics access.
		endpoint.StatsMTTR:                  true,
		endpoint.StatsExposure:              true,
		endpoint.StatsCurrentExposure:       true,
		endpoint.StatsOpen:                  true,
		endpoint.StatsFixed:                 true,
		endpoint.GlobalStatsMTTR:            true,
		endpoint.GlobalStatsExposure:        true,
		endpoint.GlobalStatsCurrentExposure: true,
		endpoint.GlobalStatsOpen:            true,
		endpoint.GlobalStatsFixed:           true,
		endpoint.GlobalStatsAssets:          true,
	}

	for name := range endpoints {
		if whitelisted[name] {
			continue
		}
		endpoints[name] = middleware.NotWhitelisted(logger)(endpoints[name])
	}

	return endpoints
}

func createVulcanitoDeps(cfg config, l log.Logger, vulnDBClient vulnerabilitydb.Client, jobsRunner *api.JobsRunner) (api.VulcanitoStore, *schedule.Client, error) {
	db, err := store.NewDB("postgres", cfg.DB.ConnString, l, cfg.DB.LogMode, cfg.Defaults)
	if err != nil {
		err = fmt.Errorf("Error opening DB connection: %v", err)
		return nil, nil, err
	}
	cdcDB, err := cdc.NewPQDB(cfg.DB.ConnString, "")
	if err != nil {
		err = fmt.Errorf("Error opening DB connection: %v", err)
		return nil, nil, err
	}
	kcfg := cfg.Kafka
	kclient, err := kafka.NewClient(kcfg.User, kcfg.Pass, kcfg.Broker, kcfg.Topics)
	if err != nil {
		err = fmt.Errorf("error creating the kafka client: %v", err)
		return nil, nil, err
	}
	var asyncAPI cdc.AsyncAPI
	// If there is no Kafka broker specified in the configuration we consider
	// the Async API to be disabled.
	if cfg.Kafka.Broker != "" {
		asyncAPILogger := asyncapi.LevelLogger{Logger: l}
		asyncAPI = asyncapi.NewVulcan(&kclient, asyncAPILogger)
	} else {
		asyncAPI = &asyncapi.NullVulcan{}
	}
	cdcProxy := cdc.NewBrokerProxy(l, cdcDB, db, cdc.NewAsyncTxParser(vulnDBClient, jobsRunner, asyncAPI, l))
	s := schedule.NewClient(cfg.Scheduler)
	return cdcProxy, s, nil
}

func newVulcanCoreAPIClient(config vulcanCoreConfig) *vulcancore.Client {
	httpClient := newHTTPClient()
	c := vulcancore.New(goaclient.HTTPClientDoer(httpClient))
	c.Client.Scheme = config.Schema
	c.Client.Host = config.Host
	return c
}

func newHTTPClient() *http.Client {
	return http.DefaultClient
}

func parseLogLevel(cfg config) level.Option {
	switch cfg.Log.Level {
	case "ERROR":
		return level.AllowError()
	case "WARN":
		return level.AllowWarn()
	case "DEBUG":
		return level.AllowDebug()
	default:
		return level.AllowInfo()
	}
}
