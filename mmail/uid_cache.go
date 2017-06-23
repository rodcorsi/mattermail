package mmail

import (
	"errors"
	"sync"
)

// ErrEmptyUID error when is empty UID stored or invalid UIDVALIDITY
var ErrEmptyUID = errors.New("Empty UID stored or invalid UIDVALIDITY")

// ErrUIDValidityZero error when UIDVALIDITY is zero
var ErrUIDValidityZero = errors.New("uidvalidity need to be greater than zero")

// UIDCache is a cache of uid for account + mailbox
type UIDCache interface {
	// GetNextUID returns the next uid for the uidvalidity, if empty or is an invalid uidvalidity returns ErrEmptyUID
	GetNextUID(uidvalidity uint32) (uint32, error)

	// SaveNextUID stores the uid and uidvalidity
	SaveNextUID(uidvalidity, uid uint32) error
}

type uidCacheMem struct {
	uidvalidity uint32
	next        uint32
	lock        sync.RWMutex
}

// GetNextUID returns the next uid for the uidvalidity, if empty or is an invalid uidvalidity returns ErrEmptyUID
func (u *uidCacheMem) GetNextUID(uidvalidity uint32) (uint32, error) {
	u.lock.RLock()
	defer u.lock.RUnlock()

	if uidvalidity == 0 {
		return 0, ErrUIDValidityZero
	}

	if u.uidvalidity != uidvalidity {
		return 0, ErrEmptyUID
	}

	return u.next, nil
}

// SaveNextUID stores the uid and uidvalidity
func (u *uidCacheMem) SaveNextUID(uidvalidity, uid uint32) error {
	u.lock.Lock()
	defer u.lock.Unlock()

	if uidvalidity == 0 {
		return ErrUIDValidityZero
	}
	u.uidvalidity = uidvalidity
	u.next = uid

	return nil
}
