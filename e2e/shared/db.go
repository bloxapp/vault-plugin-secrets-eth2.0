package shared

import (
	"fmt"
	"testing"

	"github.com/bloxapp/eth2-key-manager/core"
	"github.com/bloxapp/eth2-key-manager/stores/in_memory"
	"github.com/bloxapp/eth2-key-manager/wallet_hd"
	"github.com/stretchr/testify/require"
	types "github.com/wealdtech/go-eth2-types/v2"
)

// AccountIndex is the test account index.
const AccountIndex = 0

// BaseInmemStorage creates the in-memory storage and creates the base account.
func BaseInmemStorage(t *testing.T, minimalSlashingData bool) (*in_memory.InMemStore, error) {
	err := types.InitBLS()
	require.NoError(t, err)

	store := in_memory.NewInMemStore(core.PyrmontNetwork)

	entropy, err := core.GenerateNewEntropy()
	require.NoError(t, err)

	seed, err := core.SeedFromEntropy(entropy, "")
	require.NoError(t, err)

	// wallet
	wallet := wallet_hd.NewHDWallet(&core.WalletContext{Storage: store})
	if err := store.SaveWallet(wallet); err != nil {
		return nil, err
	}

	// account
	acc, err := wallet.CreateValidatorAccount(seed, nil)
	if err != nil {
		return nil, err
	}
	err = store.SaveAccount(acc)
	if err != nil {
		return nil, err
	}

	// base highest att.
	if minimalSlashingData {
		err = store.SaveHighestAttestation(acc.ValidatorPublicKey(), &core.BeaconAttestation{
			Source: &core.Checkpoint{
				Epoch: 0,
				Root:  nil,
			},
			Target: &core.Checkpoint{
				Epoch: 0,
				Root:  nil,
			},
		})
		if err != nil {
			return nil, err
		}
	}

	return store, nil
}

// RetrieveAccount retrieves test account fro the storage.
func RetrieveAccount(t *testing.T, store core.Storage) core.ValidatorAccount {
	accounts, err := store.ListAccounts()
	require.NoError(t, err)

	for _, acc := range accounts {
		if acc.BasePath() == fmt.Sprintf("/%d", AccountIndex) {
			return acc
		}
	}
	return nil
}
