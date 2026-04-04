package dtlscert

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestGenerateProducesParsablePEM(t *testing.T) {
	generated, err := Generate(GenerateOptions{
		CommonName: "gochat-sfu-test",
		ValidFor:   24 * time.Hour,
	})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	if !strings.Contains(string(generated.CertPEM), "BEGIN CERTIFICATE") {
		t.Fatalf("generated certificate pem missing header: %q", string(generated.CertPEM))
	}
	if !strings.Contains(string(generated.KeyPEM), "BEGIN PRIVATE KEY") {
		t.Fatalf("generated private key pem missing header: %q", string(generated.KeyPEM))
	}
	if generated.Leaf == nil {
		t.Fatal("expected generated leaf certificate")
	}
	if got := generated.Leaf.Subject.CommonName; got != "gochat-sfu-test" {
		t.Fatalf("generated common name = %q, want %q", got, "gochat-sfu-test")
	}
	if generated.Leaf.NotAfter.Before(generated.Leaf.NotBefore) {
		t.Fatalf("generated certificate has invalid validity window: %s..%s", generated.Leaf.NotBefore, generated.Leaf.NotAfter)
	}

	parsed, leaf, err := ParsePEM(generated.CertPEM, generated.KeyPEM)
	if err != nil {
		t.Fatalf("ParsePEM: %v", err)
	}
	if leaf == nil {
		t.Fatal("expected parsed leaf certificate")
	}
	if !parsed.Equals(generated.Certificate) {
		t.Fatal("expected parsed webrtc certificate to match generated certificate")
	}
}

func TestLoadFromFilesRoundTripsGeneratedCertificate(t *testing.T) {
	generated, err := Generate(GenerateOptions{CommonName: "gochat-sfu-files"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	dir := t.TempDir()
	certFile := filepath.Join(dir, "sfu.crt")
	keyFile := filepath.Join(dir, "sfu.key")
	if err := os.WriteFile(certFile, generated.CertPEM, 0o600); err != nil {
		t.Fatalf("WriteFile(%q): %v", certFile, err)
	}
	if err := os.WriteFile(keyFile, generated.KeyPEM, 0o600); err != nil {
		t.Fatalf("WriteFile(%q): %v", keyFile, err)
	}

	loaded, leaf, err := LoadFromFiles(certFile, keyFile)
	if err != nil {
		t.Fatalf("LoadFromFiles: %v", err)
	}
	if leaf == nil {
		t.Fatal("expected loaded leaf certificate")
	}
	if got := leaf.Subject.CommonName; got != "gochat-sfu-files" {
		t.Fatalf("loaded common name = %q, want %q", got, "gochat-sfu-files")
	}
	if !loaded.Equals(generated.Certificate) {
		t.Fatal("expected loaded webrtc certificate to match generated certificate")
	}
}
