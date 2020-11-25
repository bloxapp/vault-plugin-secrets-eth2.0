package backend

import (
	"context"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/require"
)

func basicProposalData() map[string]interface{} {
	return map[string]interface{}{
		"public_key":    "95087182937f6982ae99f9b06bd116f463f414513032e33a3d175d9662eddf162101fcf6ca2a9fedaded74b8047c5dcf",
		"domain":        "01000000f071c66c6561d0b939feb15f513a019d99a84bd85635221e3ad42dac",
		"slot":          88,
		"proposerIndex": 1,
		"parentRoot":    "7402fdc1ce16d449d637c34a172b349a12b2bae8d6d77e401006594d8057c33d",
		"stateRoot":     "17959acc370274756fa5e9fdd7e7adf17204f49cc8457e49438c42c4883cbfb0",
		"bodyRoot":      "7b5679277ca45ea74e1deebc9d3e8c0e7d6c570b3cfaf6884be144a81dac9a0e",
	}
}

func TestProposalSlashing(t *testing.T) {
	b, _ := getBackend(t)

	t.Run("Successfully Sign proposal", func(t *testing.T) {
		req := logical.TestRequest(t, logical.CreateOperation, "accounts/sign-proposal")
		setupBaseStorage(t, req)

		// setup storage
		err := setupStorageWithWalletAndAccounts(req.Storage)
		require.NoError(t, err)

		req.Data = basicProposalData()
		res, err := b.HandleRequest(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, res.Data)
	})

	t.Run("Sign proposal with undefined account", func(t *testing.T) {
		req := logical.TestRequest(t, logical.CreateOperation, "accounts/sign-proposal")
		setupBaseStorage(t, req)

		// setup storage
		err := setupStorageWithWalletAndAccounts(req.Storage)
		require.NoError(t, err)

		req.Data = basicProposalData()
		req.Data["public_key"] = "undefinedpublickey"
		res, err := b.HandleRequest(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, res.Data)
		require.Equal(t, 404, res.Data["http_status_code"], res.Data)
	})

	t.Run("Successfully Sign proposal (exactly same)", func(t *testing.T) {
		req := logical.TestRequest(t, logical.CreateOperation, "accounts/sign-proposal")
		setupBaseStorage(t, req)

		// setup storage
		err := setupStorageWithWalletAndAccounts(req.Storage)
		require.NoError(t, err)

		// first proposal
		req.Data = basicProposalData()
		_, err = b.HandleRequest(context.Background(), req)
		require.NoError(t, err)

		// second proposal
		req.Data = basicProposalData()
		res, err := b.HandleRequest(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, res.Data)
	})

	t.Run("Sign double proposal(different state root), should error", func(t *testing.T) {
		req := logical.TestRequest(t, logical.CreateOperation, "accounts/sign-proposal")
		setupBaseStorage(t, req)

		// setup storage
		err := setupStorageWithWalletAndAccounts(req.Storage)
		require.NoError(t, err)

		// first proposal
		req.Data = basicProposalData()
		_, err = b.HandleRequest(context.Background(), req)
		require.NoError(t, err)

		// second proposal
		req.Data = basicProposalData()
		req.Data["stateRoot"] = "17959acc370274756fa5e9fdd7e7adf17204f49cc8457e49438c42c4883cbfb1"
		_, err = b.HandleRequest(context.Background(), req)
		require.NotNil(t, err)
		require.EqualError(t, err, "failed to sign data: err, slashable proposal: DoubleProposal")
	})

	t.Run("Sign double proposal(different parent root), should error", func(t *testing.T) {
		req := logical.TestRequest(t, logical.CreateOperation, "accounts/sign-proposal")
		setupBaseStorage(t, req)

		// setup storage
		err := setupStorageWithWalletAndAccounts(req.Storage)
		require.NoError(t, err)

		// first proposal
		req.Data = basicProposalData()
		_, err = b.HandleRequest(context.Background(), req)
		require.NoError(t, err)

		// second proposal
		req.Data = basicProposalData()
		req.Data["parentRoot"] = "7402fdc1ce16d449d637c34a172b349a12b2bae8d6d77e401006594d8057c33e"
		_, err = b.HandleRequest(context.Background(), req)
		require.NotNil(t, err)
		require.EqualError(t, err, "failed to sign data: err, slashable proposal: DoubleProposal")
	})

	t.Run("Sign double proposal(different body root), should error", func(t *testing.T) {
		req := logical.TestRequest(t, logical.CreateOperation, "accounts/sign-proposal")
		setupBaseStorage(t, req)

		// setup storage
		err := setupStorageWithWalletAndAccounts(req.Storage)
		require.NoError(t, err)

		// first proposal
		req.Data = basicProposalData()
		_, err = b.HandleRequest(context.Background(), req)
		require.NoError(t, err)

		// second proposal
		req.Data = basicProposalData()
		req.Data["bodyRoot"] = "7b5679277ca45ea74e1deebc9d3e8c0e7d6c570b3cfaf6884be144a81dac9a0d"
		_, err = b.HandleRequest(context.Background(), req)
		require.NotNil(t, err)
		require.EqualError(t, err, "failed to sign data: err, slashable proposal: DoubleProposal")
	})
}
