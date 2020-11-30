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

type slashingHistoryModel struct {
	Data map[string]string `json:"data"`
}

// SlashingStorageRead tests slashing storage reading endpoint.
type SlashingStorageRead struct {
}

// Name returns the name of the test.
func (test *SlashingStorageRead) Name() string {
	return "Test slashing storage read"
}

// Run run the test.
func (test *SlashingStorageRead) Run(t *testing.T) {
	setup := e2e.Setup(t)

	// setup vault with db
	storage := setup.UpdateStorage(t, core.PyrmontNetwork, true, core.HDWallet, nil)
	account := shared.RetrieveAccount(t, storage)
	require.NotNil(t, account)
	pubKey := hex.EncodeToString(account.ValidatorPublicKey().Marshal())

	// Send sign proposal request
	_, err := setup.SignProposal(map[string]interface{}{
		"public_key":    pubKey,
		"domain":        "01000000f071c66c6561d0b939feb15f513a019d99a84bd85635221e3ad42dac",
		"slot":          78,
		"proposerIndex": 1010,
		"parentRoot":    "7b5679277ca45ea74e1deebc9d3e8c0e7d6c570b3cfaf6884be144a81dac9a0e",
		"stateRoot":     "7402fdc1ce16d449d637c34a172b349a12b2bae8d6d77e401006594d8057c33d",
		"bodyRoot":      "17959acc370274756fa5e9fdd7e7adf17204f49cc8457e49438c42c4883cbfb0",
	}, core.PyrmontNetwork)
	require.NoError(t, err)

	// Read slashing storage
	storageBytes, statusCode := setup.ReadSlashingStorage(t, core.PyrmontNetwork)
	require.Equal(t, http.StatusOK, statusCode, string(storageBytes))

	var slashingHistory slashingHistoryModel
	err = json.Unmarshal(storageBytes, &slashingHistory)
	require.NoError(t, err)

	pubKeyHistory, ok := slashingHistory.Data[pubKey]
	require.True(t, ok)
	require.NotEmpty(t, pubKeyHistory)
}
