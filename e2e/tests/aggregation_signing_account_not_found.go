package tests

import (
	"net/http"
	"testing"

	"github.com/bloxapp/eth2-key-manager/core"
	"github.com/stretchr/testify/require"

	"github.com/bloxapp/key-vault/e2e"
)

// AggregationSigningAccountNotFound tests sign aggregation when account not found
type AggregationSigningAccountNotFound struct {
}

// Name returns the name of the test
func (test *AggregationSigningAccountNotFound) Name() string {
	return "Test aggregation signing account not found"
}

// Run runs the test.
func (test *AggregationSigningAccountNotFound) Run(t *testing.T) {
	setup := e2e.Setup(t)

	// setup vault with db
	setup.UpdateStorage(t, core.PyrmontNetwork, true, core.HDWallet, nil)

	_, err := setup.SignAggregation(
		map[string]interface{}{
			"public_key": "ab321d63b7b991107a5667bf4fe853a266c2baea87d33a41c7e39a5641bfd3b5434b76f1229d452acb45ba86284e3278",
			"domain":     "17959acc370274756fa5e9fdd7e7adf17204f49cc8457e49438c42c4883cbfb0",
			"dataToSign": "7b5679277ca45ea74e1deebc9d3e8c0e7d6c570b3cfaf6884be144a81dac9a0e",
		},
		core.PyrmontNetwork,
	)
	require.Error(t, err)
	require.IsType(t, &e2e.ServiceError{}, err)
	require.EqualValues(t, "account not found", err.(*e2e.ServiceError).DataValue("message"))
	require.EqualValues(t, http.StatusNotFound, err.(*e2e.ServiceError).DataValue("status_code"))
}
