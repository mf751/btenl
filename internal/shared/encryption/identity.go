package encryption

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
)

const (
	privateKeyFileMode = 0o600
	publicKeyFileMode  = 0o644
)

// Identity is a persistent ed25519 keypair, it represents the identity of the machine.
// generated only once and later signs certificates that will be used when connecting
// to other machines and a fingerprint of the public key is used as the machine identity
type Identity struct {
	priv ed25519.PrivateKey
	pub  ed25519.PublicKey
}

// NOTE: checks if there is Identity and if not generates one
func EnsureIdentity() (*Identity, error) {
	privPath, _, err := IdentityPaths()
	if err != nil {
		return nil, err
	}

	return LoadOrGenerate(privPath)
}

func LoadOrGenerate(privPath string) (*Identity, error) {
	id, err := loadIdentity(privPath)
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	id, err = generateIdentity()
	if err != nil {
		return nil, err
	}
	if err = id.save(privPath); err != nil {
		return nil, err
	}

	return id, nil
}

// INFO: loads private key from file and checks if valid
func loadIdentity(privPath string) (*Identity, error) {
	data, err := os.ReadFile(privPath)
	if err != nil {
		return nil, err
	}

	// NOTE: decode private key into pem as was saved
	block, _ := pem.Decode(data)
	if block == nil || block.Type != "PRIVATE KEY" {
		return nil, fmt.Errorf("encryption: %s: invalid private key pem %w", privPath, err)
	}

	// NOTE: invalid if not in the format we saved it in
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("encryption: %s: %w", privPath, err)
	}

	priv, ok := key.(ed25519.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("encryption: %s: not an ed25519 key", privPath)
	}

	// NOTE: get public key from the private
	return &Identity{priv: priv, pub: priv.Public().(ed25519.PublicKey)}, nil
}

// INFO: generates new ed25519 keypair
func generateIdentity() (*Identity, error) {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("encryption: generate identity: %w", err)
	}

	// NOTE: get public key from the private
	return &Identity{priv: priv, pub: priv.Public().(ed25519.PublicKey)}, nil
}

func (id *Identity) save(privPath string) error {
	// NOTE: format private key
	privDER, err := x509.MarshalPKCS8PrivateKey(id.priv)
	if err != nil {
		return fmt.Errorf("encryption: marshal private key: %w", err)
	}

	// NOTE: format public key
	pubDER, err := x509.MarshalPKIXPublicKey(id.pub)
	if err != nil {
		return fmt.Errorf("encryption: marshal public key: %w", err)
	}

	// NOTE: convert private into pem and save atomically
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privDER})
	if err = writeFileAtomic(privPath, privPEM, privateKeyFileMode); err != nil {
		return err
	}

	// NOTE: convert private into pem and save atomically
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER})
	return writeFileAtomic(privPath+".pub", pubPEM, publicKeyFileMode)
}

// INFO: hash of the public key
func fingerprint(pub ed25519.PublicKey) string {
	sum := sha256.Sum256(pub)
	return hex.EncodeToString(sum[:])
}

func (id *Identity) Fingerprint() string {
	return fingerprint(id.pub)
}
