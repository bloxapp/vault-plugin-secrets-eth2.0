package tests

import (
	"net/http"
	"testing"

	"github.com/bloxapp/eth2-key-manager/core"
	"github.com/stretchr/testify/require"

	"github.com/bloxapp/key-vault/e2e"
)

// AttestationSigningAccountNotFound tests sign attestation when account not found
type AttestationSigningAccountNotFound struct {
}

// Name returns the name of the test
func (test *AttestationSigningAccountNotFound) Name() string {
	return "Test attestation signing account not found"
}

// Run runs the test.
func (test *AttestationSigningAccountNotFound) Run(t *testing.T) {
	setup := e2e.Setup(t)

	// setup vault with db
	setup.UpdateStorage(t, core.PyrmontNetwork, true)

	// sign
	_, err := setup.SignAttestation(
		map[string]interface{}{
			"public_key":      "ab321d63b7b991107a5667bf4fe853a266c2baea87d33a41c7e39a5641bfd3b5434b76f1229d452acb45ba86284e3278", // this account is not found
			"domain":          "01000000f071c66c6561d0b939feb15f513a019d99a84bd85635221e3ad42dac",
			"slot":            284115,
			"committeeIndex":  2,
			"beaconBlockRoot": "7b5679277ca45ea74e1deebc9d3e8c0e7d6c570b3cfaf6884be144a81dac9a0e",
			"sourceEpoch":     8877,
			"sourceRoot":      "7402fdc1ce16d449d637c34a172b349a12b2bae8d6d77e401006594d8057c33d",
			"targetEpoch":     8878,
			"targetRoot":      "17959acc370274756fa5e9fdd7e7adf17204f49cc8457e49438c42c4883cbfb0",
		},
		core.PyrmontNetwork,
	)
	require.Error(t, err)
	require.IsType(t, &e2e.ServiceError{}, err)
	require.EqualValues(t, "account not found", err.(*e2e.ServiceError).DataValue("message"))
	require.EqualValues(t, http.StatusNotFound, err.(*e2e.ServiceError).DataValue("status_code"))
}
