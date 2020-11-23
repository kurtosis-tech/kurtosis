/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package session_cache

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/gob"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"sync"
)

/*
Encryption key used for stuff stored on disk - uses 32 bytes = AES-256
IMPORTANT: DO NOT EVER LOG THIS!!!
 */
var blockKey_DO_NOT_EVER_LOG_ME = []byte{
	0x4a, 0xe5, 0x95, 0xfa, 0x2e, 0x6e, 0x3e, 0xb5, 0x61, 0xca,
	0xdf, 0xde, 0xf2, 0x40, 0xd8, 0x66, 0xd2, 0xad, 0x49, 0x80,
	0x6f, 0xda, 0x13, 0xb5, 0xba, 0xd4, 0x1d, 0x0b, 0x9c, 0xcd,
	0xd7, 0xf0,
}

/*
Session cache that stores encrypted information about a Kurtosis session on disk
 */
type EncryptedSessionCache struct {
	storageFilepath string

	storageFilepathMode os.FileMode

	lock            sync.Mutex
}

func NewEncryptedSessionCache(cacheFilepath string, cacheFilepathMode os.FileMode) *EncryptedSessionCache {
	logrus.Debugf("Initializing session cache storing data to: %v", cacheFilepath)
	var lock sync.Mutex
	return &EncryptedSessionCache{
		storageFilepath:     cacheFilepath,
		storageFilepathMode: cacheFilepathMode,
		lock:                lock,
	}
}

/*
Writes an encrypted session to the file that was given at session cache creation time
 */
func (cache *EncryptedSessionCache) SaveSession(session Session) error {
	cache.lock.Lock()
	defer cache.lock.Unlock()

	var gobbedSession bytes.Buffer
	encoder := gob.NewEncoder(&gobbedSession)
	if err := encoder.Encode(session); err != nil {
		return stacktrace.Propagate(err, "An error occurred serializing the session")
	}

	encryptedGobbedSession, err := encrypt(gobbedSession.Bytes())
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred encrypting the session before storing on disk")
	}

	if err := ioutil.WriteFile(cache.storageFilepath, encryptedGobbedSession, cache.storageFilepathMode); err != nil {
		return stacktrace.Propagate(err, "An error occurred writing the encrypted session bytes to %v", cache.storageFilepath)
	}
	return nil
}

/*
Decrypts a session from the session cache filepath given at session cache creation time
*/
func (cache *EncryptedSessionCache) LoadSession() (tokenResponse *Session, err error) {
	encryptedGobbedSession, err := ioutil.ReadFile(cache.storageFilepath)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred reading the encrypted session from %v", cache.storageFilepath)
	}

	gobbedSession, err := decrypt(encryptedGobbedSession)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred decrypting the session")
	}


	gobbedSessionReader := bytes.NewReader(gobbedSession)
	decoder := gob.NewDecoder(gobbedSessionReader)
	var session Session
	if err := decoder.Decode(&session); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred deserializing the session")
	}

	return &session, nil
}

// ================================= HELPER FUNCTIONS =========================================
// This is a separate helper function - rather than being stored on the EncryptedSessionCache object itself - because
//  I _think_ it's bad to reuse the object
func getGaloisCounterMode() (cipher.AEAD, error) {
	cipherBlock, err := aes.NewCipher(blockKey_DO_NOT_EVER_LOG_ME)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Couldn't get cipher")
	}

	gcm, err := cipher.NewGCM(cipherBlock)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Couldn't get Galois Counter Mode for cipher")
	}

	return gcm, nil
}

// From https://tutorialedge.net/golang/go-encrypt-decrypt-aes-tutorial/
func encrypt(plaintext []byte) ([]byte, error) {
	gcm, err := getGaloisCounterMode()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get Galois Counter Mode for encrypting session")
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred populating nonce with random data while encrypting session")
	}

	encrypted := gcm.Seal(nonce, nonce, plaintext, nil)
	return encrypted, nil
}

// From https://tutorialedge.net/golang/go-encrypt-decrypt-aes-tutorial/
func decrypt(ciphertext []byte) ([]byte, error) {
	gcm, err := getGaloisCounterMode()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get Galois Counter Mode for decrypting session")
	}
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, stacktrace.NewError("Length of ciphertext is less than GCM nonce size, %v", nonceSize)
	}
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred decrypting the encrypted session")
	}
	return plaintext, nil
}
