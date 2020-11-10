package backend

import (
	"context"
	"encoding/hex"
	"time"

	vault "github.com/bloxapp/eth2-key-manager"
	"github.com/bloxapp/eth2-key-manager/slashing_protection"
	"github.com/bloxapp/eth2-key-manager/validator_signer"
	"github.com/bloxapp/eth2-key-manager/wallet_hd"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/pkg/errors"
	v1 "github.com/wealdtech/eth2-signer-api/pb/v1"

	"github.com/bloxapp/key-vault/backend/store"
)

// Endpoints patterns
const (
	// SignAttestationPattern is the path pattern for sign attestation endpoint
	SignAttestationPattern = "accounts/sign-attestation"

	// SignProposalPattern is the path pattern for sign proposal endpoint
	SignProposalPattern = "accounts/sign-proposal"

	// SignAggregationPattern is the path pattern for sign aggregation endpoint
	SignAggregationPattern = "accounts/sign-aggregation"
)

func signsPaths(b *backend) []*framework.Path {
	return []*framework.Path{
		&framework.Path{
			Pattern:         SignAttestationPattern,
			HelpSynopsis:    "Sign attestation",
			HelpDescription: `Sign attestation`,
			Fields: map[string]*framework.FieldSchema{
				"public_key": &framework.FieldSchema{
					Type:        framework.TypeString,
					Description: "Public key of the account",
					Default:     "",
				},
				"domain": &framework.FieldSchema{
					Type:        framework.TypeString,
					Description: "Domain",
					Default:     "",
				},
				"slot": &framework.FieldSchema{
					Type:        framework.TypeInt,
					Description: "Data Slot",
					Default:     0,
				},
				"committeeIndex": &framework.FieldSchema{
					Type:        framework.TypeInt,
					Description: "Data CommitteeIndex",
					Default:     0,
				},
				"beaconBlockRoot": &framework.FieldSchema{
					Type:        framework.TypeString,
					Description: "Data BeaconBlockRoot",
					Default:     "",
				},
				"sourceEpoch": &framework.FieldSchema{
					Type:        framework.TypeInt,
					Description: "Data Source Epoch",
					Default:     0,
				},
				"sourceRoot": &framework.FieldSchema{
					Type:        framework.TypeString,
					Description: "Data Source Root",
					Default:     "",
				},
				"targetEpoch": &framework.FieldSchema{
					Type:        framework.TypeInt,
					Description: "Data Target Epoch",
					Default:     0,
				},
				"targetRoot": &framework.FieldSchema{
					Type:        framework.TypeString,
					Description: "Data Target Root",
					Default:     "",
				},
			},
			ExistenceCheck: b.pathExistenceCheck,
			Callbacks: map[logical.Operation]framework.OperationFunc{
				logical.CreateOperation: b.pathSignAttestation,
			},
		},
		&framework.Path{
			Pattern:         SignProposalPattern,
			HelpSynopsis:    "Sign proposal",
			HelpDescription: `Sign proposal`,
			Fields: map[string]*framework.FieldSchema{
				"public_key": &framework.FieldSchema{
					Type:        framework.TypeString,
					Description: "Public key of the account",
					Default:     "",
				},
				"domain": &framework.FieldSchema{
					Type:        framework.TypeString,
					Description: "Domain",
					Default:     "",
				},
				"slot": &framework.FieldSchema{
					Type:        framework.TypeInt,
					Description: "Data Slot",
					Default:     0,
				},
				"proposerIndex": &framework.FieldSchema{
					Type:        framework.TypeInt,
					Description: "Data ProposerIndex",
					Default:     0,
				},
				"parentRoot": &framework.FieldSchema{
					Type:        framework.TypeString,
					Description: "Data ParentRoot",
					Default:     "",
				},
				"stateRoot": &framework.FieldSchema{
					Type:        framework.TypeString,
					Description: "Data StateRoot",
					Default:     "",
				},
				"bodyRoot": &framework.FieldSchema{
					Type:        framework.TypeString,
					Description: "Data BodyRoot",
					Default:     "",
				},
			},
			ExistenceCheck: b.pathExistenceCheck,
			Callbacks: map[logical.Operation]framework.OperationFunc{
				logical.CreateOperation: b.pathSignProposal,
			},
		},
		&framework.Path{
			Pattern:         SignAggregationPattern,
			HelpSynopsis:    "Sign aggregation",
			HelpDescription: `Sign aggregation`,
			Fields: map[string]*framework.FieldSchema{
				"public_key": &framework.FieldSchema{
					Type:        framework.TypeString,
					Description: "Public key of the account",
					Default:     "",
				},
				"domain": &framework.FieldSchema{
					Type:        framework.TypeString,
					Description: "Domain",
					Default:     "",
				},
				"dataToSign": &framework.FieldSchema{
					Type:        framework.TypeString,
					Description: "Data to sign",
					Default:     "",
				},
			},
			ExistenceCheck: b.pathExistenceCheck,
			Callbacks: map[logical.Operation]framework.OperationFunc{
				logical.CreateOperation: b.pathSignAggregation,
			},
		},
	}
}

