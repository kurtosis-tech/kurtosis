/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package session_cache

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestNonexistentCache(t *testing.T) {
	cache := NewEncryptedSessionCache("/this/doesnt/exist", os.ModePerm)
	if _, err := cache.LoadSession(); err == nil {
		t.Fatal("Expected error when loading cache from nonexistent filepath, but didn't get one")
	}
}

func TestUnparseableCache(t *testing.T) {
	fp, err := ioutil.TempFile("", "testfile")
	if err != nil {
		t.Fatal("An error occurred creating the tempfile")
	}
	defer fp.Close()
	fp.WriteString("this file content will fail decryption")

	cache := NewEncryptedSessionCache(fp.Name(), os.ModePerm)
	if _, err := cache.LoadSession(); err == nil {
		t.Fatal("Expected error when loading cache from improperly-encrypted file, but didn't get one")
	}
}