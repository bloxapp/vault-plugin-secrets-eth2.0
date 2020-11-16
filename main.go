package main

import (
	"os"
	"strings"

	vault "github.com/bloxapp/eth2-key-manager"
	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/sdk/plugin"
	"github.com/sirupsen/logrus"

	"github.com/bloxapp/key-vault/backend"
	"github.com/bloxapp/key-vault/utils/logex"
)

// Version contains the current version of app binary.
// Basically, this is the commit hash
var Version = "latest"

func init() {
	// This is needed for signing methods
	vault.InitCrypto()
}

func main() {
	// Create plugin meta API
	var logOpts logex.Options
	var logLevels string
	apiClientMeta := &api.PluginAPIClientMeta{}
	flags := apiClientMeta.FlagSet()
	flags.StringVar(&logOpts.Format, "log-format", "", "logs format")
	flags.StringVar(&logLevels, "log-levels", "", "logs levels separated by comma")
	flags.StringVar(&logOpts.DSN, "log-dsn", "", "external DSN to send logs")
	if err := flags.Parse(os.Args[1:]); err != nil {
		logrus.WithError(err).Fatal("failed to parse flags")
	}

	logOpts.Levels = strings.Split(logLevels, ",")

	// Init logger for development proposes
	logger, err := logex.Init(logOpts)
	if err != nil {
		logrus.Fatal(err)
	}

	// Create TLS configuration
	tlsConfig := apiClientMeta.GetTLSConfig()
	tlsProviderFunc := api.VaultPluginTLSProvider(tlsConfig)

	// Serve plugin
	if err := plugin.Serve(&plugin.ServeOpts{
		BackendFactoryFunc: backend.Factory(Version, logger),
		TLSProviderFunc:    tlsProviderFunc,
	}); err != nil {
		logrus.Fatal(err)
	}
}
