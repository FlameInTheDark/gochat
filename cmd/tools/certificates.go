package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	cli "github.com/urfave/cli/v3"

	"github.com/FlameInTheDark/gochat/internal/dtlscert"
)

type dtlsGenerateResult struct {
	DTLSCertificateFile  string    `json:"dtls_certificate_file"`
	DTLSPrivateKeyFile   string    `json:"dtls_private_key_file"`
	CommonName           string    `json:"common_name"`
	ExpiresAt            time.Time `json:"expires_at"`
	FingerprintAlgorithm string    `json:"fingerprint_algorithm,omitempty"`
	Fingerprint          string    `json:"fingerprint,omitempty"`
}

func certificates() *cli.Command {
	return &cli.Command{
		Name:    "certificates",
		Aliases: []string{"certs"},
		Usage:   "Generate certificates for service configuration",
		Commands: []*cli.Command{
			dtlsCertificates(),
		},
	}
}

func dtlsCertificates() *cli.Command {
	return &cli.Command{
		Name:  "dtls",
		Usage: "DTLS certificate helpers for SFU WebRTC transport",
		Commands: []*cli.Command{
			generateDTLSCertificateCommand(),
		},
	}
}

func generateDTLSCertificateCommand() *cli.Command {
	return &cli.Command{
		Name:    "generate",
		Aliases: []string{"gen"},
		Usage:   "Generate a DTLS certificate/key pair and write PEM files for SFU config",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "cert-out", Usage: "output path for the certificate PEM file", Required: true},
			&cli.StringFlag{Name: "key-out", Usage: "output path for the private key PEM file", Required: true},
			&cli.StringFlag{Name: "common-name", Aliases: []string{"cn"}, Value: dtlscert.DefaultCommonName, Usage: "certificate common name"},
			&cli.DurationFlag{Name: "valid-for", Value: dtlscert.DefaultValidFor, Usage: "certificate validity duration"},
			&cli.BoolFlag{Name: "overwrite", Usage: "overwrite existing output files"},
			&cli.StringFlag{Name: "format", Aliases: []string{"f"}, Value: "text", Usage: "output: text|json"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			result, err := generateAndWriteDTLSCertificatePair(
				cmd.String("cert-out"),
				cmd.String("key-out"),
				cmd.String("common-name"),
				cmd.Duration("valid-for"),
				cmd.Bool("overwrite"),
			)
			if err != nil {
				return err
			}
			return printDTLSGenerateResult(os.Stdout, cmd.String("format"), result)
		},
	}
}

func generateAndWriteDTLSCertificatePair(certOut, keyOut, commonName string, validFor time.Duration, overwrite bool) (dtlsGenerateResult, error) {
	certPath, err := absoluteOutputPath(certOut)
	if err != nil {
		return dtlsGenerateResult{}, fmt.Errorf("resolve cert-out: %w", err)
	}
	keyPath, err := absoluteOutputPath(keyOut)
	if err != nil {
		return dtlsGenerateResult{}, fmt.Errorf("resolve key-out: %w", err)
	}
	if strings.EqualFold(certPath, keyPath) {
		return dtlsGenerateResult{}, fmt.Errorf("cert-out and key-out must be different paths")
	}
	if validFor <= 0 {
		return dtlsGenerateResult{}, fmt.Errorf("valid-for must be greater than zero")
	}
	if err := ensureWritableOutputPath(certPath, overwrite); err != nil {
		return dtlsGenerateResult{}, err
	}
	if err := ensureWritableOutputPath(keyPath, overwrite); err != nil {
		return dtlsGenerateResult{}, err
	}

	generated, err := dtlscert.Generate(dtlscert.GenerateOptions{
		CommonName: commonName,
		ValidFor:   validFor,
	})
	if err != nil {
		return dtlsGenerateResult{}, err
	}

	if err := os.MkdirAll(filepath.Dir(certPath), 0o755); err != nil {
		return dtlsGenerateResult{}, fmt.Errorf("create certificate output directory: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(keyPath), 0o755); err != nil {
		return dtlsGenerateResult{}, fmt.Errorf("create private key output directory: %w", err)
	}
	if err := os.WriteFile(certPath, generated.CertPEM, 0o644); err != nil {
		return dtlsGenerateResult{}, fmt.Errorf("write certificate file %q: %w", certPath, err)
	}
	if err := os.WriteFile(keyPath, generated.KeyPEM, 0o600); err != nil {
		return dtlsGenerateResult{}, fmt.Errorf("write private key file %q: %w", keyPath, err)
	}

	result := dtlsGenerateResult{
		DTLSCertificateFile: certPath,
		DTLSPrivateKeyFile:  keyPath,
		CommonName:          generated.Leaf.Subject.CommonName,
		ExpiresAt:           generated.Leaf.NotAfter.UTC(),
	}
	if fingerprints, fpErr := generated.Certificate.GetFingerprints(); fpErr == nil && len(fingerprints) > 0 {
		result.FingerprintAlgorithm = fingerprints[0].Algorithm
		result.Fingerprint = fingerprints[0].Value
	}

	return result, nil
}

func printDTLSGenerateResult(w io.Writer, format string, result dtlsGenerateResult) error {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "", "text":
		_, err := fmt.Fprintf(w,
			"dtls_certificate_file=%s\n"+
				"dtls_private_key_file=%s\n"+
				"common_name=%s\n"+
				"expires_at=%s\n"+
				"fingerprint_algorithm=%s\n"+
				"fingerprint=%s\n",
			result.DTLSCertificateFile,
			result.DTLSPrivateKeyFile,
			result.CommonName,
			result.ExpiresAt.Format(time.RFC3339),
			result.FingerprintAlgorithm,
			result.Fingerprint,
		)
		return err
	case "json":
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

func absoluteOutputPath(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", fmt.Errorf("path is required")
	}
	return filepath.Abs(path)
}

func ensureWritableOutputPath(path string, overwrite bool) error {
	if _, err := os.Stat(path); err == nil && !overwrite {
		return fmt.Errorf("output file %q already exists (use --overwrite)", path)
	} else if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("stat output file %q: %w", path, err)
	}
	return nil
}
