package store

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"

	"github.com/bloxapp/eth2-key-manager/core"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/pkg/errors"
	e2types "github.com/wealdtech/go-eth2-types/v2"
)

// Paths
const (
	WalletHighestAttestationPath = "highestAttestations/"
	WalletProposalsBase          = "proposals/%s/"            // account/proposal
	WalletProposalsPath          = WalletProposalsBase + "%d" // account/proposal
)

// SaveHighestAttestation saves highest attestation
func (store *HashicorpVaultStore) SaveHighestAttestation(key e2types.PublicKey, req *core.BeaconAttestation) error {
	path := fmt.Sprintf(WalletHighestAttestationPath+"%s", store.identifierFromKey(key))
	data, err := json.Marshal(req)
	if err != nil {
		return errors.Wrap(err, "failed to marshal attestation request")
	}

	return store.storage.Put(store.ctx, &logical.StorageEntry{
		Key:      path,
		Value:    data,
		SealWrap: false,
	})
}

// RetrieveHighestAttestation retrieves highest attestation
func (store *HashicorpVaultStore) RetrieveHighestAttestation(key e2types.PublicKey) *core.BeaconAttestation {
	path := fmt.Sprintf(WalletHighestAttestationPath+"%s", store.identifierFromKey(key))
	entry, err := store.storage.Get(store.ctx, path)
	if err != nil {
		return nil
	}

	// Return nothing if there is no record
	if entry == nil {
		return nil
	}

	var ret core.BeaconAttestation
	if err := json.Unmarshal(entry.Value, &ret); err != nil {
		return nil
	}

	return &ret
}

// SaveProposal implements Storage interface.
func (store *HashicorpVaultStore) SaveProposal(key e2types.PublicKey, req *core.BeaconBlockHeader) error {
	path := fmt.Sprintf(WalletProposalsPath, store.identifierFromKey(key), req.Slot)
	data, err := json.Marshal(req)
	if err != nil {
		return errors.Wrap(err, "failed to marshal proposal request")
	}

	return store.storage.Put(store.ctx, &logical.StorageEntry{
		Key:      path,
		Value:    data,
		SealWrap: false,
	})
}

// RetrieveProposal implements Storage interface.
func (store *HashicorpVaultStore) RetrieveProposal(key e2types.PublicKey, slot uint64) (*core.BeaconBlockHeader, error) {
	path := fmt.Sprintf(WalletProposalsPath, store.identifierFromKey(key), slot)
	entry, err := store.storage.Get(store.ctx, path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get record with path '%s'", path)
	}

	// Return nothing if there is no record
	if entry == nil {
		return nil, nil
	}

	var ret core.BeaconBlockHeader
	if err = json.Unmarshal(entry.Value, &ret); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal beacon block header object")
	}

	return &ret, nil
}

// ListAllProposals returns all proposal data from the DB
func (store *HashicorpVaultStore) ListAllProposals(key e2types.PublicKey) ([]*core.BeaconBlockHeader, error) {
	path := fmt.Sprintf(WalletProposalsBase, store.identifierFromKey(key))
	entries, err := store.storage.List(store.ctx, path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to list records from storage with path '%s'", path)
	}

	// Return nothing if there is no record
	if len(entries) == 0 {
		return nil, nil
	}

	proposals := make([]*core.BeaconBlockHeader, len(entries))
	errs := make([]error, len(entries))

	var wg sync.WaitGroup
	for i, entry := range entries {
		wg.Add(1)
		go func(i int, entry string) {
			defer wg.Done()

			if entry == "latest" {
				return
			}

			epoch, err := strconv.Atoi(entry)
			if err != nil {
				errs[i] = errors.Wrapf(err, "invalid epoch number %s", entry)
				return
			}

			ret, err := store.RetrieveProposal(key, uint64(epoch))
			if err != nil {
				errs[i] = errors.Wrapf(err, "failed to retrieve beacon proposal for epoch %d", epoch)
				return
			}

			proposals[i] = ret
		}(i, entry)
	}
	wg.Wait()

	for _, err := range errs {
		if err != nil {
			return nil, err
		}
	}

	clearedProposals := make([]*core.BeaconBlockHeader, 0)
	for _, proposal := range proposals {
		if proposal != nil {
			clearedProposals = append(clearedProposals, proposal)
		}
	}

	return clearedProposals, nil
}

func (store *HashicorpVaultStore) identifierFromKey(key e2types.PublicKey) string {
	return hex.EncodeToString(key.Marshal())
}
