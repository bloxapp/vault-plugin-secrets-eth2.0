package tests

import (
	"encoding/hex"
	"testing"

	"github.com/bloxapp/eth2-key-manager/core"
	"github.com/stretchr/testify/require"
	v1 "github.com/wealdtech/eth2-signer-api/pb/v1"

	"github.com/bloxapp/key-vault/e2e"
)

// RandaoReferenceSigning tests sign proposal endpoint.
type RandaoReferenceSigning struct {
}

// Name returns the name of the test.
// tested against a block and sig generated from https://github.com/prysmaticlabs/prysm/blob/master/shared/testutil/block.go#L86
func (test *RandaoReferenceSigning) Name() string {
	return "Test randao reference signing"
}

// Run run the test.
func (test *RandaoReferenceSigning) Run(t *testing.T) {
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
		"public_key": hex.EncodeToString(pubKeyBytes),
		"domain":     "0200000081509579e35e84020ad8751eca180b44df470332d3ad17fc6fd52459",
		"dataToSign": "0000000000000000000000000000000000000000000000000000000000000000", //"db89df916cfb2b957b61752be4a366148d74eae5a2c75d6cb48969c673b506b6",
	}
	// Send sign attestation request
	sig, err := setup.SignAggregation(dataToSign, core.PyrmontNetwork)
	require.NoError(t, err)

	expectedSig := _byteArray("a2c156a4bc9439f1d85f922f2abaa96e830f1c526101211bdb7d16f4ad9490a0302fc5adb089c05b5f16fd465962f47c04fc2b81a94d135a07c1613db61511c17284b51fafab984e56d3411e16e45f5068f146d9412f91d31ab0f237eac3d745")
	require.Equal(t, expectedSig, sig)
}

func (test *RandaoReferenceSigning) dataToProposalRequest(t *testing.T, pubKey []byte, data map[string]interface{}) *v1.SignBeaconProposalRequest {
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
