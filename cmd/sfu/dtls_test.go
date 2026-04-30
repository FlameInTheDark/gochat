package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pion/webrtc/v4"

	"github.com/FlameInTheDark/gochat/cmd/sfu/config"
)

func TestLoadOrGenerateDTLSCertificateGeneratesDefaultCertificate(t *testing.T) {
	got, source, err := loadOrGenerateDTLSCertificate(&config.Config{})
	if err != nil {
		t.Fatalf("loadOrGenerateDTLSCertificate: %v", err)
	}
	if source != "generated" {
		t.Fatalf("certificate source = %q, want generated", source)
	}
	if got == nil {
		t.Fatal("expected generated certificate")
	}
	fingerprints, err := got.GetFingerprints()
	if err != nil {
		t.Fatalf("GetFingerprints: %v", err)
	}
	if len(fingerprints) == 0 {
		t.Fatal("expected generated certificate to expose fingerprints")
	}
}

func TestLoadOrGenerateDTLSCertificateUsesInlinePEM(t *testing.T) {
	certPEM, keyPEM, want := newTestDTLSCertificatePEM(t)

	got, source, err := loadOrGenerateDTLSCertificate(&config.Config{
		DTLSCertificatePEM: string(certPEM),
		DTLSPrivateKeyPEM:  string(keyPEM),
	})
	if err != nil {
		t.Fatalf("loadOrGenerateDTLSCertificate: %v", err)
	}
	if source != "inline_pem" {
		t.Fatalf("certificate source = %q, want inline_pem", source)
	}
	if got == nil {
		t.Fatal("expected inline certificate")
	}
	if !got.Equals(want) {
		t.Fatal("expected inline dtls certificate to match configured certificate")
	}
}

func TestLoadOrGenerateDTLSCertificateUsesFiles(t *testing.T) {
	certPEM, keyPEM, want := newTestDTLSCertificatePEM(t)

	dir := t.TempDir()
	certFile := filepath.Join(dir, "sfu.crt")
	keyFile := filepath.Join(dir, "sfu.key")
	if err := os.WriteFile(certFile, certPEM, 0o600); err != nil {
		t.Fatalf("WriteFile(%q): %v", certFile, err)
	}
	if err := os.WriteFile(keyFile, keyPEM, 0o600); err != nil {
		t.Fatalf("WriteFile(%q): %v", keyFile, err)
	}

	got, source, err := loadOrGenerateDTLSCertificate(&config.Config{
		DTLSCertificateFile: certFile,
		DTLSPrivateKeyFile:  keyFile,
	})
	if err != nil {
		t.Fatalf("loadOrGenerateDTLSCertificate: %v", err)
	}
	if source != "files" {
		t.Fatalf("certificate source = %q, want files", source)
	}
	if got == nil {
		t.Fatal("expected file-backed certificate")
	}
	if !got.Equals(want) {
		t.Fatal("expected file-backed dtls certificate to match configured certificate")
	}
}

func newTestDTLSCertificatePEM(t *testing.T) ([]byte, []byte, webrtc.Certificate) {
	t.Helper()

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}

	template := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "gochat-sfu-test"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	der, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("CreateCertificate: %v", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	if certPEM == nil {
		t.Fatal("expected certificate pem")
	}

	keyDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		t.Fatalf("MarshalPKCS8PrivateKey: %v", err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER})
	if keyPEM == nil {
		t.Fatal("expected private key pem")
	}

	leaf, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatalf("ParseCertificate: %v", err)
	}

	return certPEM, keyPEM, webrtc.CertificateFromX509(privateKey, leaf)
}
