package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"golang.org/x/crypto/ssh"
	"strings"
	"time"
)

const expectedKeyType = "ssh-ed25519"
const validityPeriod = 60 * 60 * 24 * 90

// A Signer stores a parsed CA key, allowing it to be used to sign certificates
type Signer struct {
	privateKey ssh.Signer
}

// NewSigner accepts a string containing an ed25519 SSH private key in PEM
// format, as produced by ssh-keygen, and returns a new Signer which will use
// this key when signing certs in MakeCertificate.
func NewSigner(caKey string) (*Signer, error) {
	privateKey, err := ssh.ParsePrivateKey([]byte(caKey))
	if err != nil {
		return nil, fmt.Errorf("parsing CA key: %w", err)
	}
	keyType := privateKey.PublicKey().Type()
	if keyType != expectedKeyType {
		return nil, fmt.Errorf("Expected CA key type to be %q, got %q", expectedKeyType, keyType)
	}
	return &Signer{
		privateKey: privateKey,
	}, nil
}

// MakeCertificate accepts a ed25519 SSH public key in the format output by ssh-keygen, and
// a list of identities.  It returns a signed certificate in the same format.
func (s *Signer) MakeCertificate(publicKey string, identities []string) (string, error) {
	pk, err := parseKey(publicKey)
	if err != nil {
		return "", fmt.Errorf("parsing public key: %w", err)
	}
	keyType := pk.Type()
	if keyType != expectedKeyType {
		return "", fmt.Errorf("expected key type %q, got %q", expectedKeyType, keyType)
	}

	if len(identities) == 0 {
		return "", fmt.Errorf("identities must be non-empty")
	}

	now := time.Now().UTC()
	ts := uint64(now.Unix())
	keyID := fmt.Sprintf("%s at %s", identities[0], now.Format(time.RFC3339))

	c := ssh.Certificate{
		Key:             pk,
		Serial:          0,
		CertType:        ssh.HostCert,
		KeyId:           keyID,
		ValidPrincipals: identities,
		ValidAfter:      ts,
		ValidBefore:     ts + validityPeriod,
	}
	err = c.SignCert(rand.Reader, s.privateKey)
	if err != nil {
		return "", fmt.Errorf("signing cert: %w", err)
	}

	return fmt.Sprintf("ssh-ed25519-cert-v01@openssh.com %s %s\n",
		base64.StdEncoding.EncodeToString(c.Marshal()),
		keyID,
	), nil
}

func parseKey(key string) (ssh.PublicKey, error) {
	keyParts := strings.Split(key, " ")
	if len(keyParts) < 2 {
		return nil, fmt.Errorf("too few elements")
	}
	if keyParts[0] != expectedKeyType {
		return nil, fmt.Errorf("expected header %q, got %q", expectedKeyType, keyParts[0])
	}

	keyBytes, err := base64.StdEncoding.DecodeString(keyParts[1])
	if err != nil {
		return nil, fmt.Errorf("decoding base64: %w", err)
	}

	k, err := ssh.ParsePublicKey(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("parsing pubkey bytes: %w", err)
	}
	return k, nil
}
