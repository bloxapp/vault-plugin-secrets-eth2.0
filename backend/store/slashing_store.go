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
	WalletAttestationsBase      = "attestations/%s/"
	WalletAttestationPath       = WalletAttestationsBase + "%d"     // account/attestation
	WalletLatestAttestationPath = WalletAttestationsBase + "latest" // account/latest
	WalletProposalsBase         = "proposals/%s/"                   // account/proposal
	WalletProposalsPath         = WalletProposalsBase + "%d"        // account/proposal
)

// SaveAttestation implements Storage interface.
func (store *HashicorpVaultStore) SaveAttestation(key e2types.PublicKey, req *core.BeaconAttestation) error {
	path := fmt.Sprintf(WalletAttestationPath, store.identifierFromKey(key), req.Target.Epoch)
	data, err := json.Marshal(req)
	if err != nil {
		return errors.Wrap(err, "failed to marshal attestation request")
	}

	entry := &logical.StorageEntry{
		Key:      path,
		Value:    data,
		SealWrap: false,
	}
	return store.storage.Put(store.ctx, entry)
}

// RetrieveAttestation implements Storage interface.
func (store *HashicorpVaultStore) RetrieveAttestation(key e2types.PublicKey, epoch uint64) (*core.BeaconAttestation, error) {
	path := fmt.Sprintf(WalletAttestationPath, store.identifierFromKey(key), epoch)
	entry, err := store.storage.Get(store.ctx, path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get record from storage with path '%s'", path)
	}

	// Return nothing if there is no record
	if entry == nil {
		return nil, nil
	}

	var ret *core.BeaconAttestation
	if err := json.Unmarshal(entry.Value, &ret); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal beacon attestation object")
	}

	return ret, nil
}

// ListAttestations both epochStart and epochEnd reflect saved attestations by their target epoch
func (store *HashicorpVaultStore) ListAttestations(key e2types.PublicKey, epochStart uint64, epochEnd uint64) ([]*core.BeaconAttestation, error) {
	length := epochEnd - epochStart + 1
	ret := make([]*core.BeaconAttestation, length)
	errs := make([]error, length)

	var wg sync.WaitGroup
	var i int
	for epoch := epochStart; epoch <= epochEnd; epoch++ {
		wg.Add(1)
		go func(i int, epoch uint64) {
			defer wg.Done()

			att, err := store.RetrieveAttestation(key, epoch)
			if err != nil {
				errs[i] = errors.Wrapf(err, "failed to retrieve attestation with epoch %d", epoch)
				return
			}

			ret[i] = att
		}(i, epoch)
		i++
	}
	wg.Wait()

	for _, err := range errs {
		if err != nil {
			return nil, err
		}
	}

	preparedRes := make([]*core.BeaconAttestation, 0)
	for _, r := range ret {
		if r != nil {
			preparedRes = append(preparedRes, r)
		}
	}

	return preparedRes, nil
}

// ListAllAttestations returns all attestation data from the DB
func (store *HashicorpVaultStore) ListAllAttestations(key e2types.PublicKey) ([]*core.BeaconAttestation, error) {
	path := fmt.Sprintf(WalletAttestationsBase, store.identifierFromKey(key))
	entries, err := store.storage.List(store.ctx, path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to list records from storage with path '%s'", path)
	}

	// Return nothing if there is no record
	if len(entries) == 0 {
		return nil, nil
	}

	attestations := make([]*core.BeaconAttestation, len(entries))
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

			ret, err := store.RetrieveAttestation(key, uint64(epoch))
			if err != nil {
				errs[i] = errors.Wrapf(err, "failed to retrieve beacon attestation for epoch %d", epoch)
				return
			}

			attestations[i] = ret
		}(i, entry)
	}
	wg.Wait()

	for _, err := range errs {
		if err != nil {
			return nil, err
		}
	}

	clearedAttestations := make([]*core.BeaconAttestation, 0)
	for _, attestation := range attestations {
		if attestation != nil {
			clearedAttestations = append(clearedAttestations, attestation)
		}
	}

	return clearedAttestations, nil
}

// SaveProposal implements Storage interface.
func (store *HashicorpVaultStore) SaveProposal(key e2types.PublicKey, req *core.BeaconBlockHeader) error {
	path := fmt.Sprintf(WalletProposalsPath, store.identifierFromKey(key), req.Slot)
	data, err := json.Marshal(req)
	if err != nil {
		return errors.Wrap(err, "failed to marshal proposal request")
	}

	entry := &logical.StorageEntry{
		Key:      path,
		Value:    data,
		SealWrap: false,
	}
	return store.storage.Put(store.ctx, entry)
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

	var ret *core.BeaconBlockHeader
	if err = json.Unmarshal(entry.Value, &ret); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal beacon block header object")
	}

	return ret, nil
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

// SaveLatestAttestation implements Storage interface.
func (store *HashicorpVaultStore) SaveLatestAttestation(key e2types.PublicKey, req *core.BeaconAttestation) error {
	path := fmt.Sprintf(WalletLatestAttestationPath, store.identifierFromKey(key))
	data, err := json.Marshal(req)
	if err != nil {
		return errors.Wrap(err, "failed to marshal beacon attestation object")
	}

	entry := &logical.StorageEntry{
		Key:      path,
		Value:    data,
		SealWrap: false,
	}
	return store.storage.Put(store.ctx, entry)
}

// RetrieveLatestAttestation implements Storage imterface.
func (store *HashicorpVaultStore) RetrieveLatestAttestation(key e2types.PublicKey) (*core.BeaconAttestation, error) {
	path := fmt.Sprintf(WalletLatestAttestationPath, store.identifierFromKey(key))
	entry, err := store.storage.Get(store.ctx, path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get record with path '%s'", path)
	}

	// Return nothing if there is no record
	if entry == nil {
		return nil, nil
	}

	var ret *core.BeaconAttestation
	if err := json.Unmarshal(entry.Value, &ret); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal beacon attestation object")
	}

	return ret, nil
}

func (store *HashicorpVaultStore) identifierFromKey(key e2types.PublicKey) string {
	return hex.EncodeToString(key.Marshal())
}
