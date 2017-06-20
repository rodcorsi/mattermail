package mmail

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestUIDCacheFile_GetNextUID(t *testing.T) {
	cache := NewUIDCacheFile(os.TempDir(), "test@example.com", "INBOX")
	defer os.Remove(cache.filename)

	if _, err := cache.GetNextUID(0); err != ErrEmptyUID {
		t.Fatal("Expected ErrEmptyUID err:", err)
	}

	ioutil.WriteFile(cache.filename, []byte{}, 0640)
	if _, err := cache.GetNextUID(0); err == nil {
		t.Fatal("Expected error invalid size")
	}

	if err := cache.SaveNextUID(10, 100); err != nil {
		t.Fatal("Error on save next uid", err.Error())
	}

	if _, err := cache.GetNextUID(9); err != ErrEmptyUID {
		t.Fatal("Expected for uidvalidity 9 ErrEmptyUID err:", err)
	}

	if val, err := cache.GetNextUID(10); err != nil || val != 100 {
		t.Fatalf("Expected value 100 result %v err:%v", val, err)
	}
}
