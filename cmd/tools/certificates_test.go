package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/FlameInTheDark/gochat/internal/dtlscert"
)

func TestGenerateAndWriteDTLSCertificatePairWritesFiles(t *testing.T) {
	dir := t.TempDir()
	certFile := filepath.Join(dir, "sfu.crt")
	keyFile := filepath.Join(dir, "sfu.key")

	result, err := generateAndWriteDTLSCertificatePair(certFile, keyFile, "gochat-sfu-test", 24*time.Hour, false)
	if err != nil {
		t.Fatalf("generateAndWriteDTLSCertificatePair: %v", err)
	}

	certPEM, err := os.ReadFile(certFile)
	if err != nil {
		t.Fatalf("ReadFile(%q): %v", certFile, err)
	}
	keyPEM, err := os.ReadFile(keyFile)
	if err != nil {
		t.Fatalf("ReadFile(%q): %v", keyFile, err)
	}

	if !strings.Contains(string(certPEM), "BEGIN CERTIFICATE") {
		t.Fatalf("certificate pem missing header: %q", string(certPEM))
	}
	if !strings.Contains(string(keyPEM), "BEGIN PRIVATE KEY") {
		t.Fatalf("private key pem missing header: %q", string(keyPEM))
	}

	_, leaf, err := dtlscert.ParsePEM(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("ParsePEM: %v", err)
	}
	if leaf == nil {
		t.Fatal("expected parsed leaf certificate")
	}
	if got := leaf.Subject.CommonName; got != "gochat-sfu-test" {
		t.Fatalf("certificate common name = %q, want %q", got, "gochat-sfu-test")
	}

	absCertFile, err := filepath.Abs(certFile)
	if err != nil {
		t.Fatalf("Abs(%q): %v", certFile, err)
	}
	if result.DTLSCertificateFile != absCertFile {
		t.Fatalf("result certificate path = %q, want %q", result.DTLSCertificateFile, absCertFile)
	}
	if result.CommonName != "gochat-sfu-test" {
		t.Fatalf("result common name = %q, want %q", result.CommonName, "gochat-sfu-test")
	}
	if result.ExpiresAt.IsZero() {
		t.Fatal("expected expiration timestamp in result")
	}
}

func TestGenerateAndWriteDTLSCertificatePairRejectsExistingFilesWithoutOverwrite(t *testing.T) {
	dir := t.TempDir()
	certFile := filepath.Join(dir, "sfu.crt")
	keyFile := filepath.Join(dir, "sfu.key")
	if err := os.WriteFile(certFile, []byte("existing"), 0o600); err != nil {
		t.Fatalf("WriteFile(%q): %v", certFile, err)
	}

	if _, err := generateAndWriteDTLSCertificatePair(certFile, keyFile, "", time.Hour, false); err == nil {
		t.Fatal("expected existing certificate file to require --overwrite")
	}
}

func TestPrintDTLSGenerateResultJSON(t *testing.T) {
	var buf bytes.Buffer
	result := dtlsGenerateResult{
		DTLSCertificateFile:  "/tmp/sfu.crt",
		DTLSPrivateKeyFile:   "/tmp/sfu.key",
		CommonName:           "gochat-sfu",
		ExpiresAt:            time.Unix(1700000000, 0).UTC(),
		FingerprintAlgorithm: "sha-256",
		Fingerprint:          "AA:BB",
	}

	if err := printDTLSGenerateResult(&buf, "json", result); err != nil {
		t.Fatalf("printDTLSGenerateResult: %v", err)
	}

	var decoded dtlsGenerateResult
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if decoded.DTLSCertificateFile != result.DTLSCertificateFile || decoded.DTLSPrivateKeyFile != result.DTLSPrivateKeyFile {
		t.Fatalf("decoded result mismatch: %#v", decoded)
	}
}
