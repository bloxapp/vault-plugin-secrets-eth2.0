package tests

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/bloxapp/eth2-key-manager/core"

	"github.com/stretchr/testify/require"

	"github.com/bloxapp/key-vault/e2e"
)

type configModel struct {
	Data struct {
		Network     string    `json:"network"`
		GenesisTime time.Time `json:"genesis_time"`
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

	expectedGenesisTime, err := time.Parse("2006-01-02 15:04:05 MST", "2020-11-18 12:00:07 UTC")
	require.NoError(t, err)

	// parse to json
	var config configModel
	err = json.Unmarshal(configBytes, &config)
	require.NoError(t, err)
	require.EqualValues(t, core.PyrmontNetwork, config.Data.Network)
	require.EqualValues(t, expectedGenesisTime, config.Data.GenesisTime)
}
