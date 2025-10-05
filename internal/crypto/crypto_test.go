package crypto

import (
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	data := []byte("this is a secret message")
	passphrase := "supersecret"

	encrypted, err := Encrypt(data, passphrase)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	decrypted, err := Decrypt(encrypted, passphrase)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	if string(decrypted) != string(data) {
		t.Errorf("Decrypted data does not match original data. got=%s, want=%s", string(decrypted), string(data))
	}
}