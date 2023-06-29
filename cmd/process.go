package cmd

import (
	"context"
	"os"
	"os/signal"

	"go.uber.org/zap"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"go.infratographer.com/ipam-api/pkg/ipamclient"
	"go.infratographer.com/loadbalancer-manager-haproxy/pkg/lbapi"
	"go.infratographer.com/loadbalancer-manager-haproxy/x/oauth2x"
	"go.infratographer.com/x/echox"
	"go.infratographer.com/x/events"
	"go.infratographer.com/x/versionx"
	"go.infratographer.com/x/viperx"

	"go.infratographer.com/loadbalancer-provider-haproxy/internal/config"
	"go.infratographer.com/loadbalancer-provider-haproxy/internal/server"
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

	processCmd.PersistentFlags().String("api-endpoint", "http://localhost:7608", "endpoint for load balancer API. defaults to supergraph-endpoint if set")
	viperx.MustBindFlag(viper.GetViper(), "api-endpoint", processCmd.PersistentFlags().Lookup("api-endpoint"))

	processCmd.PersistentFlags().String("ipam-endpoint", "http://localhost:7608", "endpoint for IPAM API. defaults to supergraph-endpoint if set")
	viperx.MustBindFlag(viper.GetViper(), "ipam-endpoint", processCmd.PersistentFlags().Lookup("ipam-endpoint"))

	processCmd.PersistentFlags().String("supergraph-endpoint", "", "endpoint for load balancer API")
	viperx.MustBindFlag(viper.GetViper(), "supergraph-endpoint", processCmd.PersistentFlags().Lookup("supergraph-endpoint"))

	processCmd.PersistentFlags().StringSlice("event-locations", nil, "location id(s) to filter events for")
	viperx.MustBindFlag(viper.GetViper(), "event-locations", processCmd.PersistentFlags().Lookup("event-locations"))

	processCmd.PersistentFlags().StringSlice("event-topics", nil, "event topics to subscribe to")
	viperx.MustBindFlag(viper.GetViper(), "event-topics", processCmd.PersistentFlags().Lookup("event-topics"))

	processCmd.PersistentFlags().String("ipblock", "", "ip block id to use for requesting load balancer IPs")
	viperx.MustBindFlag(viper.GetViper(), "ipblock", processCmd.PersistentFlags().Lookup("ipblock"))

	events.MustViperFlagsForPublisher(viper.GetViper(), processCmd.Flags(), appName)
	events.MustViperFlagsForSubscriber(viper.GetViper(), processCmd.Flags())
	oauth2x.MustViperFlags(viper.GetViper(), processCmd.Flags())

	rootCmd.AddCommand(processCmd)
}

func process(ctx context.Context, logger *zap.SugaredLogger) error {
	cx, cancel := context.WithCancel(ctx)

	eSrv, err := echox.NewServer(
		logger.Desugar(),
		echox.ConfigFromViper(viper.GetViper()),
		versionx.BuildDetails(),
	)
	if err != nil {
		logger.Fatal("failed to initialize new server", zap.Error(err))
	}

	server := &server.Server{
		Context:          cx,
		Debug:            viper.GetBool("logging.debug"),
		Echo:             eSrv,
		Locations:        viper.GetStringSlice("event-locations"),
		Logger:           logger,
		SubscriberConfig: config.AppConfig.Events.Subscriber,
		Topics:           viper.GetStringSlice("event-topics"),
		IPBlock:          viper.GetString("ipblock"),
	}

	// init lbapi client and ipam client
	if config.AppConfig.OIDC.ClientID != "" {
		oauthHTTPClient := oauth2x.NewClient(ctx, oauth2x.NewClientCredentialsTokenSrc(ctx, config.AppConfig.OIDC))
		server.APIClient = lbapi.NewClient(determineEndpoint(viper.GetString("api-endpoint"), viper.GetString("supergraph-endpoint")),
			lbapi.WithHTTPClient(oauthHTTPClient),
		)
		server.IPAMClient = ipamclient.NewClient(determineEndpoint(viper.GetString("api-endpoint"), viper.GetString("supergraph-endpoint")),
			ipamclient.WithHTTPClient(oauthHTTPClient),
		)
	} else {
		server.APIClient = lbapi.NewClient(determineEndpoint(viper.GetString("api-endpoint"), viper.GetString("supergraph-endpoint")))
		server.IPAMClient = ipamclient.NewClient(determineEndpoint(viper.GetString("ipam-endpoint"), viper.GetString("supergraph-endpoint")))
	}

	pub, err := events.NewPublisher(config.AppConfig.Events.Publisher)
	if err != nil {
		logger.Fatalw("failed to create publisher", "error", err)
	}

	server.Publisher = pub

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

func determineEndpoint(endpoint string, supergraph string) string {
	if supergraph != "" {
		return supergraph
	}

	return endpoint
}
