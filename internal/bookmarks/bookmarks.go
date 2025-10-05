package bookmarks

import (
	"dosc/internal/crypto"
	"encoding/json"
	"io/ioutil"
	"os"
)

const bookmarksFile = "bookmarks.json.enc"

// Bookmark represents a single connection bookmark.
type Bookmark struct {
	Name          string `json:"name"`
	Endpoint      string `json:"endpoint"`
	BucketName    string `json:"bucket_name"`
	ClientKey     string `json:"client_key"`
	SecretKey     string `json:"secret_key"`
	Region        string `json:"region"`
	DefaultPrefix string `json:"default_prefix"`
}

// SaveBookmarks encrypts and saves a list of bookmarks to a file.
func SaveBookmarks(bms []Bookmark, passphrase string) error {
	data, err := json.Marshal(bms)
	if err != nil {
		return err
	}

	encryptedData, err := crypto.Encrypt(data, passphrase)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(bookmarksFile, encryptedData, 0600)
}

// LoadBookmarks loads, decrypts, and parses bookmarks from a file.
func LoadBookmarks(passphrase string) ([]Bookmark, error) {
	if _, err := os.Stat(bookmarksFile); os.IsNotExist(err) {
		// If the file doesn't exist, return an empty list.
		return []Bookmark{}, nil
	}

	encryptedData, err := ioutil.ReadFile(bookmarksFile)
	if err != nil {
		return nil, err
	}

	data, err := crypto.Decrypt(encryptedData, passphrase)
	if err != nil {
		return nil, err
	}

	var bms []Bookmark
	if err := json.Unmarshal(data, &bms); err != nil {
		return nil, err
	}

	return bms, nil
}