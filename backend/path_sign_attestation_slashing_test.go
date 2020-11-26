package backend

import (
	"context"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/bloxapp/eth2-key-manager/core"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/require"
	e2types "github.com/wealdtech/go-eth2-types/v2"

	"github.com/bloxapp/key-vault/backend/store"
)

func basicAttestationData() map[string]interface{} {
	return map[string]interface{}{
		"public_key":      "95087182937f6982ae99f9b06bd116f463f414513032e33a3d175d9662eddf162101fcf6ca2a9fedaded74b8047c5dcf",
		"domain":          "01000000f071c66c6561d0b939feb15f513a019d99a84bd85635221e3ad42dac",
		"slot":            284115,
		"committeeIndex":  2,
		"beaconBlockRoot": "7b5679277ca45ea74e1deebc9d3e8c0e7d6c570b3cfaf6884be144a81dac9a0e",
		"sourceEpoch":     77,
		"sourceRoot":      "7402fdc1ce16d449d637c34a172b349a12b2bae8d6d77e401006594d8057c33d",
		"targetEpoch":     78,
		"targetRoot":      "17959acc370274756fa5e9fdd7e7adf17204f49cc8457e49438c42c4883cbfb0",
	}
}

func updateWithBasicHighestAtt(storage logical.Storage) error {
	s := store.NewHashicorpVaultStore(context.Background(), storage, core.PyrmontNetwork)
	pkByts, _ := hex.DecodeString("95087182937f6982ae99f9b06bd116f463f414513032e33a3d175d9662eddf162101fcf6ca2a9fedaded74b8047c5dcf")
	pk, _ := e2types.BLSPublicKeyFromBytes(pkByts)
	return s.SaveHighestAttestation(pk, &core.BeaconAttestation{
		Source: &core.Checkpoint{
			Epoch: 0,
			Root:  nil,
		},
		Target: &core.Checkpoint{
			Epoch: 0,
			Root:  nil,
		},
	})
}

