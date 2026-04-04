package dtlscert

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/pion/webrtc/v4"
)

const (
	DefaultCommonName = "gochat-sfu"
	DefaultValidFor   = 365 * 24 * time.Hour
	defaultBackdate   = time.Hour
)

type GenerateOptions struct {
	CommonName string
	ValidFor   time.Duration
}

type Generated struct {
	Certificate webrtc.Certificate
	Leaf        *x509.Certificate
	CertPEM     []byte
	KeyPEM      []byte
}

func Generate(opts GenerateOptions) (*Generated, error) {
	commonName := strings.TrimSpace(opts.CommonName)
	if commonName == "" {
		commonName = DefaultCommonName
	}

	validFor := opts.ValidFor
	if validFor <= 0 {
		validFor = DefaultValidFor
	}

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generate dtls private key: %w", err)
	}

	serialNumber, err := randomSerialNumber()
	if err != nil {
		return nil, fmt.Errorf("generate certificate serial number: %w", err)
	}

	notBefore := time.Now().Add(-defaultBackdate)
	template := &x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               pkix.Name{CommonName: commonName},
		Issuer:                pkix.Name{CommonName: commonName},
		NotBefore:             notBefore,
		NotAfter:              notBefore.Add(validFor),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	der, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, fmt.Errorf("create dtls x509 certificate: %w", err)
	}

	leaf, err := x509.ParseCertificate(der)
	if err != nil {
		return nil, fmt.Errorf("parse generated dtls certificate: %w", err)
	}

	keyDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("marshal dtls private key: %w", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	if certPEM == nil {
		return nil, fmt.Errorf("encode dtls certificate pem")
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER})
	if keyPEM == nil {
		return nil, fmt.Errorf("encode dtls private key pem")
	}

	return &Generated{
		Certificate: webrtc.CertificateFromX509(privateKey, leaf),
		Leaf:        leaf,
		CertPEM:     certPEM,
		KeyPEM:      keyPEM,
	}, nil
}

func ParsePEM(certPEM, keyPEM []byte) (*webrtc.Certificate, *x509.Certificate, error) {
	keyPair, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, nil, fmt.Errorf("parse x509 key pair: %w", err)
	}

	leaf := keyPair.Leaf
	if leaf == nil {
		if len(keyPair.Certificate) == 0 {
			return nil, nil, fmt.Errorf("x509 key pair missing certificate chain")
		}

		leaf, err = x509.ParseCertificate(keyPair.Certificate[0])
		if err != nil {
			return nil, nil, fmt.Errorf("parse leaf certificate: %w", err)
		}
	}

	certificate := webrtc.CertificateFromX509(keyPair.PrivateKey, leaf)
	return &certificate, leaf, nil
}

func LoadFromFiles(certFile, keyFile string) (*webrtc.Certificate, *x509.Certificate, error) {
	certPEM, err := os.ReadFile(certFile)
	if err != nil {
		return nil, nil, fmt.Errorf("read dtls certificate file %q: %w", certFile, err)
	}
	keyPEM, err := os.ReadFile(keyFile)
	if err != nil {
		return nil, nil, fmt.Errorf("read dtls private key file %q: %w", keyFile, err)
	}

	certificate, leaf, err := ParsePEM(certPEM, keyPEM)
	if err != nil {
		return nil, nil, fmt.Errorf("parse dtls certificate files: %w", err)
	}
	return certificate, leaf, nil
}

func randomSerialNumber() (*big.Int, error) {
	maxBigInt := new(big.Int)
	maxBigInt.Exp(big.NewInt(2), big.NewInt(130), nil).Sub(maxBigInt, big.NewInt(1))
	return rand.Int(rand.Reader, maxBigInt)
}
