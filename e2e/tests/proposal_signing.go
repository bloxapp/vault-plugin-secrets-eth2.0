package tests

import (
	"encoding/hex"
	"testing"

	"github.com/bloxapp/eth2-key-manager/core"
	"github.com/bloxapp/eth2-key-manager/slashing_protection"
	"github.com/bloxapp/eth2-key-manager/stores/in_memory"
	"github.com/bloxapp/eth2-key-manager/validator_signer"
	"github.com/stretchr/testify/require"
	v1 "github.com/wealdtech/eth2-signer-api/pb/v1"

	"github.com/bloxapp/key-vault/e2e"
	"github.com/bloxapp/key-vault/e2e/shared"
)

// ProposalSigning tests sign proposal endpoint.
type ProposalSigning struct {
}

// Name returns the name of the test.
func (test *ProposalSigning) Name() string {
	return "Test proposal signing"
}

// Run run the test.
func (test *ProposalSigning) Run(t *testing.T) {
	setup := e2e.Setup(t)

	// setup vault with db
	storage := setup.UpdateStorage(t, core.PyrmontNetwork, true, core.HDWallet, nil)
	account := shared.RetrieveAccount(t, storage)
	require.NotNil(t, account)
	pubKeyBytes := account.ValidatorPublicKey().Marshal()

	// Get wallet
	wallet, err := storage.OpenWallet()
	require.NoError(t, err)

	dataToSign := map[string]interface{}{
		"public_key":    hex.EncodeToString(pubKeyBytes),
		"domain":        "01000000f071c66c6561d0b939feb15f513a019d99a84bd85635221e3ad42dac",
		"slot":          78,
		"proposerIndex": 1010,
		"parentRoot":    "7b5679277ca45ea74e1deebc9d3e8c0e7d6c570b3cfaf6884be144a81dac9a0e",
		"stateRoot":     "7402fdc1ce16d449d637c34a172b349a12b2bae8d6d77e401006594d8057c33d",
		"bodyRoot":      "17959acc370274756fa5e9fdd7e7adf17204f49cc8457e49438c42c4883cbfb0",
	}

	// Sign data
	protector := slashing_protection.NewNormalProtection(in_memory.NewInMemStore(core.PyrmontNetwork))
	var signer validator_signer.ValidatorSigner = validator_signer.NewSimpleSigner(wallet, protector, storage.Network())

	res, err := signer.SignBeaconProposal(test.dataToProposalRequest(t, pubKeyBytes, dataToSign))
	require.NoError(t, err)

	// Send sign attestation request
	sig, err := setup.SignProposal(dataToSign, core.PyrmontNetwork)
	require.NoError(t, err)

	require.Equal(t, res.GetSignature(), sig)
}

func (test *ProposalSigning) dataToProposalRequest(t *testing.T, pubKey []byte, data map[string]interface{}) *v1.SignBeaconProposalRequest {
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
