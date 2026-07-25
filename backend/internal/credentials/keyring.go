package credentials

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
)

var (
	ErrNoKeys         = errors.New("credential encryption keys are not configured")
	ErrUnknownVersion = errors.New("unknown credential encryption key version")
)

type Keyring struct {
	current string
	keys    map[string][]byte
}

func Parse(value string) (*Keyring, error) {
	ring := &Keyring{keys: map[string][]byte{}}
	for _, item := range strings.Split(value, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		parts := strings.SplitN(item, ":", 2)
		if len(parts) != 2 || parts[0] == "" {
			return nil, errors.New("encryption keys must use version:base64 format")
		}
		key, err := base64.StdEncoding.DecodeString(parts[1])
		if err != nil || len(key) != 32 {
			return nil, fmt.Errorf("encryption key %q must be 32 bytes encoded as base64", parts[0])
		}
		if ring.current == "" {
			ring.current = parts[0]
		}
		ring.keys[parts[0]] = key
	}
	if ring.current == "" {
		return nil, ErrNoKeys
	}
	return ring, nil
}

func (k *Keyring) Encrypt(plaintext []byte, aad string) (ciphertext, nonce []byte, version string, err error) {
	block, err := aes.NewCipher(k.keys[k.current])
	if err != nil {
		return nil, nil, "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, "", err
	}
	nonce = make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, "", err
	}
	return gcm.Seal(nil, nonce, plaintext, []byte(aad)), nonce, k.current, nil
}

func (k *Keyring) Decrypt(ciphertext, nonce []byte, version, aad string) ([]byte, error) {
	key, ok := k.keys[version]
	if !ok {
		return nil, ErrUnknownVersion
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return gcm.Open(nil, nonce, ciphertext, []byte(aad))
}
