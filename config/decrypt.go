package config

import (
	"encoding/base64"
	"fmt"
)

// Decrypt decrypt string.
func Decrypt(encryptedText string) (string, error) {
	// Base64 decode
	data, err := base64.StdEncoding.DecodeString(encryptedText)
	if err != nil {
		return "", fmt.Errorf("base64 decode failed: %w", err)
	}

	// For now, just return the decoded data as string
	// In a real implementation, you would decrypt the data here
	return string(data), nil
}

// Encrypt encrypt string.
func Encrypt(plaintext string) (string, error) {
	// For now, just base64 encode the plaintext
	// In a real implementation, you would encrypt the data here
	return base64.StdEncoding.EncodeToString([]byte(plaintext)), nil
}
