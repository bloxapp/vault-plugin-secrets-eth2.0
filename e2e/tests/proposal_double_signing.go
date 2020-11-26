package tests

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/bloxapp/eth2-key-manager/core"
	"github.com/stretchr/testify/require"

	"github.com/bloxapp/key-vault/e2e"
	"github.com/bloxapp/key-vault/e2e/shared"
)

// ProposalDoubleSigning tests signing method concurrently.
type ProposalDoubleSigning struct {
}

// Name returns the name of the test.
func (test *ProposalDoubleSigning) Name() string {
	return "Test proposal double signing"
}

// Run runs the test.
func (test *ProposalDoubleSigning) Run(t *testing.T) {
	setup := e2e.Setup(t)

	// setup vault with db
	store := setup.UpdateStorage(t, core.PyrmontNetwork, true)
	account := shared.RetrieveAccount(t, store)
	pubKey := hex.EncodeToString(account.ValidatorPublicKey().Marshal())

	// Sign and save the valid proposal
	_, err := setup.SignProposal(
		map[string]interface{}{
			"public_key":    pubKey,
			"domain":        "01000000f071c66c6561d0b939feb15f513a019d99a84bd85635221e3ad42dac",
			"slot":          78,
			"proposerIndex": 1,
			"parentRoot":    "7b5679277ca45ea74e1deebc9d3e8c0e7d6c570b3cfaf6884be144a81dac9a0e",
			"stateRoot":     "7402fdc1ce16d449d637c34a172b349a12b2bae8d6d77e401006594d8057c33d",
			"bodyRoot":      "17959acc370274756fa5e9fdd7e7adf17204f49cc8457e49438c42c4883cbfb0",
		},
		core.PyrmontNetwork,
	)
	require.NoError(t, err)

	// Sign and save the slashable proposal
	_, err = setup.SignProposal(
		map[string]interface{}{
			"public_key":    pubKey,
			"domain":        "01000000f071c66c6561d0b939feb15f513a019d99a84bd85635221e3ad42dac",
			"slot":          78,
			"proposerIndex": 1231,
			"parentRoot":    "7b5679277ca45ea74e1deebc9d3e8c0e7d6c570b3cfaf6884be144a81dac9a0e",
			"stateRoot":     "7402fdc1ce16d449d637c34a172b349a12b2bae8d6d77e401006594d8057c33d",
			"bodyRoot":      "17959acc370274756fa5e9fdd7e7adf17204f49cc8457e49438c42c4883cbfb0",
		},
		core.PyrmontNetwork,
	)
	require.Error(t, err, "did not slash")
	require.IsType(t, &e2e.ServiceError{}, err)

	errValue := err.(*e2e.ServiceError).ErrorValue()
	protected := errValue == fmt.Sprintf("1 error occurred:\n\t* failed to sign data: err, slashable proposal: DoubleProposal\n\n")
	require.True(t, protected, err.Error())
}
