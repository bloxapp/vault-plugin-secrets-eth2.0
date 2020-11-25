package backend

import (
	"context"
	"fmt"
	"time"

	"github.com/bloxapp/eth2-key-manager/core"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/pkg/errors"
)

// Helpers
const (
	// This is the format of genesis time, e.g. 2020-08-04 13:00:08 UTC
	genesisTimeFormat = "2006-01-02 15:04:05 MST"
)

// Endpoints patterns
const (
	// ConfigPattern is the path pattern for config endpoint
	ConfigPattern = "config"
)

// Config contains the configuration for each mount
type Config struct {
	Network     core.Network `json:"network"`
	GenesisTime time.Time    `json:"genesis_time"`
}

func configPaths(b *backend) []*framework.Path {
	return []*framework.Path{
		{
			Pattern: ConfigPattern,
			Callbacks: map[logical.Operation]framework.OperationFunc{
				logical.CreateOperation: b.pathWriteConfig,
				logical.UpdateOperation: b.pathWriteConfig,
				logical.ReadOperation:   b.pathReadConfig,
			},
			HelpSynopsis:    "Configure the Vault Ethereum plugin.",
			HelpDescription: "Configure the Vault Ethereum plugin.",
			Fields: map[string]*framework.FieldSchema{
				"network": {
					Type: framework.TypeString,
					Description: `Ethereum network - can be one of the following values:
					mainnet - MainNet Network
					pyrmont - Pyrmont Test Network`,
					AllowedValues: []interface{}{
						string(core.PyrmontNetwork),
						string(core.MainNetwork),
					},
				},
				"genesis_time": {
					Type:        framework.TypeString,
					Description: `Genesis time of the network`,
				},
			},
		},
	}
}

// pathWriteConfig is the write config path handler
func (b *backend) pathWriteConfig(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	network := data.Get("network").(string)
	genesisTimeStr := data.Get("genesis_time").(string)

	// Parse genesis time
	genesisTime, err := time.Parse(genesisTimeFormat, genesisTimeStr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse genesis time")
	}

	configBundle := Config{
		Network:     core.NetworkFromString(network),
		GenesisTime: genesisTime,
	}

	// Create storage entry
	entry, err := logical.StorageEntryJSON("config", configBundle)
	if err != nil {
		return nil, err
	}

	// Store config
	if err := req.Storage.Put(ctx, entry); err != nil {
		return nil, err
	}

	// Return the secret
	return &logical.Response{
		Data: map[string]interface{}{
			"network":      configBundle.Network,
			"genesis_time": configBundle.GenesisTime,
		},
	}, nil
}

// pathReadConfig is the read config path handler
func (b *backend) pathReadConfig(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	configBundle, err := b.readConfig(ctx, req.Storage)
	if err != nil {
		return nil, err
	}

	if configBundle == nil {
		return nil, nil
	}

	// Return the secret
	return &logical.Response{
		Data: map[string]interface{}{
			"network":      configBundle.Network,
			"genesis_time": configBundle.GenesisTime,
		},
	}, nil
}

// readConfig returns the configuration for this PluginBackend.
func (b *backend) readConfig(ctx context.Context, s logical.Storage) (*Config, error) {
	entry, err := s.Get(ctx, "config")
	if err != nil {
		return nil, err
	}

	if entry == nil {
		return nil, fmt.Errorf("the plugin has not been configured yet")
	}

	var result Config
	if err := entry.DecodeJSON(&result); err != nil {
		return nil, errors.Wrap(err, "error reading configuration")
	}

	return &result, nil
}

func (b *backend) configured(ctx context.Context, req *logical.Request) (*Config, error) {
	config, err := b.readConfig(ctx, req.Storage)
	if err != nil {
		return nil, err
	}

	return config, nil
}
