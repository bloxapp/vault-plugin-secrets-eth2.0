package tests

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"strconv"
	"testing"

	"github.com/bloxapp/eth2-key-manager/core"
	"github.com/stretchr/testify/require"

	"github.com/bloxapp/key-vault/e2e"
	"github.com/bloxapp/key-vault/e2e/shared"
)

// AttestationConcurrentSigning tests signing method concurrently.
type AttestationConcurrentSigning struct {
}

// Name returns the name of the test.
func (test *AttestationConcurrentSigning) Name() string {
	return "Test attestation concurrent signing"
}

// Run runs the test.
func (test *AttestationConcurrentSigning) Run(t *testing.T) {
	setup := e2e.Setup(t)

	// setup vault with db
	store := setup.UpdateStorage(t, core.PyrmontNetwork, true)
	account := shared.RetrieveAccount(t, store)
	pubKey := hex.EncodeToString(account.ValidatorPublicKey().Marshal())

	// sign and save the valid attestation
	_, err := setup.SignAttestation(
		map[string]interface{}{
			"public_key":      pubKey,
			"domain":          "01000000f071c66c6561d0b939feb15f513a019d99a84bd85635221e3ad42dac",
			"slot":            284115,
			"committeeIndex":  1,
			"beaconBlockRoot": "7b5679277ca45ea74e1deebc9d3e8c0e7d6c570b3cfaf6884be144a81dac9a0e",
			"sourceEpoch":     77,
			"sourceRoot":      "7402fdc1ce16d449d637c34a172b349a12b2bae8d6d77e401006594d8057c33d",
			"targetEpoch":     78,
			"targetRoot":      "17959acc370274756fa5e9fdd7e7adf17204f49cc8457e49438c42c4883cbfb0",
		},
		core.PyrmontNetwork,
	)
	require.NoError(t, err)

	// Send requests in parallel
	t.Run("concurrent signing", func(t *testing.T) {
		t.Parallel()
		for i := 0; i < 5; i++ {
			t.Run("concurrent signing "+strconv.Itoa(i), func(t *testing.T) {
				t.Parallel()
				runSlashableAttestation(t, setup, pubKey)
			})
		}
	})
}

// will return no error if trying to sign a slashable attestation will not work
func runSlashableAttestation(t *testing.T, setup *e2e.BaseSetup, pubKey string) {
	randomCommittee := func() int {
		max := 1000
		min := 2
		return rand.Intn(max-min) + min
	}

	_, err := setup.SignAttestation(
		map[string]interface{}{
			"public_key":      pubKey,
			"domain":          "01000000f071c66c6561d0b939feb15f513a019d99a84bd85635221e3ad42dac",
			"slot":            284115,
			"committeeIndex":  randomCommittee(),
			"beaconBlockRoot": "7b5679277ca45ea74e1deebc9d3e8c0e7d6c570b3cfaf6884be144a81dac9a0e",
			"sourceEpoch":     77,
			"sourceRoot":      "7402fdc1ce16d449d637c34a172b349a12b2bae8d6d77e401006594d8057c33d",
			"targetEpoch":     78,
			"targetRoot":      "17959acc370274756fa5e9fdd7e7adf17204f49cc8457e49438c42c4883cbfb0",
		},
		core.PyrmontNetwork,
	)
	require.Error(t, err, "did not slash")
	require.IsType(t, &e2e.ServiceError{}, err)

	errValue := err.(*e2e.ServiceError).ErrorValue()
	protected := errValue == fmt.Sprintf("1 error occurred:\n\t* failed to sign attestation: slashable attestation (HighestAttestationVote), not signing\n\n") ||
		errValue == fmt.Sprintf("1 error occurred:\n\t* locked\n\n")
	require.True(t, protected, err.Error())
}
