package cmd

import (
	"context"
	"fmt"
	"github.com/opslevel/opslevel-go/v2024"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/automaxprocs/maxprocs"
	"opslevel-agent/signal"
	"opslevel-agent/workers"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

var (
	_version    string
	_commit     string
	_date       string
	concurrency int
)

var rootCmd = &cobra.Command{
	Use:   "opslevel-agent",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := signal.Init(context.Background())
		var wg sync.WaitGroup
		//go workers.NewWebhookWorker().Run(ctx, &wg)
		go workers.NewK8SWorker(viper.GetString("cluster"), viper.GetString("integration"), newClient()).Run(ctx, &wg)
		time.Sleep(1 * time.Second)
		wg.Wait()
	},
}

func Execute(version, commit, date string) {
	_version = version
	_commit = commit
	_date = date
	err := rootCmd.Execute()
	if err != nil {
		log.Error().Err(err).Msgf("error executing")
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().String("log-format", "TEXT", "overrides environment variable 'OPSLEVEL_LOG_FORMAT' (options [\"JSON\", \"TEXT\"])")
	rootCmd.PersistentFlags().String("log-level", "INFO", "overrides environment variable 'OPSLEVEL_LOG_LEVEL' (options [\"ERROR\", \"WARN\", \"INFO\", \"DEBUG\"])")
	rootCmd.PersistentFlags().String("api-token", "", "The OpsLevel API Token. Overrides environment variable 'OPSLEVEL_API_TOKEN'")
	rootCmd.PersistentFlags().String("api-url", "https://app.opslevel.com/", "The OpsLevel API Url. Overrides environment variable 'OPSLEVEL_API_URL'")
	rootCmd.PersistentFlags().Int("api-timeout", 40, "The OpsLevel API timeout in seconds. Overrides environment variable 'OPSLEVEL_API_TIMEOUT'")

	rootCmd.PersistentFlags().String("integration", "", "The OpsLevel integration id or alias to send the data for.")
	rootCmd.PersistentFlags().String("cluster", "dev", "The name of the cluster the agent is deployed in.")

	cobra.CheckErr(viper.BindPFlags(rootCmd.PersistentFlags()))
	cobra.CheckErr(viper.BindEnv("api-url", "OPSLEVEL_API_URL"))
	cobra.CheckErr(viper.BindEnv("api-token", "OPSLEVEL_API_TOKEN"))
	cobra.CheckErr(viper.BindEnv("api-timeout", "OPSLEVEL_API_TIMEOUT"))

	cobra.OnInitialize(func() {
		setupEnv()
		setupLogging()
		setupConcurrency()
	})
}

func setupEnv() {
	viper.SetEnvPrefix("OPSLEVEL")
	viper.AutomaticEnv()
}

func setupLogging() {
	logFormat := strings.ToLower(viper.GetString("log-format"))
	logLevel := strings.ToLower(viper.GetString("log-level"))

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	if logFormat == "text" {
		output := zerolog.ConsoleWriter{Out: os.Stderr}
		log.Logger = log.Output(output)
	}

	switch {
	case logLevel == "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case logLevel == "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case logLevel == "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case logLevel == "trace":
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}

func setupConcurrency() {
	_, err := maxprocs.Set(maxprocs.Logger(log.Debug().Msgf))
	cobra.CheckErr(err)

	// TODO: how does this work in this app??
	concurrency = viper.GetInt("workers")
	if concurrency <= 0 {
		concurrency = runtime.GOMAXPROCS(0)
	}
}

func newClient() *opslevel.Client {
	client := opslevel.NewGQLClient(
		opslevel.SetAPIToken(viper.GetString("api-token")),
		opslevel.SetURL(viper.GetString("api-url")),
		opslevel.SetUserAgentExtra(fmt.Sprintf("agent-%s", _version)),
		opslevel.SetTimeout(time.Second*time.Duration(viper.GetInt("api-timeout"))),
	)
	cobra.CheckErr(client.Validate())
	return client
}
