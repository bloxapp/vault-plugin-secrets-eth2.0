package backend

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/bloxapp/eth2-key-manager/core"
	"github.com/bloxapp/eth2-key-manager/stores/in_memory"
	"github.com/bloxapp/eth2-key-manager/wallet_hd"
	uuid "github.com/google/uuid"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/require"

	"github.com/bloxapp/key-vault/backend/store"
)

func _byteArray(input string) []byte {
	res, _ := hex.DecodeString(input)
	return res
}

func baseInmemStorage() (*in_memory.InMemStore, uuid.UUID, error) {
	inMemStore := in_memory.NewInMemStore(core.PyrmontNetwork)

	// wallet
	wallet := wallet_hd.NewHDWallet(&core.WalletContext{Storage: inMemStore})
	err := inMemStore.SaveWallet(wallet)
	if err != nil {
		return nil, uuid.UUID{}, err
	}

	// account
	acc, err := wallet.CreateValidatorAccount(
		_byteArray("0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1fff"),
		nil,
	)
	if err != nil {
		return nil, uuid.UUID{}, err
	}
	acc.ID()

	if err := inMemStore.SaveAccount(acc); err != nil {
		return nil, uuid.UUID{}, err
	}

	err = inMemStore.SaveHighestAttestation(acc.ValidatorPublicKey(), &core.BeaconAttestation{
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
		return nil, uuid.UUID{}, err
	}

	return inMemStore, acc.ID(), nil
}

func baseHashicorpStorage(logicalStorage logical.Storage, ctx context.Context) (*store.HashicorpVaultStore, error) {
	inMem, _, err := baseInmemStorage()
	if err != nil {
		return nil, err
	}
	return store.FromInMemoryStore(ctx, inMem, logicalStorage)
}

func TestStorage(t *testing.T) {
	b, _ := getBackend(t)
	inMemStore, _, err := baseInmemStorage()
	require.NoError(t, err)

	// marshal and to string
	byts, err := json.Marshal(inMemStore)
	require.NoError(t, err)
	data := hex.EncodeToString(byts)

	getVault := func(tt *testing.T) *store.HashicorpVaultStore {
		req := logical.TestRequest(tt, logical.CreateOperation, "storage")
		req.Data = map[string]interface{}{
			"data": data,
		}
		res, err := b.HandleRequest(context.Background(), req)
		require.NoError(tt, err)
		require.True(tt, res.Data["status"].(bool))
		return store.NewHashicorpVaultStore(context.Background(), req.Storage, core.PyrmontNetwork)
	}

	t.Run("verify wallet and account", func(t *testing.T) {
		vault := getVault(t)
		wallet, err := vault.OpenWallet()
		require.NoError(t, err)
		require.Equal(t, wallet.ID().String(), wallet.ID().String())

		_, err = wallet.AccountByPublicKey("95087182937f6982ae99f9b06bd116f463f414513032e33a3d175d9662eddf162101fcf6ca2a9fedaded74b8047c5dcf")
		require.NoError(t, err)
	})

	t.Run("verify highest att.", func(t *testing.T) {
		vault := getVault(t)
		wallet, err := vault.OpenWallet()
		require.NoError(t, err)
		acc, err := wallet.AccountByPublicKey("95087182937f6982ae99f9b06bd116f463f414513032e33a3d175d9662eddf162101fcf6ca2a9fedaded74b8047c5dcf")
		require.NoError(t, err)

		att := vault.RetrieveHighestAttestation(acc.ValidatorPublicKey())
		require.NotNil(t, att)
		require.EqualValues(t, att.Source.Epoch, 0)
		require.EqualValues(t, att.Target.Epoch, 0)
	})
}
