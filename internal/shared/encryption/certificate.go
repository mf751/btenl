package encryption

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"time"
)

// Long lived certificates because there is no actuall identity
// in the certificate lifetime itself. Identity comes from the
// public key carried by the certificate
const (
	certificateValidity = 365 * 24 * time.Hour
	notBeforeSkew       = time.Hour
)

// EnsureCertificate returns the TLS certificate.
// It checks the certificate on disk and if it's missing, expired, corrupt or not
// bound to the identity keypair, it generates a new self-signed one from the
// identity: the identity public key is embedded in the certificate (and the
// certificate is signed by the identity private key)
func EnsureCertificate(id *Identity) (*tls.Certificate, error) {
	certPath, err := CertsPaths()
	if err != nil {
		return nil, err
	}

	if cert, err := loadCertificate(certPath, id); err == nil {
		return cert, nil
	}

	cert, err := generateCertificate(id)
	if err != nil {
		return nil, err
	}

	return cert, nil
}

func loadCertificate(path string, id *Identity) (*tls.Certificate, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil || block.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("encryption: %s: invalid certificate pem", path)
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("encryption: %s: %w", path, err)
	}

	now := time.Now()
	if now.Before(cert.NotBefore) || now.After(cert.NotAfter) {
		return nil, fmt.Errorf("encryption: %s: certificate not valid at %v", path, now)
	}

	pub, ok := cert.PublicKey.(ed25519.PublicKey)
	if !ok || !bytes.Equal(pub, id.pub) {
		return nil, fmt.Errorf("encryption: %s: certificate not bound to identity", path)
	}

	// CheckSignature verifies against the cert's own public key, which we just
	// confirmed equals the identity public key, so this proves the identity
	// keypair signed the certificate.
	if err := cert.CheckSignature(
		cert.SignatureAlgorithm,
		cert.RawTBSCertificate,
		cert.Signature,
	); err != nil {
		return nil, fmt.Errorf("encryption: %s: certificate signature mismatch: %w", path, err)
	}

	return &tls.Certificate{
		Certificate: [][]byte{cert.Raw},
		PrivateKey:  id.priv,
		Leaf:        cert,
	}, nil
}

func generateCertificate(id *Identity) (*tls.Certificate, error) {
	serialLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serial, err := rand.Int(rand.Reader, serialLimit)
	if err != nil {
		return nil, fmt.Errorf("encryption: generate certificate serial: %w", err)
	}

	now := time.Now()

	// NOTE: To Be Signed
	template := &x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{CommonName: "btenl"},
		NotBefore:    now.Add(-notBeforeSkew),
		NotAfter:     now.Add(certificateValidity),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageClientAuth,
		},
		BasicConstraintsValid: true,
	}

	// NOTE: Returns Cert Bytes with the fileds and the public key as a field and the signature of them as a field
	der, err := x509.CreateCertificate(rand.Reader, template, template, id.pub, id.priv)
	if err != nil {
		return nil, fmt.Errorf("encryption: create certificate: %w", err)
	}

	certPath, err := CertsPaths()
	if err != nil {
		return nil, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	if err := writeFileAtomic(certPath, certPEM, publicKeyFileMode); err != nil {
		return nil, err
	}

	leaf, err := x509.ParseCertificate(der)
	if err != nil {
		return nil, fmt.Errorf("encryption: parse certificate: %w", err)
	}

	return &tls.Certificate{
		Certificate: [][]byte{der},
		PrivateKey:  id.priv,
		Leaf:        leaf,
	}, nil
}
