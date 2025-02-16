package services

import (
	"crypto/x509"
	"encoding/pem"

	"main/crypto"
	"main/types"

	uuid "github.com/satori/go.uuid"
	"go.mongodb.org/mongo-driver/v2/mongo" // Update to v2
)

type InstanceService struct {
	collection *mongo.Collection
}

func NewInstanceService(collection *mongo.Collection) *InstanceService {
	return &InstanceService{collection: collection}
}

func (s *InstanceService) CreateNewInstance(req types.PartialInstanceInfo) types.InstanceInfo {
	return types.InstanceInfo{
		InstanceId:     uuid.NewV4().String(),
		InstanceName:   req.InstanceName,
		InstanceType:   req.InstanceType,
		InstanceStatus: 0,
		Protocol:       req.Protocol,
		Hostname:       req.Hostname,
	}
}

func (s *InstanceService) SetupInstanceEncryption(instance *types.InstanceInfo, token string) (string, error) {
	privateKey, err := crypto.GenerateEncryptionMaterial(instance.InstanceId)
	if err != nil {
		return "", err
	}

	sha256_hash, encrypted_token, err := crypto.EncryptToken(&privateKey.PublicKey, token)
	if err != nil {
		return "", err
	}

	instance.SHA256Hash = sha256_hash.Sum(nil)
	instance.EncryptedToken = encrypted_token

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	return string(pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})), nil
}
