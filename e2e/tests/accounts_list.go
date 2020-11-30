package tests

import (
	"encoding/hex"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/bloxapp/eth2-key-manager/core"
	"github.com/stretchr/testify/require"

	"github.com/bloxapp/key-vault/e2e"
	"github.com/bloxapp/key-vault/e2e/shared"
)

type accountModel struct {
	Data struct {
		Accounts []struct {
			ID               string `json:"id"`
			Name             string `json:"name"`
			ValidationPubKey string `json:"validationPubKey"`
			WithdrawalPubKey string `json:"withdrawalPubKey"`
		} `json:"accounts"`
	} `json:"data"`
}

// AccountsList tests accounts list endpoint.
type AccountsList struct {
}

// Name returns the name of the test.
func (test *AccountsList) Name() string {
	return "Test listing accounts"
}

// Run runs the test.
func (test *AccountsList) Run(t *testing.T) {
	setup := e2e.Setup(t)

	// setup vault with db
	store := setup.UpdateStorage(t, core.PyrmontNetwork, true, core.HDWallet, nil)
	account := shared.RetrieveAccount(t, store)
	pubKey := hex.EncodeToString(account.ValidatorPublicKey().Marshal())
	withdrawalKey := hex.EncodeToString(account.WithdrawalPublicKey().Marshal())

	// sign and save the valid aggregation
	accountsBytes, statusCode := setup.ListAccounts(t, core.PyrmontNetwork)
	require.Equal(t, http.StatusOK, statusCode)

	// parse to json
	var accounts accountModel
	err := json.Unmarshal(accountsBytes, &accounts)
	require.NoError(t, err)
	require.Len(t, accounts.Data.Accounts, 1)
	require.NotEmpty(t, accounts.Data.Accounts[0].ID)
	require.NotEmpty(t, accounts.Data.Accounts[0].Name)
	require.Equal(t, pubKey, accounts.Data.Accounts[0].ValidationPubKey)
	require.Equal(t, withdrawalKey, accounts.Data.Accounts[0].WithdrawalPubKey)
}
