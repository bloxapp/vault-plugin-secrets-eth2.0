package tests

import (
	"encoding/hex"
	"testing"

	eth "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	validatorpb "github.com/prysmaticlabs/prysm/proto/validator/accounts/v2"

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
	setup.UpdateStorage(t, core.PyrmontNetwork, true, core.HDWallet, nil)

	// sign
	att := &eth.AttestationData{
		Slot:            284115,
		CommitteeIndex:  2,
		BeaconBlockRoot: _byteArray32("7b5679277ca45ea74e1deebc9d3e8c0e7d6c570b3cfaf6884be144a81dac9a0e"),
		Source: &eth.Checkpoint{
			Epoch: 5,
			Root:  _byteArray32("7402fdc1ce16d449d637c34a172b349a12b2bae8d6d77e401006594d8057c33d"),
		},
		Target: &eth.Checkpoint{
			Epoch: 6,
			Root:  _byteArray32("17959acc370274756fa5e9fdd7e7adf17204f49cc8457e49438c42c4883cbfb0"),
		},
	}
	domain := _byteArray32("01000000f071c66c6561d0b939feb15f513a019d99a84bd85635221e3ad42dac")
	pubKey := _byteArray("ab321d63b7b991107a5667bf4fe853a266c2baea87d33a41c7e39a5641bfd3b5434b76f1229d452acb45ba86284e3278") // this account is not found
	req, err := test.serializedReq(pubKey, nil, domain, att)
	require.NoError(t, err)

	// send
	res, err := setup.Sign("sign", req, core.PyrmontNetwork)
	require.Error(t, err)
	require.IsType(t, &e2e.ServiceError{}, err)
	require.EqualValues(t, "1 error occurred:\n\t* failed to sign: account not found\n\n", err.(*e2e.ServiceError).ErrorValue())
	require.Nil(t, res)
}

func (test *AttestationSigningAccountNotFound) serializedReq(pk, root, domain []byte, attestation *eth.AttestationData) (map[string]interface{}, error) {
	req := &validatorpb.SignRequest{
		PublicKey:       pk,
		SigningRoot:     root,
		SignatureDomain: domain,
		Object:          &validatorpb.SignRequest_AttestationData{AttestationData: attestation},
	}

	byts, err := req.Marshal()
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"sign_req": hex.EncodeToString(byts),
	}, nil
}
