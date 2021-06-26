package backend

import (
	"context"
	"encoding/hex"
	"encoding/json"
	vault "github.com/bloxapp/eth2-key-manager"
	"github.com/bloxapp/eth2-key-manager/core"

	"github.com/bloxapp/eth2-key-manager/stores/inmemory"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/pkg/errors"

	"github.com/bloxapp/key-vault/backend/store"
)

// Endpoints patterns
const (
	// StoragePattern is the path pattern for storage endpoint
	StoragePattern         = "storage"
)

func storagePaths(b *backend) []*framework.Path {
	return []*framework.Path{
		{
			Pattern:         StoragePattern,
			HelpSynopsis:    "Update storage using one or more accounts",
			HelpDescription: `Manage KeyVault storage`,
			Fields: map[string]*framework.FieldSchema{
				"data": &framework.FieldSchema{
					Type:        framework.TypeString,
					Description: "storage to update",
				},
			},
			ExistenceCheck: b.pathExistenceCheck,
			Callbacks: map[logical.Operation]framework.OperationFunc{
				logical.CreateOperation: b.pathStorageUpdateV2,
			},
		},
	}
}

// getVaultMemoryStorage returns unmarshalled in-memory storage
func getVaultMemoryStorage(data *framework.FieldData) (*inmemory.InMemStore, error) {
	storage := data.Get("data").(string)
	storageBytes, err := hex.DecodeString(storage)
	if err != nil {
		return nil, errors.Wrap(err, "failed to HEX decode storage")
	}

	var vaultMemoryStorage *inmemory.InMemStore
	err = json.Unmarshal(storageBytes, &vaultMemoryStorage)
	if err != nil {
		return nil, errors.Wrap(err, "failed to JSON un-marshal storage")
	}
	return vaultMemoryStorage, nil
}

// pathStorageUpdate updates all accounts from new uploaded storage
func (b *backend) pathStorageUpdate(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	vaultMemoryStorage, err := getVaultMemoryStorage(data)
	if err != nil {
		return nil, err
	}

	_, err = store.FromInMemoryStore(ctx, vaultMemoryStorage, req.Storage)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update storage")
	}

	return &logical.Response{
		Data: map[string]interface{}{
			"status": true,
		},
	}, nil
}

// BuildKeyVault building new key vault and wallet if not exists - otherwise return existing ones
func (b *backend) BuildKeyVault(ctx context.Context, req *logical.Request, data *framework.FieldData) (*inmemory.InMemStore, *store.HashicorpVaultStore, core.Wallet, error) {
	// Load config
	config, err := b.readConfig(ctx, req.Storage)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "failed to get config")
	}

	// Get hashicorp storage
	hashicorpStorage := store.NewHashicorpVaultStore(ctx, req.Storage, config.Network)
	options := vault.KeyVaultOptions{}
	options.SetStorage(hashicorpStorage)

	// Try to get existing wallet
	var wallet core.Wallet

	// Get memory storage for new posted accounts
	vaultMemoryStorage, err := getVaultMemoryStorage(data)
	if err != nil {
		return nil, nil, nil, err
	}

	portfolio, err := vault.OpenKeyVault(&options)
	if err != nil {
		// If no existing wallet - create new one
		wallet, err = vaultMemoryStorage.OpenWallet()
		if err != nil {
			return nil, nil, nil, errors.Wrap(err, "failed to open new wallet")
		}
	} else {
		// Use existing wallet
		wallet, err = portfolio.Wallet()
		if err != nil {
			return nil, nil, nil, errors.Wrap(err, "failed to retrieve wallet by name")
		}
	}

	// Save wallet in key vault
	err = hashicorpStorage.SaveWallet(wallet)
	if err != nil {
		return nil, nil, nil, err
	}

	return vaultMemoryStorage, hashicorpStorage, wallet, nil
}

// pathStorageUpdateV2 updates accounts from one account storage
func (b *backend) pathStorageUpdateV2(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	// Build new or use existing key vault with wallet
	vaultMemoryStorage, hashicorpStorage, wallet, err := b.BuildKeyVault(ctx, req, data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build key vault and wallet")
	}

	// Update accounts
	_, err = store.UpdateAccounts(vaultMemoryStorage, wallet, hashicorpStorage)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update storage")
	}

	return &logical.Response{
		Data: map[string]interface{}{
			"status": true,
		},
	}, nil
}
