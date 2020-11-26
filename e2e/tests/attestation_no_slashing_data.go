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

// AttestationNoSlashingDataSigning tests sign attestation endpoint.
type AttestationNoSlashingDataSigning struct {
}

// Name returns the name of the test.
func (test *AttestationNoSlashingDataSigning) Name() string {
	return "Test attestation no slashing data signing"
}

// Run run the test.
func (test *AttestationNoSlashingDataSigning) Run(t *testing.T) {
	setup := e2e.Setup(t)

	// setup vault with db
	storage := setup.UpdateStorage(t, core.PyrmontNetwork, false)
	account := shared.RetrieveAccount(t, storage)
	require.NotNil(t, account)
	pubKeyBytes := account.ValidatorPublicKey().Marshal()

	dataToSign := map[string]interface{}{
		"public_key":      hex.EncodeToString(pubKeyBytes),
		"domain":          "01000000f071c66c6561d0b939feb15f513a019d99a84bd85635221e3ad42dac",
		"slot":            284115,
		"committeeIndex":  2,
		"beaconBlockRoot": "7b5679277ca45ea74e1deebc9d3e8c0e7d6c570b3cfaf6884be144a81dac9a0e",
		"sourceEpoch":     77,
		"sourceRoot":      "7402fdc1ce16d449d637c34a172b349a12b2bae8d6d77e401006594d8057c33d",
		"targetEpoch":     78,
		"targetRoot":      "17959acc370274756fa5e9fdd7e7adf17204f49cc8457e49438c42c4883cbfb0",
	}

	// Send sign attestation request
	_, err := setup.SignAttestation(dataToSign, core.PyrmontNetwork)
	expectedErr := fmt.Sprintf("map[string]interface {}{\"errors\":[]interface {}{\"1 error occurred:\\n\\t* failed to sign attestation: highest attestation data is nil, can't determine if attestation is slashable\\n\\n\"}}")
	require.EqualError(t, err, expectedErr, fmt.Sprintf("actual: %s\n", err.Error()))
}

func (test *AttestationNoSlashingDataSigning) dataToAttestationRequest(t *testing.T, pubKey []byte, data map[string]interface{}) *v1.SignBeaconAttestationRequest {
	// Decode domain
	domainBytes, err := hex.DecodeString(data["domain"].(string))
	require.NoError(t, err)

	// Decode block root
	beaconBlockRoot, err := hex.DecodeString(data["beaconBlockRoot"].(string))
	require.NoError(t, err)

	// Decode source root
	sourceRootBytes, err := hex.DecodeString(data["sourceRoot"].(string))
	require.NoError(t, err)

	// Decode target root
	targetRootBytes, err := hex.DecodeString(data["targetRoot"].(string))
	require.NoError(t, err)

	return &v1.SignBeaconAttestationRequest{
		Id:     &v1.SignBeaconAttestationRequest_PublicKey{PublicKey: pubKey},
		Domain: domainBytes,
		Data: &v1.AttestationData{
			Slot:            uint64(data["slot"].(int)),
			CommitteeIndex:  uint64(data["committeeIndex"].(int)),
			BeaconBlockRoot: beaconBlockRoot,
			Source: &v1.Checkpoint{
				Epoch: uint64(data["sourceEpoch"].(int)),
				Root:  sourceRootBytes,
			},
			Target: &v1.Checkpoint{
				Epoch: uint64(data["targetEpoch"].(int)),
				Root:  targetRootBytes,
			},
		},
	}
}
