package backend

import (
	"context"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/require"
)

func setupStorageWithWalletAndAccounts(storage logical.Storage) error {
	_, err := baseHashicorpStorage(storage, context.Background())
	return err
}

func TestSignAttestation(t *testing.T) {
	b, _ := getBackend(t)

	t.Run("Sign Attestation in non existing key vault", func(t *testing.T) {
		req := logical.TestRequest(t, logical.CreateOperation, "accounts/sign-attestation")
		setupBaseStorage(t, req)
		_, err := b.HandleRequest(context.Background(), req)
		require.EqualError(t, err, "failed to open key vault: wallet not found")
	})

	t.Run("Sign Attestation of unknown account", func(t *testing.T) {
		req := logical.TestRequest(t, logical.CreateOperation, "accounts/sign-attestation")
		setupBaseStorage(t, req)

		// setup storage
		err := setupStorageWithWalletAndAccounts(req.Storage)
		require.NoError(t, err)

		data := map[string]interface{}{
			"public_key": "ab321d63b7b991107a5667bf4fe853a266c2baea87d33a41c7e39a5641bfd3b5434b76f1229d452acb45ba86284e3270",
		}
		req.Data = data
		resp, err := b.HandleRequest(context.Background(), req)
		require.NoError(t, err)
		require.EqualValues(t, 404, resp.Data["http_status_code"], resp.Data)
	})
}

func TestSignProposal(t *testing.T) {
	b, _ := getBackend(t)

	t.Run("Successfully Sign Proposal", func(t *testing.T) {
		req := logical.TestRequest(t, logical.CreateOperation, "accounts/sign-proposal")
		setupBaseStorage(t, req)

		// setup storage
		err := setupStorageWithWalletAndAccounts(req.Storage)
		require.NoError(t, err)

		data := map[string]interface{}{
			"public_key": "95087182937f6982ae99f9b06bd116f463f414513032e33a3d175d9662eddf162101fcf6ca2a9fedaded74b8047c5dcf",
		}
		req.Data = data
		res, err := b.HandleRequest(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, res.Data)
	})

	t.Run("Sign Proposal in non existing key vault", func(t *testing.T) {
		req := logical.TestRequest(t, logical.CreateOperation, "accounts/sign-proposal")
		setupBaseStorage(t, req)
		_, err := b.HandleRequest(context.Background(), req)
		require.EqualError(t, err, "failed to open key vault: wallet not found")
	})

	t.Run("Sign Proposal of unknown account", func(t *testing.T) {
		req := logical.TestRequest(t, logical.CreateOperation, "accounts/sign-proposal")
		setupBaseStorage(t, req)

		// setup storage
		err := setupStorageWithWalletAndAccounts(req.Storage)
		require.NoError(t, err)

		data := map[string]interface{}{
			"public_key": "ab321d63b7b991107a5667bf4fe853a266c2baea87d33a41c7e39a5641bfd3b5434b76f1229d452acb45ba86284e3270",
		}
		req.Data = data
		resp, err := b.HandleRequest(context.Background(), req)
		require.NoError(t, err)
		require.EqualValues(t, 404, resp.Data["http_status_code"], resp.Data)
	})
}

func TestSignAggregation(t *testing.T) {
	b, _ := getBackend(t)

	t.Run("Successfully Sign Aggregation", func(t *testing.T) {
		req := logical.TestRequest(t, logical.CreateOperation, "accounts/sign-aggregation")
		setupBaseStorage(t, req)

		// setup storage
		err := setupStorageWithWalletAndAccounts(req.Storage)
		require.NoError(t, err)

		data := map[string]interface{}{
			"public_key": "95087182937f6982ae99f9b06bd116f463f414513032e33a3d175d9662eddf162101fcf6ca2a9fedaded74b8047c5dcf",
		}
		req.Data = data
		_, err = b.HandleRequest(context.Background(), req)
		require.NoError(t, err)
	})

	t.Run("Sign Aggregation in non existing key vault", func(t *testing.T) {
		req := logical.TestRequest(t, logical.CreateOperation, "accounts/sign-aggregation")
		setupBaseStorage(t, req)
		_, err := b.HandleRequest(context.Background(), req)
		require.EqualError(t, err, "failed to open key vault: wallet not found")
	})

	t.Run("Sign Aggregation of unknown account", func(t *testing.T) {
		req := logical.TestRequest(t, logical.CreateOperation, "accounts/sign-aggregation")
		setupBaseStorage(t, req)

		// setup storage
		err := setupStorageWithWalletAndAccounts(req.Storage)
		require.NoError(t, err)

		data := map[string]interface{}{
			"public_key": "ab321d63b7b991107a5667bf4fe853a266c2baea87d33a41c7e39a5641bfd3b5434b76f1229d452acb45ba86284e3270",
		}
		req.Data = data
		resp, err := b.HandleRequest(context.Background(), req)
		require.NoError(t, err)
		require.EqualValues(t, 404, resp.Data["http_status_code"], resp.Data)
	})
}
