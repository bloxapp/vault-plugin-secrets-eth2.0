package tests

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/bloxapp/eth2-key-manager/core"
	"github.com/stretchr/testify/require"
	v1 "github.com/wealdtech/eth2-signer-api/pb/v1"

	"github.com/bloxapp/key-vault/e2e"
	"github.com/bloxapp/key-vault/e2e/shared"
)

// ProposalFarFutureSigning tests sign proposal endpoint with future signing.
type ProposalFarFutureSigning struct {
}

// Name returns the name of the test.
func (test *ProposalFarFutureSigning) Name() string {
	return "Test proposal far future signing"
}

// Run run the test.
func (test *ProposalFarFutureSigning) Run(t *testing.T) {
	setup := e2e.Setup(t)

	// setup vault with db
	storage := setup.UpdateStorage(t, core.PyrmontNetwork, true, core.HDWallet, nil)
	account := shared.RetrieveAccount(t, storage)
	require.NotNil(t, account)
	pubKeyBytes := account.ValidatorPublicKey().Marshal()

	dataToSign := map[string]interface{}{
		"public_key":    hex.EncodeToString(pubKeyBytes),
		"domain":        "01000000f071c66c6561d0b939feb15f513a019d99a84bd85635221e3ad42dac",
		"slot":          88878,
		"proposerIndex": 1010,
		"parentRoot":    "7b5679277ca45ea74e1deebc9d3e8c0e7d6c570b3cfaf6884be144a81dac9a0e",
		"stateRoot":     "7402fdc1ce16d449d637c34a172b349a12b2bae8d6d77e401006594d8057c33d",
		"bodyRoot":      "17959acc370274756fa5e9fdd7e7adf17204f49cc8457e49438c42c4883cbfb0",
	}

	// Send sign attestation request
	_, err := setup.SignProposal(dataToSign, core.PyrmontNetwork)
	require.NotNil(t, err)
	expectedErr := fmt.Sprintf("map[string]interface {}{\"errors\":[]interface {}{\"1 error occurred:\\n\\t* failed to sign data: proposed block slot too far into the future\\n\\n\"}}")
	require.EqualError(t, err, expectedErr, fmt.Sprintf("actual: %s\n", err.Error()))
}

func (test *ProposalFarFutureSigning) dataToProposalRequest(t *testing.T, pubKey []byte, data map[string]interface{}) *v1.SignBeaconProposalRequest {
	// Decode domain
	domainBytes, err := hex.DecodeString(data["domain"].(string))
	require.NoError(t, err)

	// Decode parent root
	parentRoot, err := hex.DecodeString(data["parentRoot"].(string))
	require.NoError(t, err)

	// Decode state root
	stateRoot, err := hex.DecodeString(data["stateRoot"].(string))
	require.NoError(t, err)

	// Decode body root
	bodyRoot, err := hex.DecodeString(data["bodyRoot"].(string))
	require.NoError(t, err)

	return &v1.SignBeaconProposalRequest{
		Id:     &v1.SignBeaconProposalRequest_PublicKey{PublicKey: pubKey},
		Domain: domainBytes,
		Data: &v1.BeaconBlockHeader{
			Slot:          uint64(data["slot"].(int)),
			ProposerIndex: uint64(data["proposerIndex"].(int)),
			ParentRoot:    parentRoot,
			StateRoot:     stateRoot,
			BodyRoot:      bodyRoot,
		},
	}
}
