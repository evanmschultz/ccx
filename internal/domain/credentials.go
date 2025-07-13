package domain

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
)

// Credentials represents encrypted account credentials
type Credentials struct {
	accountID     AccountID
	encryptedData []byte
	encryptionKey []byte
}

// credentialsJSON is used for serialization
type credentialsJSON struct {
	AccountID     string `json:"accountId"`
	EncryptedData string `json:"encryptedData"`
}

// deriveKey derives an encryption key from the account ID
// In production, this should use a proper KDF and possibly a master key
func deriveKey(accountID AccountID) []byte {
	hash := sha256.Sum256([]byte("ccx-encryption-" + string(accountID)))
	return hash[:]
}

// NewCredentials creates new encrypted credentials
func NewCredentials(accountID AccountID, data []byte) (*Credentials, error) {
	if accountID == "" {
		return nil, errors.New("account ID cannot be empty")
	}

	if len(data) == 0 {
		return nil, errors.New("credentials data cannot be empty")
	}

	key := deriveKey(accountID)
	encryptedData, err := encrypt(data, key)
	if err != nil {
		return nil, err
	}

	return &Credentials{
		accountID:     accountID,
		encryptedData: encryptedData,
		encryptionKey: key,
	}, nil
}

// AccountID returns the account ID associated with these credentials
func (c *Credentials) AccountID() AccountID {
	return c.accountID
}

// EncryptedData returns the encrypted credential data
func (c *Credentials) EncryptedData() []byte {
	// Return a copy to prevent external modification
	data := make([]byte, len(c.encryptedData))
	copy(data, c.encryptedData)
	return data
}

// Decrypt decrypts and returns the credential data
func (c *Credentials) Decrypt() ([]byte, error) {
	return decrypt(c.encryptedData, c.encryptionKey)
}

// UpdateData updates the encrypted credentials with new data
func (c *Credentials) UpdateData(newData []byte) error {
	if len(newData) == 0 {
		return errors.New("credentials data cannot be empty")
	}

	encryptedData, err := encrypt(newData, c.encryptionKey)
	if err != nil {
		return err
	}

	c.encryptedData = encryptedData
	return nil
}

// Clone creates a deep copy of the credentials
func (c *Credentials) Clone() *Credentials {
	encryptedData := make([]byte, len(c.encryptedData))
	copy(encryptedData, c.encryptedData)

	key := make([]byte, len(c.encryptionKey))
	copy(key, c.encryptionKey)

	return &Credentials{
		accountID:     c.accountID,
		encryptedData: encryptedData,
		encryptionKey: key,
	}
}

// Serialize converts credentials to JSON for storage
func (c *Credentials) Serialize() ([]byte, error) {
	data := credentialsJSON{
		AccountID:     string(c.accountID),
		EncryptedData: base64.StdEncoding.EncodeToString(c.encryptedData),
	}
	return json.Marshal(data)
}

// DeserializeCredentials recreates credentials from JSON
func DeserializeCredentials(data []byte) (*Credentials, error) {
	if len(data) == 0 {
		return nil, errors.New("empty serialization data")
	}

	var jsonData credentialsJSON
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return nil, err
	}

	if jsonData.AccountID == "" {
		return nil, errors.New("missing account ID in serialized data")
	}

	if jsonData.EncryptedData == "" {
		return nil, errors.New("missing encrypted data in serialized data")
	}

	encryptedData, err := base64.StdEncoding.DecodeString(jsonData.EncryptedData)
	if err != nil {
		return nil, err
	}

	accountID := AccountID(jsonData.AccountID)
	return &Credentials{
		accountID:     accountID,
		encryptedData: encryptedData,
		encryptionKey: deriveKey(accountID),
	}, nil
}

// encrypt encrypts data using AES-GCM
func encrypt(plaintext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// decrypt decrypts data using AES-GCM
func decrypt(ciphertext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
