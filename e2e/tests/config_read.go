package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/bloxapp/eth2-key-manager/core"
	"github.com/stretchr/testify/require"

	"github.com/bloxapp/key-vault/e2e"
)

type configModel struct {
	Data struct {
		Network string `json:"network"`
	} `json:"data"`
}

// ConfigRead tests read config endpoint.
type ConfigRead struct {
}

// Name returns the name of the test.
func (test *ConfigRead) Name() string {
	return "Test read config endpoint"
}

// Run runs the test.
func (test *ConfigRead) Run(t *testing.T) {
	setup := e2e.Setup(t)

	// sign and save the valid aggregation
	configBytes, statusCode := setup.ReadConfig(t, core.PyrmontNetwork)
	require.Equal(t, http.StatusOK, statusCode)

	// parse to json
	var config configModel
	err := json.Unmarshal(configBytes, &config)
	require.NoError(t, err)
	require.EqualValues(t, core.PyrmontNetwork, config.Data.Network)
}
