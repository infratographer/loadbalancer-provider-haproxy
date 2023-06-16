package cmd

import (
	"context"
	"errors"
	"os"
	"os/signal"

	"go.uber.org/zap"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"go.infratographer.com/loadbalancer-manager-haproxy/pkg/lbapi"
	"go.infratographer.com/loadbalancer-manager-haproxy/x/oauth2x"
	"go.infratographer.com/x/echox"
	"go.infratographer.com/x/versionx"
	"go.infratographer.com/x/viperx"

	"go.infratographer.com/loadbalanceroperator/internal/config"
	"go.infratographer.com/loadbalanceroperator/internal/srv"
)

// processCmd represents the base command when called without any subcommands
var processCmd = &cobra.Command{
	Use:   "process",
	Short: "Begin processing requests related to LBs",
	Long:  `Begin processing requests from message queues to manage LBs`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return process(cmd.Context(), logger)
	},
}

var (
	processDevMode bool
)

func init() {
	// only available as a CLI arg because it shouldn't be something that could accidentially end up in a config file or env var
	processCmd.Flags().BoolVar(&processDevMode, "dev", false, "dev mode: disables all auth checks, pretty logging, etc.")

	processCmd.PersistentFlags().String("api-endpoint", "http://localhost:7608", "endpoint for load balancer API")
	viperx.MustBindFlag(viper.GetViper(), "api-endpoint", processCmd.PersistentFlags().Lookup("api-endpoint"))

	processCmd.PersistentFlags().String("ipam-endpoint", "http://localhost:7609", "endpoint for IPAM API")
	viperx.MustBindFlag(viper.GetViper(), "ipam-endpoint", processCmd.PersistentFlags().Lookup("ipam-endpoint"))

	processCmd.PersistentFlags().String("supergraph-endpoint", "http://localhost:8067", "endpoint for infratographer supergraph")
	viperx.MustBindFlag(viper.GetViper(), "supergraph-endpoint", processCmd.PersistentFlags().Lookup("supergraph-endpoint"))

	processCmd.PersistentFlags().StringSlice("event-locations", nil, "location id(s) to filter events for")
	viperx.MustBindFlag(viper.GetViper(), "event-locations", processCmd.PersistentFlags().Lookup("event-locations"))

	processCmd.PersistentFlags().StringSlice("event-topics", nil, "event topics to subscribe to")
	viperx.MustBindFlag(viper.GetViper(), "event-topics", processCmd.PersistentFlags().Lookup("event-topics"))

	rootCmd.AddCommand(processCmd)
}

func process(ctx context.Context, logger *zap.SugaredLogger) error {
	if err := validateFlags(); err != nil {
		return err
	}

	cx, cancel := context.WithCancel(ctx)

	eSrv, err := echox.NewServer(
		logger.Desugar(),
		echox.ConfigFromViper(viper.GetViper()),
		versionx.BuildDetails(),
	)
	if err != nil {
		logger.Fatal("failed to initialize new server", zap.Error(err))
	}

	server := &srv.Server{
		Context:          cx,
		Debug:            viper.GetBool("logging.debug"),
		Echo:             eSrv,
		Locations:        viper.GetStringSlice("event-locations"),
		Logger:           logger,
		SubscriberConfig: config.AppConfig.Events.Subscriber,
		Topics:           viper.GetStringSlice("event-topics"),
	}

	// init lbapi client
	if config.AppConfig.OIDC.ClientID != "" {
		oauthHTTPClient := oauth2x.NewClient(ctx, oauth2x.NewClientCredentialsTokenSrc(ctx, config.AppConfig.OIDC))
		server.APIClient = lbapi.NewClient(viper.GetString("api-endpoint"),
			lbapi.WithHTTPClient(oauthHTTPClient),
		)
	} else {
		server.APIClient = lbapi.NewClient(viper.GetString("api-endpoint"))
	}

	if err := server.Run(cx); err != nil {
		logger.Fatalw("failed starting server", "error", err)
		cancel()
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	recvSig := <-sigCh
	signal.Stop(sigCh)
	cancel()
	logger.Infof("exiting. Performing necessary cleanup", recvSig)

	return nil
}

func newKubeAuth(path string) (*rest.Config, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		if path != "" {
			config, err = clientcmd.BuildConfigFromFlags("", path)
			if err != nil {
				return nil, errors.Join(err, errInvalidKubeClient)
			}
		} else {
			return nil, errors.Join(err, errInvalidKubeClient)
		}
	}

	return config, nil
}

func validateFlags() error {
	if viper.GetString("chart-path") == "" {
		return errChartPath
	}

	if len(viper.GetStringSlice("event-topics")) < 1 {
		return errRequiredTopics
	}

	return nil
}

func loadHelmChart(chartPath string) (*chart.Chart, error) {
	chart, err := loader.Load(chartPath)
	if err != nil {
		logger.Errorw("failed to load helm chart", "error", err)

		return nil, errors.Join(err, errInvalidHelmChart)
	}

	return chart, nil
}
