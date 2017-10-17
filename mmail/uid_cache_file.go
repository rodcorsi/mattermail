package mmail

import (
	"encoding/binary"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/pkg/errors"
)

// UIDCacheFile implements UIDCache using filesystem to store
type UIDCacheFile struct {
	filename string
	lock     sync.RWMutex
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

	if uidvalidity == 0 {
		return 0, ErrUIDValidityZero
	}

	if _, err := os.Stat(u.filename); err != nil {
		if os.IsNotExist(err) {
			return 0, ErrEmptyUID
		}
	}

	data, err := ioutil.ReadFile(u.filename)
	if err != nil {
		return 0, errors.Wrapf(err, "Error on read file '%v'", u.filename)
	}

	if len(data) != 8 {
		return 0, errors.Errorf("Error on read file '%v' invalid size", u.filename)
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

	if uidvalidity == 0 {
		return ErrUIDValidityZero
	}

	bval := make([]byte, 4)
	binary.LittleEndian.PutUint32(bval, uidvalidity)

	buid := make([]byte, 4)
	binary.LittleEndian.PutUint32(buid, uid)

	b := append(bval, buid...)

	return ioutil.WriteFile(u.filename, b, 0640)
}
