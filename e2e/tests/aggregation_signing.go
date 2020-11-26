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

// AggregationSigning tests sign aggregation endpoint.
type AggregationSigning struct {
}

// Name returns the name of the test.
func (test *AggregationSigning) Name() string {
	return "Test aggregation signing"
}

// Run run the test.
func (test *AggregationSigning) Run(t *testing.T) {
	setup := e2e.Setup(t)

	// setup vault with db
	storage := setup.UpdateStorage(t, core.PyrmontNetwork, true)
	account := shared.RetrieveAccount(t, storage)
	require.NotNil(t, account)
	pubKeyBytes := account.ValidatorPublicKey().Marshal()

	// Get wallet
	wallet, err := storage.OpenWallet()
	require.NoError(t, err)

	dataToSign := map[string]interface{}{
		"public_key": hex.EncodeToString(pubKeyBytes),
		"domain":     "01000000f071c66c6561d0b939feb15f513a019d99a84bd85635221e3ad42dac",
		"dataToSign": "7402fdc1ce16d449d637c34a172b349a12b2bae8d6d77e401006594d8057c33d",
	}

	// Sign data
	protector := slashing_protection.NewNormalProtection(in_memory.NewInMemStore(core.PyrmontNetwork))
	var signer validator_signer.ValidatorSigner = validator_signer.NewSimpleSigner(wallet, protector, storage.Network())

	res, err := signer.Sign(test.dataToSignRequest(t, pubKeyBytes, dataToSign))
	require.NoError(t, err)

	// Send sign attestation request
	sig, err := setup.SignAggregation(dataToSign, core.PyrmontNetwork)
	require.NoError(t, err)

	require.Equal(t, res.GetSignature(), sig)
}

func (test *AggregationSigning) dataToSignRequest(t *testing.T, pubKey []byte, data map[string]interface{}) *v1.SignRequest {
	// Decode domain
	domainBytes, err := hex.DecodeString(data["domain"].(string))
	require.NoError(t, err)

	// Decode data to sign
	dataToSign, err := hex.DecodeString(data["dataToSign"].(string))
	require.NoError(t, err)

	return &v1.SignRequest{
		Id:     &v1.SignRequest_PublicKey{PublicKey: pubKey},
		Domain: domainBytes,
		Data:   dataToSign,
	}
}
