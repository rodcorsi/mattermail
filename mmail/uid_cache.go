package mmail

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// ErrEmptyUID error when is empty UID stored or invalid UIDVALIDITY
var ErrEmptyUID = errors.New("Empty UID stored or invalid UIDVALIDITY")

// UIDCache is a cache of uid for account + mailbox
type UIDCache interface {
	// GetNextUID returns the next uid for the uidvalidity, if empty or is an invalid uidvalidity returns ErrEmptyUID
	GetNextUID(uidvalidity uint32) (uint32, error)

	// SaveNextUID stores the uid and uidvalidity
	SaveNextUID(uidvalidity, uid uint32) error
}

// UIDCacheFile implements UIDCache using filesystem to store
type UIDCacheFile struct {
	filename    string
	uidvalidity uint32
	lock        sync.RWMutex
}

// NewUIDCacheFile return a new UIDCacheFile
func NewUIDCacheFile(directory, account, mailbox string) *UIDCacheFile {
	filename := filepath.Join(directory, strings.ToLower(account+"_"+mailbox+".dat"))

	return &UIDCacheFile{
		filename: filename,
	}
}

// GetNextUID returns the next uid for the uidvalidity, if empty or is an invalid uidvalidity returns ErrEmptyUID
func (u *UIDCacheFile) GetNextUID(uidvalidity uint32) (uint32, error) {
	u.lock.RLock()
	defer u.lock.RUnlock()

	if _, err := os.Stat(u.filename); err != nil {
		if os.IsNotExist(err) {
			return 0, ErrEmptyUID
		}
	}

	data, err := ioutil.ReadFile(u.filename)
	if err != nil {
		return 0, fmt.Errorf("Error on read file '%v' err:%v", u.filename, err.Error())
	}

	if len(data) != 8 {
		return 0, fmt.Errorf("Error on read file '%v' invalid size", u.filename)
	}

	uidval := binary.LittleEndian.Uint32(data[:4])

	if uidval != uidvalidity {
		return 0, ErrEmptyUID
	}

	return binary.LittleEndian.Uint32(data[4:]), nil
}

// SaveNextUID stores the uid and uidvalidity
func (u *UIDCacheFile) SaveNextUID(uidvalidity, uid uint32) error {
	u.lock.Lock()
	defer u.lock.Unlock()

	bval := make([]byte, 4)
	binary.LittleEndian.PutUint32(bval, uidvalidity)

	buid := make([]byte, 4)
	binary.LittleEndian.PutUint32(buid, uid)

	b := append(bval, buid...)

	return ioutil.WriteFile(u.filename, b, 0640)
}
