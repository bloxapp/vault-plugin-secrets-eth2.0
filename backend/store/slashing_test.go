package store_test

import (
	"context"
	"testing"

	"github.com/bloxapp/eth2-key-manager/core"
	"github.com/bloxapp/eth2-key-manager/stores"
	"github.com/hashicorp/vault/sdk/logical"

	"github.com/bloxapp/key-vault/backend/store"
)

func getSlashingStorage() core.SlashingStore {
	return store.NewHashicorpVaultStore(context.Background(), &logical.InmemStorage{}, core.PyrmontNetwork)
}

func TestSavingProposal(t *testing.T) {
	stores.TestingSaveProposal(getSlashingStorage(), t)
}

func TestSavingAttestation(t *testing.T) {
	stores.TestingSaveAttestation(getSlashingStorage(), t)
}

func TestSavingHighestAttestation(t *testing.T) {
	stores.TestingSaveHighestAttestation(getSlashingStorage(), t)
}