func TestAttestationSlashing(t *testing.T) {
	b, _ := getBackend(t)

	t.Run("Successfully Sign Attestation", func(t *testing.T) {
		req := logical.TestRequest(t, logical.CreateOperation, "accounts/sign-attestation")
		setupBaseStorage(t, req)

		// setup storage
		err := setupStorageWithWalletAndAccounts(req.Storage)
		require.NoError(t, err)
		require.NoError(t, updateWithBasicHighestAtt(req.Storage))

		req.Data = basicAttestationData()
		res, err := b.HandleRequest(context.Background(), req)
		fmt.Printf("res %#v\n", res)
		require.NoError(t, err)
		require.NotNil(t, res.Data)
		require.Equal(t,
			"866c25f50559c26d70a2918b29f1a1eb4e36a5cc4834ac114cec2e4ab59e8a69a23444a3a904512744caf0bfa34bd0510e436d4e1c9d66f17ed6f03a03a44baf18f8b164c69006b5538e0f4de904e66d470e5c73f0863aee220033f5f13ec050",
			res.Data["signature"],
		)
	})

	t.Run("Sign duplicated Attestation (exactly same), should NOT sign", func(t *testing.T) {
		req := logical.TestRequest(t, logical.CreateOperation, "accounts/sign-attestation")
		setupBaseStorage(t, req)

		// setup storage
		err := setupStorageWithWalletAndAccounts(req.Storage)
		require.NoError(t, err)
		require.NoError(t, updateWithBasicHighestAtt(req.Storage))

		// first attestation
		req.Data = basicAttestationData()
		res, err := b.HandleRequest(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, res.Data)

		// duplicated attestation
		req.Data = basicAttestationData()
		res, err = b.HandleRequest(context.Background(), req)
		require.EqualError(t, err, "failed to sign attestation: slashable attestation (HighestAttestationVote), not signing")
		require.Nil(t, res)
		//require.Equal(t,
		//	"b162e5ee4ee141738518f36707c6891ae33f929d12f5cb51ad989774ccb4b04d6aeabc2d4f68e8624b1743ef863cc72b10e090a0ae8d2d7f2abaf08372b9ab5c3312b31537bfebb858c41b41ebbe8dfffa1354185c6b4b881326b1bc7554f04d",
		//	res.Data["signature"],
		//)
	})

	t.Run("Sign double Attestation (different block root), should return error", func(t *testing.T) {
		req := logical.TestRequest(t, logical.CreateOperation, "accounts/sign-attestation")
		setupBaseStorage(t, req)

		// setup storage
		err := setupStorageWithWalletAndAccounts(req.Storage)
		require.NoError(t, err)
		require.NoError(t, updateWithBasicHighestAtt(req.Storage))

		// first attestation
		req.Data = basicAttestationData()
		res, err := b.HandleRequest(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, res.Data)

		// slashable attestation
		data := basicAttestationData()
		data["beaconBlockRoot"] = "7b5679277ca45ea74e1deebc9d3e8c0e7d6c570b3cfaf6884be144a81dac9a0f"
		req.Data = data
		res, err = b.HandleRequest(context.Background(), req)
		require.Error(t, err)
		require.EqualError(t, err, "failed to sign attestation: slashable attestation (HighestAttestationVote), not signing")
		require.Nil(t, res)
	})

	t.Run("Sign double Attestation (different source root), should return error", func(t *testing.T) {
		req := logical.TestRequest(t, logical.CreateOperation, "accounts/sign-attestation")
		setupBaseStorage(t, req)

		// setup storage
		err := setupStorageWithWalletAndAccounts(req.Storage)
		require.NoError(t, err)
		require.NoError(t, updateWithBasicHighestAtt(req.Storage))

		// first attestation
		req.Data = basicAttestationData()
		res, err := b.HandleRequest(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, res.Data)

		// slashable attestation
		data := basicAttestationData()
		data["sourceRoot"] = "7402fdc1ce16d449d637c34a172b349a12b2bae8d6d77e401006594d8057c33e"
		req.Data = data
		res, err = b.HandleRequest(context.Background(), req)
		require.Error(t, err)
		require.EqualError(t, err, "failed to sign attestation: slashable attestation (HighestAttestationVote), not signing")
		require.Nil(t, res)
	})

	t.Run("Sign double Attestation (different target root), should return error", func(t *testing.T) {
		req := logical.TestRequest(t, logical.CreateOperation, "accounts/sign-attestation")
		setupBaseStorage(t, req)

		// setup storage
		err := setupStorageWithWalletAndAccounts(req.Storage)
		require.NoError(t, err)
		require.NoError(t, updateWithBasicHighestAtt(req.Storage))

		// first attestation
		req.Data = basicAttestationData()
		res, err := b.HandleRequest(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, res.Data)

		// slashable attestation
		data := basicAttestationData()
		data["targetRoot"] = "17959acc370274756fa5e9fdd7e7adf17204f49cc8457e49438c42c4883cbfb1"
		req.Data = data
		res, err = b.HandleRequest(context.Background(), req)
		require.Error(t, err)
		require.EqualError(t, err, "failed to sign attestation: slashable attestation (HighestAttestationVote), not signing")
		require.Nil(t, res)
	})

	t.Run("Sign Attestation (different domain), should sign", func(t *testing.T) {
		req := logical.TestRequest(t, logical.CreateOperation, "accounts/sign-attestation")
		setupBaseStorage(t, req)

		// setup storage
		err := setupStorageWithWalletAndAccounts(req.Storage)
		require.NoError(t, err)
		require.NoError(t, updateWithBasicHighestAtt(req.Storage))

		// first attestation
		req.Data = basicAttestationData()
		res, err := b.HandleRequest(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, res.Data)

		// slashable attestation
		data := basicAttestationData()
		data["domain"] = "01000000f071c66c6561d0b939feb15f513a019d99a84bd85635221e3ad42dad"
		req.Data = data
		res, err = b.HandleRequest(context.Background(), req)
		require.EqualError(t, err, "failed to sign attestation: slashable attestation (HighestAttestationVote), not signing")
		require.Nil(t, res)
	})

	t.Run("Sign surrounding Attestation, should error", func(t *testing.T) {
		req := logical.TestRequest(t, logical.CreateOperation, "accounts/sign-attestation")
		setupBaseStorage(t, req)

		// setup storage
		err := setupStorageWithWalletAndAccounts(req.Storage)
		require.NoError(t, err)
		require.NoError(t, updateWithBasicHighestAtt(req.Storage))

		// first attestation
		req.Data = basicAttestationData()
		_, err = b.HandleRequest(context.Background(), req)
		require.NoError(t, err)

		// add another attestation building on the base
		// 77 <- 78 <- 79
		req.Data = map[string]interface{}{
			"public_key":      "95087182937f6982ae99f9b06bd116f463f414513032e33a3d175d9662eddf162101fcf6ca2a9fedaded74b8047c5dcf",
			"domain":          "01000000f071c66c6561d0b939feb15f513a019d99a84bd85635221e3ad42dac",
			"slot":            284116,
			"committeeIndex":  2,
			"beaconBlockRoot": "7b5679277ca45ea74e1deebc9d3e8c0e7d6c570b3cfaf6884be144a81dac9a0e",
			"sourceEpoch":     78,
			"sourceRoot":      "7402fdc1ce16d449d637c34a172b349a12b2bae8d6d77e401006594d8057c33d",
			"targetEpoch":     79,
			"targetRoot":      "17959acc370274756fa5e9fdd7e7adf17204f49cc8457e49438c42c4883cbfb0",
		}
		_, err = b.HandleRequest(context.Background(), req)
		require.NoError(t, err)

		// surround previous vote
		// 77 <- 78 <- 79
		// 	<------------ 80
		// slashable
		req.Data = map[string]interface{}{
			"public_key":      "95087182937f6982ae99f9b06bd116f463f414513032e33a3d175d9662eddf162101fcf6ca2a9fedaded74b8047c5dcf",
			"domain":          "01000000f071c66c6561d0b939feb15f513a019d99a84bd85635221e3ad42dac",
			"slot":            284117,
			"committeeIndex":  2,
			"beaconBlockRoot": "7b5679277ca45ea74e1deebc9d3e8c0e7d6c570b3cfaf6884be144a81dac9a0e",
			"sourceEpoch":     77,
			"sourceRoot":      "7402fdc1ce16d449d637c34a172b349a12b2bae8d6d77e401006594d8057c33d",
			"targetEpoch":     80,
			"targetRoot":      "17959acc370274756fa5e9fdd7e7adf17204f49cc8457e49438c42c4883cbfb0",
		}
		res, err := b.HandleRequest(context.Background(), req)
		require.Error(t, err)
		require.EqualError(t, err, "failed to sign attestation: slashable attestation (HighestAttestationVote), not signing")
		require.Nil(t, res)
	})

	t.Run("Sign surrounded Attestation, should error", func(t *testing.T) {
		req := logical.TestRequest(t, logical.CreateOperation, "accounts/sign-attestation")
		setupBaseStorage(t, req)

		// setup storage
		err := setupStorageWithWalletAndAccounts(req.Storage)
		require.NoError(t, err)
		require.NoError(t, updateWithBasicHighestAtt(req.Storage))

		// first attestation
		req.Data = basicAttestationData()
		_, err = b.HandleRequest(context.Background(), req)
		require.NoError(t, err)

		// add another attestation building on the base
		// 77 <- 78 <- 79 <----------------------90
		req.Data = map[string]interface{}{
			"public_key":      "95087182937f6982ae99f9b06bd116f463f414513032e33a3d175d9662eddf162101fcf6ca2a9fedaded74b8047c5dcf",
			"domain":          "01000000f071c66c6561d0b939feb15f513a019d99a84bd85635221e3ad42dac",
			"slot":            284116,
			"committeeIndex":  2,
			"beaconBlockRoot": "7b5679277ca45ea74e1deebc9d3e8c0e7d6c570b3cfaf6884be144a81dac9a0e",
			"sourceEpoch":     79,
			"sourceRoot":      "7402fdc1ce16d449d637c34a172b349a12b2bae8d6d77e401006594d8057c33d",
			"targetEpoch":     95,
			"targetRoot":      "17959acc370274756fa5e9fdd7e7adf17204f49cc8457e49438c42c4883cbfb0",
		}
		_, err = b.HandleRequest(context.Background(), req)
		require.NoError(t, err)

		// surround previous vote
		// 77 <- 78 <- 79 <----------------------95
		// 								88 <- 89
		// slashable
		req.Data = map[string]interface{}{
			"public_key":      "95087182937f6982ae99f9b06bd116f463f414513032e33a3d175d9662eddf162101fcf6ca2a9fedaded74b8047c5dcf",
			"domain":          "01000000f071c66c6561d0b939feb15f513a019d99a84bd85635221e3ad42dac",
			"slot":            284117,
			"committeeIndex":  2,
			"beaconBlockRoot": "7b5679277ca45ea74e1deebc9d3e8c0e7d6c570b3cfaf6884be144a81dac9a0e",
			"sourceEpoch":     89,
			"sourceRoot":      "7402fdc1ce16d449d637c34a172b349a12b2bae8d6d77e401006594d8057c33d",
			"targetEpoch":     88,
			"targetRoot":      "17959acc370274756fa5e9fdd7e7adf17204f49cc8457e49438c42c4883cbfb0",
		}
		res, err := b.HandleRequest(context.Background(), req)
		require.Error(t, err)
		require.EqualError(t, err, "failed to sign attestation: slashable attestation (HighestAttestationVote), not signing")
		require.Nil(t, res)
	})
}
