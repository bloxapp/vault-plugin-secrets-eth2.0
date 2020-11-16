package backend

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/pkg/errors"
)

var (
	// ErrLocked represents the locked error.
	ErrLocked = errors.New("locked")
)

// DBLock implements DB locking mechanism.
type DBLock struct {
	id      uuid.UUID
	storage logical.Storage
}

// NewDBLock is the constructor of DBLock.
func NewDBLock(id uuid.UUID, storage logical.Storage) *DBLock {
	return &DBLock{
		id:      id,
		storage: storage,
	}
}

// Lock locks the DB.
func (lock *DBLock) Lock() error {
	// if locked return error
	locked, err := lock.IsLocked()
	if err != nil {
		return errors.Wrap(err, "failed to check if it is locked")
	}
	if locked {
		return ErrLocked
	}

	// add lock to db
	return lock.storage.Put(context.Background(), &logical.StorageEntry{
		Key:      lock.key(),
		Value:    []byte("1"),
		SealWrap: false,
	})
}

// UnLock unlocks the DB.
func (lock *DBLock) UnLock() error {
	locked, err := lock.IsLocked()
	if err != nil {
		return errors.Wrap(err, "failed to check if it is locked")
	}
	if !locked {
		return nil
	}

	// if not, unlock
	return lock.storage.Delete(context.Background(), lock.key())
}

// IsLocked returns true if the DB is locked.
func (lock *DBLock) IsLocked() (bool, error) {
	entry, err := lock.storage.Get(context.Background(), lock.key())
	if err != nil {
		return false, err
	}

	return entry != nil, nil
}

func (lock *DBLock) key() string {
	return fmt.Sprintf("lock/%s", lock.id.String())
}
