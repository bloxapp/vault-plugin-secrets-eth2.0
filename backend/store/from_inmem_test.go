package store_test

import (
	"context"
	"encoding/hex"
	"testing"

	ethkeymanager "github.com/bloxapp/eth2-key-manager"
	"github.com/bloxapp/eth2-key-manager/core"
	"github.com/bloxapp/eth2-key-manager/encryptor/keystorev4"
	"github.com/bloxapp/eth2-key-manager/stores/in_memory"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/require"

	"github.com/bloxapp/key-vault/backend/store"
)

func _byteArray(input string) []byte {
	res, _ := hex.DecodeString(input)
	return res
}

func baseKeyVault(seed []byte, t *testing.T) (*in_memory.InMemStore, core.Wallet, []core.ValidatorAccount) {
	// store
	inMemStore := in_memory.NewInMemStore(core.TestNetwork)
	// seed
	// create keyvault in a normal in mem store
	options := &ethkeymanager.KeyVaultOptions{}
	options.SetStorage(inMemStore)
	options.SetSeed(seed)
	options.SetEncryptor(keystorev4.New())
	options.SetPassword("password")
	kv, err := ethkeymanager.NewKeyVault(options)
	require.NoError(t, err)
	require.NotNil(t, kv)
	// get wallet and accounts to compare
	inMemWallet, err := kv.Wallet()
	require.NoError(t, err)
	require.NotNil(t, inMemWallet)
	inMemAcc1, err := inMemWallet.CreateValidatorAccount(seed, nil)
	require.NoError(t, err)
	require.NotNil(t, inMemAcc1)
	inMemAcc2, err := inMemWallet.CreateValidatorAccount(seed, nil)
	require.NoError(t, err)
	require.NotNil(t, inMemAcc2)

	return inMemStore, inMemWallet, []core.ValidatorAccount{inMemAcc1, inMemAcc2}
}

func TestImportAndDeleteFromInMem(t *testing.T) {
	oldInMemStore, _, oldInMemAccounts := baseKeyVault(
		_byteArray("0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1fff"),
		t,
	)

	hashiStorage := &logical.InmemStorage{}

	// import to hashicorp
	oldHashi, err := store.FromInMemoryStore(context.Background(), oldInMemStore, hashiStorage)
	require.NoError(t, err)

	// create another in mem base keyvault to override (different seed and account indexes)
	inMemStore, inMemWallet, inMemAccounts := baseKeyVault(
		_byteArray("0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1fdf"),
		t,
	)

	// import to hashicorp, should override
	hashi, err := store.FromInMemoryStore(context.Background(), inMemStore, hashiStorage)
	require.NoError(t, err)

	// verify deletion
	// accounts fetched should no longer match old accounts
	res, err := oldHashi.OpenAccount(oldInMemAccounts[0].ID())
	require.Nil(t, res)
	res, err = oldHashi.OpenAccount(oldInMemAccounts[1].ID())
	require.Nil(t, res)

	// get hasicorp's wallet and accounts
	hashiWallet, err := hashi.OpenWallet()
	require.NoError(t, err)
	require.NotNil(t, hashiWallet)
	hashiAcc3, err := hashiWallet.AccountByPublicKey("800ab4ff899fd668cb9ea4712cc6414d61cb41cdca0a302bf5a1829f35bbdf3277a9e9a09693529b4e8175d6a7e0d5cb")
	require.NoError(t, err)
	require.NotNil(t, hashiAcc3)
	hashiAcc4, err := hashiWallet.AccountByPublicKey("b0c8e53f65c3e0ec8dc3b929f0d664481a905fa976c1022d6d1f3c8be8594fb3d42ed1edafe3da642efe7d76819598a1")
	require.NoError(t, err)
	require.NotNil(t, hashiAcc4)

	// compare
	require.Equal(t, inMemWallet.ID().String(), hashiWallet.ID().String())
	require.Equal(t, inMemAccounts[1].ID().String(), hashiAcc3.ID().String())
	require.Equal(t, inMemAccounts[1].ValidatorPublicKey().Marshal(), hashiAcc3.ValidatorPublicKey().Marshal())
	require.Equal(t, inMemAccounts[0].ID().String(), hashiAcc4.ID().String())
	require.Equal(t, inMemAccounts[0].ValidatorPublicKey().Marshal(), hashiAcc4.ValidatorPublicKey().Marshal())
}

func TestImportFromInMem(t *testing.T) {
	inMemStore, inMemWallet, inMemAccounts := baseKeyVault(
		_byteArray("0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1fff"),
		t,
	)

	// import to hashicorp
	hashi, err := store.FromInMemoryStore(context.Background(), inMemStore, &logical.InmemStorage{})
	require.NoError(t, err)

	// get hasicorp's wallet and accounts
	hashiWallet, err := hashi.OpenWallet()
	require.NoError(t, err)
	require.NotNil(t, hashiWallet)
	hashiAcc1, err := hashiWallet.AccountByPublicKey("b41df3c322a6fd305fc9425df52501f7f8067dbba551466d82d506c83c6ab287580202aa1a3449f54b9bc464a04b70e6")
	require.NoError(t, err)
	require.NotNil(t, hashiAcc1)
	hashiAcc2, err := hashiWallet.AccountByPublicKey("95087182937f6982ae99f9b06bd116f463f414513032e33a3d175d9662eddf162101fcf6ca2a9fedaded74b8047c5dcf")
	require.NoError(t, err)
	require.NotNil(t, hashiAcc2)

	// compare
	require.Equal(t, inMemWallet.ID().String(), hashiWallet.ID().String())
	require.Equal(t, inMemAccounts[1].ID().String(), hashiAcc1.ID().String())
	require.Equal(t, inMemAccounts[1].ValidatorPublicKey().Marshal(), hashiAcc1.ValidatorPublicKey().Marshal())
	require.Equal(t, inMemAccounts[0].ID().String(), hashiAcc2.ID().String())
	require.Equal(t, inMemAccounts[0].ValidatorPublicKey().Marshal(), hashiAcc2.ValidatorPublicKey().Marshal())
}
