// Package cmd is the root of our application
package cmd

import (
	"fmt"
	"os"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"go.infratographer.com/loadbalancer-provider-haproxy/internal/config"

	"go.infratographer.com/x/loggingx"
	"go.infratographer.com/x/viperx"
)

var (
	cfgFile string
	logger  *zap.SugaredLogger
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "loadbalancer-provider-haproxy",
	Short: "A controller for processing requests related to loadbalancer provisioning",
	Long:  "A controller for processing requests related to loadbalancer provisioning",
}

var appName = "loadbalancerproviderhaproxy"

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.loadbalancerproviderhaproxy.yaml)")

	loggingx.MustViperFlags(viper.GetViper(), rootCmd.PersistentFlags())

	rootCmd.PersistentFlags().String("healthcheck-port", ":8080", "port to run healthcheck probe on")
	viperx.MustBindFlag(viper.GetViper(), "healthcheck-port", rootCmd.PersistentFlags().Lookup("healthcheck-port"))

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".loadbalancerproviderhaproxy" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".loadbalancerproviderhaproxy")
	}

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.SetEnvPrefix("loadbalancerproviderhaproxy")
	viper.AutomaticEnv() // read in environment variables that match

	setupAppConfig()

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}

	logger = loggingx.InitLogger(appName, config.AppConfig.Logging)
}

// setupAppConfig loads our config.AppConfig struct with the values bound by
// viper. Then, anywhere we need these values, we can just return to AppConfig
// instead of performing viper.GetString(...), viper.GetBool(...), etc.
func setupAppConfig() {
	err := viper.Unmarshal(&config.AppConfig)
	if err != nil {
		fmt.Printf("unable to decode app config: %s", err)
		os.Exit(1)
	}
}
