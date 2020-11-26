package tests

import (
	"encoding/hex"
	"testing"

	"github.com/bloxapp/eth2-key-manager/core"
	"github.com/stretchr/testify/require"

	"github.com/bloxapp/key-vault/e2e"
	"github.com/bloxapp/key-vault/e2e/shared"
)

// AggregationDoubleSigning tests aggregation signing method concurrently.
type AggregationDoubleSigning struct {
}

// Name returns the name of the test.
func (test *AggregationDoubleSigning) Name() string {
	return "Test aggregation double signing"
}

// Run runs the test.
func (test *AggregationDoubleSigning) Run(t *testing.T) {
	setup := e2e.Setup(t)

	// setup vault with db
	store := setup.UpdateStorage(t, core.PyrmontNetwork, true)
	account := shared.RetrieveAccount(t, store)
	pubKey := hex.EncodeToString(account.ValidatorPublicKey().Marshal())

	_, err := setup.SignAggregation(
		map[string]interface{}{
			"public_key": pubKey,
			"domain":     "17959acc370274756fa5e9fdd7e7adf17204f49cc8457e49438c42c4883cbfb0",
			"dataToSign": "7b5679277ca45ea74e1deebc9d3e8c0e7d6c570b3cfaf6884be144a81dac9a0e",
		},
		core.PyrmontNetwork,
	)
	require.NoError(t, err)

	_, err = setup.SignAggregation(
		map[string]interface{}{
			"public_key": pubKey,
			"domain":     "17959acc370274756fa5e9fdd7e7adf17204f49cc8457e49438c42c4883cbfb0",
			"dataToSign": "7b5679277ca45ea74e1deebc9d3e8c0e7d6c570b3cfaf6884be144a81dac9a0e",
		},
		core.PyrmontNetwork,
	)
	require.NoError(t, err)
}
