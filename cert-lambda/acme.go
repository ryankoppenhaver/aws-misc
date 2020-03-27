package main

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"

	"github.com/go-acme/lego/providers/dns/route53"
	"golang.org/x/crypto/acme"
	"gopkg.in/square/go-jose.v2"
)

const userAgent = "cert-lambda"

type CertClient struct {
	acme.Client
	ctx    context.Context
	ToSURL string
}

func NewCertClient(directoryURL, tosURL string) *CertClient {
	key := generateKey()

	return &CertClient{
		Client: acme.Client{
			Key:          key,
			DirectoryURL: directoryURL,
			UserAgent:    userAgent,
		},
		ToSURL: tosURL,
	}
}

// acme doesn't seem to expose this, and lego doesn't provide a convenient
// way to calcuate it without instantiating the whole client :(
func (c *CertClient) calculateKeyAuth(token string) string {
	k := jose.JSONWebKey{Key: c.Key}
	thumb, err := k.Thumbprint(crypto.SHA256)
	if err != nil {
		panic(err)
	}
	thumbB64 := base64.RawURLEncoding.EncodeToString(thumb)

	return fmt.Sprintf("%s.%s", token, thumbB64)
}

func (c *CertClient) doAuthorization(authz *acme.Authorization) error {
	for _, chall := range authz.Challenges {
		if chall.Type != "dns-01" {
			continue
		}

		//p, err := route53.NewDNSProviderConfig(route53.NewDefaultConfig())
		provider, err := route53.NewDNSProvider()
		if err != nil {
			return fmt.Errorf("NewDNSProvider: %w", err)
		}

		keyAuth := c.calculateKeyAuth(chall.Token)

		err = provider.Present(authz.Identifier.Value, chall.Token, keyAuth)
		if err != nil {
			return fmt.Errorf("Present: %w", err)
		}

		_, err = c.Accept(c.ctx, chall)
		if err != nil {
			return fmt.Errorf("Accept: %w", err)
		}

		//success
		return nil
	}
	return fmt.Errorf("No supported challenges in %v", authz.Challenges)
}

func generateKey() *ecdsa.PrivateKey {
	curve := elliptic.P256()
	key, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		panic(err)
	}
	return key
}

// TODO? rate limiting/locking
func (c *CertClient) GetCert(domain string) error {
	c.ctx = context.Background()

	dir, err := c.Discover(c.ctx)
	if err != nil {
		panic(err)
	}
	if dir.Terms != c.ToSURL {
		return fmt.Errorf("unexpected ToS URL: %s", dir.Terms)
	}

	account := &acme.Account{}
	_, err = c.Register(c.ctx, account, acme.AcceptTOS) //TOS URL checked above
	if err != nil {
		return fmt.Errorf("Register: %w", err)
	}

	order, err := c.AuthorizeOrder(c.ctx, acme.DomainIDs(domain)) //XXX support multi?
	if err != nil {
		return fmt.Errorf("AuthorizeOrder: %w", err)
	}

	for i, aURL := range order.AuthzURLs {
		authz, err := c.GetAuthorization(c.ctx, aURL)
		if err != nil {
			return fmt.Errorf("getAuthorization[%d => %s]: %w", i, aURL, err)
		}
		err = c.doAuthorization(authz)
		if err != nil {
			return fmt.Errorf("Authorization failed: %w", err)
		}
	}

	order, err = c.WaitOrder(c.ctx, order.URI)
	if err != nil {
		return fmt.Errorf("WaitOrder: %w", err)
	}
	if order.Status != acme.StatusReady {
		return fmt.Errorf("After WaitOrder, expected status=ready, got %v", order.Status)
	}

	certKey := generateKey()
	ckBytes, err := x509.MarshalPKCS8PrivateKey(certKey)
	if err != nil {
		return fmt.Errorf("Failed to marshal certkey: %w", err)
	}
	//TODO
	err = pem.Encode(os.Stdout, &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: ckBytes,
	})
	if err != nil {
		return fmt.Errorf("Failed to write key PEM: %w", err)
	}

	csrTemplate := x509.CertificateRequest{
		SignatureAlgorithm: x509.ECDSAWithSHA256,
		//    Subject:
		DNSNames: []string{domain}, //XXX support multi?
		//    EmailAddresses
		//    IPAddresses
		//    URIs
		//    ExtraExtensions
		//    Attributes (deprecated)
	}
	csr, err := x509.CreateCertificateRequest(rand.Reader, &csrTemplate, certKey)
	if err != nil {
		return fmt.Errorf("CreateCertificateRequest: %w", err)
	}

	derCerts, _, err := c.CreateOrderCert(c.ctx, order.FinalizeURL, csr, true)
	if err != nil {
		return fmt.Errorf("CreateOrderCert: %w", err)
	}
	for _, cert := range derCerts {
		err = pem.Encode(os.Stdout, &pem.Block{
			Type:  "CERTIFICATE",
			Bytes: cert,
		})
	}

	err = c.DeactivateReg(c.ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[non fatal] DeactiveReg: %w\n", err)
	}

	return nil
}
