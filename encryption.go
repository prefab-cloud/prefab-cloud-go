package prefab

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"errors"
	"strings"
)

type Encryption struct{}

var ErrInvalidValueFormat = errors.New("invalid value format")

func (d *Encryption) DecryptValue(secretKeyString string, value string) (string, error) {
	// Decode the hex-encoded secret key
	secretKey, err := hex.DecodeString(strings.ToUpper(secretKeyString))
	if err != nil {
		return "", err
	}

	// Split the value into data, IV, and auth tag parts
	parts := strings.SplitN(strings.ToUpper(value), "--", 3)
	if len(parts) < 3 {
		return "", ErrInvalidValueFormat
	}

	dataStr, ivStr, authTagStr := parts[0], parts[1], parts[2]

	// Decode the hex-encoded parts
	iv, err := hex.DecodeString(ivStr)
	if err != nil {
		return "", err
	}

	dataToProcess, err := hex.DecodeString(dataStr + authTagStr)
	if err != nil {
		return "", err
	}

	// Initialize AES block cipher
	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return "", err
	}

	// Initialize GCM
	gcm, err := cipher.NewGCMWithNonceSize(block, len(iv))
	if err != nil {
		return "", err
	}

	// The actual data to decrypt is `dataToProcess` without the auth tag
	data := dataToProcess[:len(dataToProcess)-gcm.Overhead()]
	authTag := dataToProcess[len(dataToProcess)-gcm.Overhead():]

	// Decrypt the data
	decryptedData, err := gcm.Open(nil, iv, append(data, authTag...), nil)
	if err != nil {
		return "", err
	}

	return string(decryptedData), nil
}
