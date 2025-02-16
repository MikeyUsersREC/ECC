package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"hash"
	"os"
)

func GenerateEncryptionMaterial(instanceId string) (*rsa.PrivateKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	publicKey := &privateKey.PublicKey

	pubKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return nil, err
	}
	pubKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: pubKeyBytes,
	})
	err = os.WriteFile(fmt.Sprintf("keys/%s.pem", instanceId), pubKeyPEM, 0600)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

func EncryptToken(rsaPublicKey *rsa.PublicKey, token string) (hash.Hash, []byte, error) {
	sha256_hash := sha256.New()
	encrypted, err := rsa.EncryptOAEP(sha256_hash, rand.Reader, rsaPublicKey, []byte(token), nil)
	if err != nil {
		return nil, nil, err
	}
	return sha256_hash, encrypted, err
}
