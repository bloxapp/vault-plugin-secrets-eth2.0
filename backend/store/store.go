package store

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/bloxapp/eth2-key-manager/core"
	"github.com/bloxapp/eth2-key-manager/encryptor"
	"github.com/bloxapp/eth2-key-manager/stores/in_memory"
	"github.com/bloxapp/eth2-key-manager/wallets"
	"github.com/bloxapp/eth2-key-manager/wallets/hd"
	"github.com/google/uuid"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/pkg/errors"
)

// Paths
const (
	WalletDataPath = "wallet/data"

	AccountBase = "wallet/accounts/"
	AccountPath = AccountBase + "%s"
)

// HashicorpVaultStore implements store.Store interface using Vault.
type HashicorpVaultStore struct {
	storage logical.Storage
	ctx     context.Context
	network core.Network

	encryptor          encryptor.Encryptor
	encryptionPassword []byte
}

// NewHashicorpVaultStore is the constructor of HashicorpVaultStore.
func NewHashicorpVaultStore(ctx context.Context, storage logical.Storage, network core.Network) *HashicorpVaultStore {
	return &HashicorpVaultStore{
		storage: storage,
		network: network,
		ctx:     ctx,
	}
}

// FromInMemoryStore creates the HashicorpVaultStore based on the given in-memory store.
func FromInMemoryStore(ctx context.Context, inMem *in_memory.InMemStore, storage logical.Storage) (*HashicorpVaultStore, error) {
	// first delete old data
	// delete all accounts
	res, err := storage.List(ctx, AccountBase)
	if err != nil {
		return nil, err
	}

	for _, accountID := range res {
		path := fmt.Sprintf(AccountPath, accountID)
		if err := storage.Delete(ctx, path); err != nil {
			return nil, err
		}
	}

	if err := storage.Delete(ctx, WalletDataPath); err != nil {
		return nil, err
	}

	if err := storage.Delete(ctx, AccountBase); err != nil {
		return nil, err
	}

	if err := storage.Delete(ctx, WalletHighestAttestationPath); err != nil {
		return nil, err
	}

	// Create new store
	newStore := NewHashicorpVaultStore(ctx, storage, inMem.Network())

	// Save wallet
	wallet, err := inMem.OpenWallet()
	if err != nil {
		return nil, err
	}

	if err := newStore.SaveWallet(wallet); err != nil {
		return nil, err
	}

	// Save accounts
	for _, acc := range wallet.Accounts() {
		if err := newStore.SaveAccount(acc); err != nil {
			return nil, err
		}
	}

	// save highest att.
	for _, acc := range wallet.Accounts() {
		if val := inMem.RetrieveHighestAttestation(acc.ValidatorPublicKey()); val != nil {
			if err := newStore.SaveHighestAttestation(acc.ValidatorPublicKey(), val); err != nil {
				return nil, err
			}
		}
	}

	return newStore, nil
}

// Name returns the name of the store.
func (store *HashicorpVaultStore) Name() string {
	return "Hashicorp Vault"
}

// Network returns the network the storage is related to.
func (store *HashicorpVaultStore) Network() core.Network {
	return store.network
}

// SaveWallet implements Storage interface.
func (store *HashicorpVaultStore) SaveWallet(wallet core.Wallet) error {
	data, err := json.Marshal(wallet)
	if err != nil {
		return errors.Wrap(err, "failed to marshal wallet")
	}

	return store.storage.Put(store.ctx, &logical.StorageEntry{
		Key:      WalletDataPath,
		Value:    data,
		SealWrap: false,
	})
}

// OpenWallet returns nil,nil if no wallet was found
func (store *HashicorpVaultStore) OpenWallet() (core.Wallet, error) {
	path := WalletDataPath
	entry, err := store.storage.Get(store.ctx, path)
	if err != nil {
		return nil, err
	}

	// Return nothing if there is no record
	if entry == nil {
		return nil, fmt.Errorf("wallet not found")
	}

	var ret hd.HDWallet
	ret.SetContext(store.freshContext())
	if err := json.Unmarshal(entry.Value, &ret); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal HD Wallet object")
	}

	return &ret, nil
}

// ListAccounts returns an empty array for no accounts
func (store *HashicorpVaultStore) ListAccounts() ([]core.ValidatorAccount, error) {
	w, err := store.OpenWallet()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get wallet")
	}

	return w.Accounts(), nil
}

// SaveAccount stores the given account in DB.
func (store *HashicorpVaultStore) SaveAccount(account core.ValidatorAccount) error {
	data, err := json.Marshal(account)
	if err != nil {
		return errors.Wrap(err, "failed to marshal account object")
	}

	return store.storage.Put(store.ctx, &logical.StorageEntry{
		Key:      fmt.Sprintf(AccountPath, account.ID().String()),
		Value:    data,
		SealWrap: false,
	})
}

// OpenAccount opens an account by the given ID. Returns nil,nil if no account was found.
func (store *HashicorpVaultStore) OpenAccount(accountID uuid.UUID) (core.ValidatorAccount, error) {
	path := fmt.Sprintf(AccountPath, accountID)
	entry, err := store.storage.Get(store.ctx, path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get record with path '%s'", path)
	}

	// Return nothing if there is no record
	if entry == nil {
		return nil, nil
	}

	// un-marshal
	var ret wallets.HDAccount
	ret.SetContext(store.freshContext())
	if err := json.Unmarshal(entry.Value, &ret); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal HD account object")
	}
	return &ret, nil
}

// DeleteAccount deletes the given account
func (store *HashicorpVaultStore) DeleteAccount(accountID uuid.UUID) error {
	path := fmt.Sprintf(AccountPath, accountID)
	if err := store.storage.Delete(store.ctx, path); err != nil {
		return errors.Wrapf(err, "failed to delete record with path '%s'", path)
	}
	return nil
}

// SetEncryptor sets the given encryptor. Could be nil value.
func (store *HashicorpVaultStore) SetEncryptor(encryptor encryptor.Encryptor, password []byte) {
	store.encryptor = encryptor
	store.encryptionPassword = password
}

func (store *HashicorpVaultStore) freshContext() *core.WalletContext {
	return &core.WalletContext{
		Storage: store,
	}
}

func (store *HashicorpVaultStore) canEncrypt() bool {
	return store.encryptor != nil && store.encryptionPassword != nil
}
