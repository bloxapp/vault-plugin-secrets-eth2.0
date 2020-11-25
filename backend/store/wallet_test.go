package store_test

import (
	"context"
	"testing"

	"github.com/bloxapp/eth2-key-manager/core"
	"github.com/bloxapp/eth2-key-manager/stores"
	"github.com/hashicorp/vault/sdk/logical"

	"github.com/bloxapp/key-vault/backend/store"
)

func getStorage() logical.Storage {
	return &logical.InmemStorage{}
}

func getWalletStorage() core.Storage {
	return store.NewHashicorpVaultStore(context.Background(), getStorage(), core.PyrmontNetwork)
}

func TestOpeningAccounts(t *testing.T) {
	stores.TestingOpenAccounts(getWalletStorage(), t)
}

func TestNonExistingWallet(t *testing.T) {
	stores.TestingNonExistingWallet(getWalletStorage(), t)
}

func TestWalletStorage(t *testing.T) {
	stores.TestingWalletStorage(getWalletStorage(), t)
}