func (b *backend) pathSignAttestation(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	// Load config
	config, err := b.configured(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get config")
	}

	// bring up KeyVault and wallet
	storage := store.NewHashicorpVaultStore(ctx, req.Storage, config.Network)
	options := vault.KeyVaultOptions{}
	options.SetStorage(storage)

	// Parse request data
	publicKey := data.Get("public_key").(string)
	domain := data.Get("domain").(string)
	slot := data.Get("slot").(int)
	committeeIndex := data.Get("committeeIndex").(int)
	beaconBlockRoot := data.Get("beaconBlockRoot").(string)
	sourceEpoch := data.Get("sourceEpoch").(int)
	sourceRoot := data.Get("sourceRoot").(string)
	targetEpoch := data.Get("targetEpoch").(int)
	targetRoot := data.Get("targetRoot").(string)

	// Open wallet
	kv, err := vault.OpenKeyVault(&options)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open key vault")
	}

	wallet, err := kv.Wallet()
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve wallet")
	}

	account, err := wallet.AccountByPublicKey(publicKey)
	if err != nil {
		if err == wallet_hd.ErrAccountNotFound {
			return b.notFoundResponse()
		}

		return nil, errors.Wrap(err, "failed to retrieve account")
	}

	// try to lock signature lock, if it fails return error
	lock := NewDBLock(account.ID(), req.Storage)
	defer lock.UnLock()
	if err := lock.Lock(); err != nil {
		return nil, err
	}

	// Decode public key
	publicKeyBytes, err := hex.DecodeString(publicKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to HEX decode public key")
	}

	// Decode domain
	domainBytes, err := hex.DecodeString(domain)
	if err != nil {
		return nil, errors.Wrap(err, "failed to HEX decode domain")
	}

	// Decode beacon block root
	beaconBlockRootBytes, err := hex.DecodeString(beaconBlockRoot)
	if err != nil {
		return nil, errors.Wrap(err, "failed to HEX decode beacon block root")
	}

	// Decode source root
	sourceRootBytes, err := hex.DecodeString(sourceRoot)
	if err != nil {
		return nil, errors.Wrap(err, "failed to HEX decode source root")
	}

	// Decode target root
	targetRootBytes, err := hex.DecodeString(targetRoot)
	if err != nil {
		return nil, errors.Wrap(err, "failed to HEX decode target root")
	}

	// Check if the given slot came in time
	if !b.isSlotTime(config.GenesisTime, slot) {
		return nil, errors.Wrap(err, "it's not a slot time")
	}

	protector := slashing_protection.NewNormalProtection(storage)
	var signer validator_signer.ValidatorSigner = validator_signer.NewSimpleSigner(wallet, protector)

	res, err := signer.SignBeaconAttestation(&v1.SignBeaconAttestationRequest{
		Id:     &v1.SignBeaconAttestationRequest_PublicKey{PublicKey: publicKeyBytes},
		Domain: domainBytes,
		Data: &v1.AttestationData{
			Slot:            uint64(slot),
			CommitteeIndex:  uint64(committeeIndex),
			BeaconBlockRoot: beaconBlockRootBytes,
			Source: &v1.Checkpoint{
				Epoch: uint64(sourceEpoch),
				Root:  sourceRootBytes,
			},
			Target: &v1.Checkpoint{
				Epoch: uint64(targetEpoch),
				Root:  targetRootBytes,
			},
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to sign attestation")
	}

	return &logical.Response{
		Data: map[string]interface{}{
			"signature": hex.EncodeToString(res.GetSignature()),
		},
	}, nil
}

func (b *backend) pathSignProposal(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	// Load config
	config, err := b.configured(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get config")
	}

	// bring up KeyVault and wallet
	storage := store.NewHashicorpVaultStore(ctx, req.Storage, config.Network)
	options := vault.KeyVaultOptions{}
	options.SetStorage(storage)

	// Parse request data
	publicKey := data.Get("public_key").(string)
	domain := data.Get("domain").(string)
	slot := data.Get("slot").(int)
	proposerIndex := data.Get("proposerIndex").(int)
	parentRoot := data.Get("parentRoot").(string)
	stateRoot := data.Get("stateRoot").(string)
	bodyRoot := data.Get("bodyRoot").(string)

	// Open wallet
	kv, err := vault.OpenKeyVault(&options)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open key vault")
	}

	wallet, err := kv.Wallet()
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve wallet by name")
	}

	account, err := wallet.AccountByPublicKey(publicKey)
	if err != nil {
		if err == wallet_hd.ErrAccountNotFound {
			return b.notFoundResponse()
		}

		return nil, errors.Wrap(err, "failed to retrieve account")
	}

	// try to lock signature lock, if it fails return error
	lock := NewDBLock(account.ID(), req.Storage)
	defer lock.UnLock()
	if err := lock.Lock(); err != nil {
		return nil, err
	}

	// Decode public key
	publicKeyBytes, err := hex.DecodeString(publicKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to HEX decode public key")
	}

	// Decode domain
	domainBytes, err := hex.DecodeString(domain)
	if err != nil {
		return nil, errors.Wrap(err, "failed to HEX decode domain")
	}

	// Decode parent root
	parentRootBytes, err := hex.DecodeString(parentRoot)
	if err != nil {
		return nil, errors.Wrap(err, "failed to HEX decode parent root")
	}

	// Decode state root
	stateRootBytes, err := hex.DecodeString(stateRoot)
	if err != nil {
		return nil, errors.Wrap(err, "failed to HEX decode state root")
	}

	// Decode body root
	bodyRootBytes, err := hex.DecodeString(bodyRoot)
	if err != nil {
		return nil, errors.Wrap(err, "failed to HEX decode body root")
	}

	// Check if the given slot came in time
	if !b.isSlotTime(config.GenesisTime, slot) {
		return nil, errors.Wrap(err, "it's not a slot time")
	}

	protector := slashing_protection.NewNormalProtection(storage)
	var signer validator_signer.ValidatorSigner = validator_signer.NewSimpleSigner(wallet, protector)

	res, err := signer.SignBeaconProposal(&v1.SignBeaconProposalRequest{
		Id:     &v1.SignBeaconProposalRequest_PublicKey{PublicKey: publicKeyBytes},
		Domain: domainBytes,
		Data: &v1.BeaconBlockHeader{
			Slot:          uint64(slot),
			ProposerIndex: uint64(proposerIndex),
			ParentRoot:    parentRootBytes,
			StateRoot:     stateRootBytes,
			BodyRoot:      bodyRootBytes,
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to sign data")
	}

	return &logical.Response{
		Data: map[string]interface{}{
			"signature": hex.EncodeToString(res.GetSignature()),
		},
	}, nil
}

func (b *backend) pathSignAggregation(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	// Load config
	config, err := b.configured(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get config")
	}

	// bring up KeyVault and wallet
	storage := store.NewHashicorpVaultStore(ctx, req.Storage, config.Network)
	options := vault.KeyVaultOptions{}
	options.SetStorage(storage)

	// Parse request data
	publicKey := data.Get("public_key").(string)
	domain := data.Get("domain").(string)
	dataToSign := data.Get("dataToSign").(string)

	// Open wallet
	kv, err := vault.OpenKeyVault(&options)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open key vault")
	}

	wallet, err := kv.Wallet()
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve wallet by name")
	}

	account, err := wallet.AccountByPublicKey(publicKey)
	if err != nil {
		if err == wallet_hd.ErrAccountNotFound {
			return b.notFoundResponse()
		}

		return nil, errors.Wrap(err, "failed to retrieve account")
	}

	// try to lock signature lock, if it fails return error
	lock := NewDBLock(account.ID(), req.Storage)
	defer lock.UnLock()
	if err := lock.Lock(); err != nil {
		return nil, err
	}

	// Decode public key
	publicKeyBytes, err := hex.DecodeString(publicKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to HEX decode public key")
	}

	// Decode domain
	domainBytes, err := hex.DecodeString(domain)
	if err != nil {
		return nil, errors.Wrap(err, "failed to HEX decode domain")
	}

	// Decode data to sign
	dataToSignBytes, err := hex.DecodeString(dataToSign)
	if err != nil {
		return nil, errors.Wrap(err, "failed to HEX decode data to sign")
	}

	protector := slashing_protection.NewNormalProtection(storage)
	var signer validator_signer.ValidatorSigner = validator_signer.NewSimpleSigner(wallet, protector)

	res, err := signer.Sign(&v1.SignRequest{
		Id:     &v1.SignRequest_PublicKey{PublicKey: publicKeyBytes},
		Domain: domainBytes,
		Data:   dataToSignBytes,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to sign data")
	}

	return &logical.Response{
		Data: map[string]interface{}{
			"signature": hex.EncodeToString(res.GetSignature()),
		},
	}, nil
}

func (b *backend) isSlotTime(genesisTime time.Time, slot int) bool {
	timeSinceGenesisStart := uint64(slot) * 12
	start := genesisTime.Add(time.Duration(timeSinceGenesisStart) * time.Second)
	left := start.Sub(time.Now().UTC())

	// Deviation = seconds per one slot, that's should be enough
	deviation := time.Minute

	return left < deviation || left > -deviation
}
