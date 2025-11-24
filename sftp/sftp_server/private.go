package sftpserver

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"log"
	"os"

	"golang.org/x/crypto/ssh"
)

// generateHostKey generates or loads an SSH host key for the SFTP server
// It first tries to load an existing key from "id_rsa", and generates a new one if it doesn't exist.
// The generated key is saved to "id_rsa" for future use.
func generateHostKey() (ssh.Signer, error) {
	// Load or generate SSH host key
	privateBytes, err := os.ReadFile("id_rsa")
	if err != nil {
		log.Printf("Private key file not found, generating new key pair...\n")

		// If key doesn't exist, generate a new key pair
		privateKey, err := generatePrivateKey()
		if err != nil {
			log.Fatalf("Failed to generate private key: %v", err)
		}

		log.Printf("Generated new private key successfully.\n")

		privateBytes = encodePrivateKeyToPEM(privateKey)
		if err = os.WriteFile("id_rsa", privateBytes, 0600); err != nil {
			log.Fatalf("Failed to save private key: %v", err)
		}

		log.Printf("Saved private key to id_rsa successfully.\n")
	}

	return ssh.ParsePrivateKey(privateBytes)
}

// generatePrivateKey generates an RSA private key
// This function generates a 2048-bit RSA key pair using cryptographically secure random number generation.
func generatePrivateKey() (*rsa.PrivateKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}

// encodePrivateKeyToPEM encodes an RSA private key to PEM format
// This function converts the RSA private key into PKCS#1 format and wraps it in a PEM block.
func encodePrivateKeyToPEM(privateKey *rsa.PrivateKey) []byte {
	privDER := x509.MarshalPKCS1PrivateKey(privateKey)
	privBlock := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privDER,
	}
	privatePEM := pem.EncodeToMemory(&privBlock)
	return privatePEM
}
