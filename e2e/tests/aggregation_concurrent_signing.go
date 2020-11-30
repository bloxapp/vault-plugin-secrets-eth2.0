package tests

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"testing"

	"github.com/bloxapp/eth2-key-manager/core"
	"github.com/stretchr/testify/require"

	"github.com/bloxapp/key-vault/e2e"
	"github.com/bloxapp/key-vault/e2e/shared"
)

// AggregationConcurrentSigning tests aggregation signing method concurrently.
type AggregationConcurrentSigning struct {
}

// Name returns the name of the test.
func (test *AggregationConcurrentSigning) Name() string {
	return "Test aggregation concurrent signing"
}

// Run runs the test.
func (test *AggregationConcurrentSigning) Run(t *testing.T) {
	setup := e2e.Setup(t)

	// setup vault with db
	store := setup.UpdateStorage(t, core.PyrmontNetwork, true, core.HDWallet, nil)
	account := shared.RetrieveAccount(t, store)
	pubKey := hex.EncodeToString(account.ValidatorPublicKey().Marshal())

	// sign and save the valid aggregation
	_, err := setup.SignAggregation(
		map[string]interface{}{
			"public_key": pubKey,
			"domain":     "01000000f071c66c6561d0b939feb15f513a019d99a84bd85635221e3ad42dac",
			"dataToSign": "7402fdc1ce16d449d637c34a172b349a12b2bae8d6d77e401006594d8057c33d",
		},
		core.PyrmontNetwork,
	)
	require.NoError(t, err)

	// Send requests in parallel
	t.Run("concurrent signing", func(t *testing.T) {
		t.Parallel()
		for i := 0; i < 5; i++ {
			i := i
			t.Run("concurrent signing "+strconv.Itoa(i), func(t *testing.T) {
				t.Parallel()

				_, err := setup.SignAggregation(
					map[string]interface{}{
						"public_key": pubKey,
						"domain":     "17959acc370274756fa5e9fdd7e7adf17204f49cc8457e49438c42c4883cbfb0",
						"dataToSign": "7b5679277ca45ea74e1deebc9d3e8c0e7d6c570b3cfaf6884be144a81dac9a0e",
					},
					core.PyrmontNetwork,
				)
				if err == nil {
					return
				}

				require.Error(t, err)
				require.IsType(t, &e2e.ServiceError{}, err)

				errValue := err.(*e2e.ServiceError).ErrorValue()
				protected := errValue == fmt.Sprintf("1 error occurred:\n\t* locked\n\n")
				require.True(t, protected, err.Error())
			})
		}
	})
}
