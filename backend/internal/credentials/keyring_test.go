package credentials

import (
	"encoding/base64"
	"testing"
)

func testConfig() string {
	return "v2:" + base64.StdEncoding.EncodeToString([]byte("01234567890123456789012345678901"))
}

func TestEncryptDecryptAndAAD(t *testing.T) {
	ring, err := Parse(testConfig())
	if err != nil {
		t.Fatal(err)
	}
	ciphertext, nonce, version, err := ring.Encrypt([]byte("admin-secret"), "connection-1")
	if err != nil {
		t.Fatal(err)
	}
	plaintext, err := ring.Decrypt(ciphertext, nonce, version, "connection-1")
	if err != nil || string(plaintext) != "admin-secret" {
		t.Fatalf("decrypt = %q, %v", plaintext, err)
	}
	if _, err := ring.Decrypt(ciphertext, nonce, version, "connection-2"); err == nil {
		t.Fatal("expected AAD tamper detection")
	}
	ciphertext[0] ^= 1
	if _, err := ring.Decrypt(ciphertext, nonce, version, "connection-1"); err == nil {
		t.Fatal("expected ciphertext tamper detection")
	}
}

func TestParseRejectsInvalidKeys(t *testing.T) {
	if _, err := Parse(""); err == nil {
		t.Fatal("expected missing key error")
	}
	if _, err := Parse("v1:c2hvcnQ="); err == nil {
		t.Fatal("expected invalid key length")
	}
}
