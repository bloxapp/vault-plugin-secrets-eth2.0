package tests

import (
	"encoding/hex"
	"testing"

	"github.com/bloxapp/eth2-key-manager/core"
	"github.com/stretchr/testify/require"
	v1 "github.com/wealdtech/eth2-signer-api/pb/v1"

	"github.com/bloxapp/key-vault/e2e"
)

// ProposalReferenceSigning tests sign proposal endpoint.
type ProposalReferenceSigning struct {
}

// Name returns the name of the test.
// tested against a block and sig generated from https://github.com/prysmaticlabs/prysm/blob/master/shared/testutil/block.go#L86
func (test *ProposalReferenceSigning) Name() string {
	return "Test proposal reference signing"
}

// Run run the test.
func (test *ProposalReferenceSigning) Run(t *testing.T) {
	setup := e2e.Setup(t)

	// setup vault with db
	storage := setup.UpdateStorage(t, core.PyrmontNetwork, true, core.NDWallet, _byteArray("5470813f7deef638dc531188ca89e36976d536f680e89849cd9077fd096e20bc"))
	// Get wallet
	wallet, err := storage.OpenWallet()
	require.NoError(t, err)
	account, err := wallet.AccountByPublicKey("a3862121db5914d7272b0b705e6e3c5336b79e316735661873566245207329c30f9a33d4fb5f5857fc6fd0a368186972")
	require.NoError(t, err)
	require.NotNil(t, account)
	pubKeyBytes := account.ValidatorPublicKey().Marshal()

	dataToSign := map[string]interface{}{
		"public_key":    hex.EncodeToString(pubKeyBytes),
		"domain":        "0000000081509579e35e84020ad8751eca180b44df470332d3ad17fc6fd52459",
		"slot":          1,
		"proposerIndex": 85,
		"parentRoot":    "3a43a4bf26fb5947e809c1f24f7dc6857c8ac007e535d48e6e4eca2122fd776b",
		"stateRoot":     "add5cb7e6a4bffd8ce7fe9697aed511661861e312ad546dcf5480159698f47a5",
		"bodyRoot":      "c7f94e8fd9d2bb7713ab5f8d50ae6d6530e61ae5ba0b8d2c585bb1a831ab796c",
	}
	// Send sign attestation request
	sig, err := setup.SignProposal(dataToSign, core.PyrmontNetwork)
	require.NoError(t, err)

	expectedSig := _byteArray("911ac2f6d74039279f16eee4cc46f4c6eea0ef9d18f0d9739b407c150c07ccb104c1c4b034ad46b25719bafc22fad05205975393000ea09636f5ce427814e2fe12ea72041099cc7f6ec249e504992dbf65e968ab448ddf4e124cbcbc722829b5")
	require.Equal(t, expectedSig, sig)
}

func (test *ProposalReferenceSigning) dataToProposalRequest(t *testing.T, pubKey []byte, data map[string]interface{}) *v1.SignBeaconProposalRequest {
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

func _byteArray(input string) []byte {
	res, _ := hex.DecodeString(input)
	return res
}
