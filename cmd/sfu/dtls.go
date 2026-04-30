package main

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/pion/webrtc/v4"

	"github.com/FlameInTheDark/gochat/cmd/sfu/config"
	"github.com/FlameInTheDark/gochat/internal/dtlscert"
)

func loadDTLSCertificates(logger *slog.Logger, cfg *config.Config) ([]webrtc.Certificate, error) {
	certificate, source, err := loadOrGenerateDTLSCertificate(cfg)
	if err != nil {
		return nil, err
	}

	if logger != nil {
		attrs := []any{slog.String("source", source)}
		if expiresAt := certificate.Expires(); !expiresAt.IsZero() {
			attrs = append(attrs, slog.Time("expires_at", expiresAt))
		}
		if fingerprints, fpErr := certificate.GetFingerprints(); fpErr == nil && len(fingerprints) > 0 {
			attrs = append(attrs,
				slog.String("fingerprint_algorithm", fingerprints[0].Algorithm),
				slog.String("fingerprint", fingerprints[0].Value),
			)
		}
		logger.Info("configured sfu dtls certificate", attrs...)
	}

	return []webrtc.Certificate{*certificate}, nil
}

func loadOrGenerateDTLSCertificate(cfg *config.Config) (*webrtc.Certificate, string, error) {
	if cfg == nil {
		generated, err := dtlscert.Generate(dtlscert.GenerateOptions{})
		if err != nil {
			return nil, "", err
		}
		return &generated.Certificate, "generated", nil
	}

	inlineCert := strings.TrimSpace(cfg.DTLSCertificatePEM)
	inlineKey := strings.TrimSpace(cfg.DTLSPrivateKeyPEM)
	switch {
	case inlineCert != "" || inlineKey != "":
		certificate, _, err := dtlscert.ParsePEM([]byte(inlineCert), []byte(inlineKey))
		if err != nil {
			return nil, "", fmt.Errorf("parse configured dtls certificate pem: %w", err)
		}
		return certificate, "inline_pem", nil
	}

	certFile := strings.TrimSpace(cfg.DTLSCertificateFile)
	keyFile := strings.TrimSpace(cfg.DTLSPrivateKeyFile)
	switch {
	case certFile != "" || keyFile != "":
		certificate, _, err := dtlscert.LoadFromFiles(certFile, keyFile)
		if err != nil {
			return nil, "", err
		}
		return certificate, "files", nil
	}

	generated, err := dtlscert.Generate(dtlscert.GenerateOptions{
		CommonName: defaultDTLSCommonName(cfg),
	})
	if err != nil {
		return nil, "", err
	}
	return &generated.Certificate, "generated", nil
}

func defaultDTLSCommonName(cfg *config.Config) string {
	if cfg == nil {
		return dtlscert.DefaultCommonName
	}
	if serviceID := strings.TrimSpace(cfg.ServiceID); serviceID != "" {
		return serviceID
	}
	return dtlscert.DefaultCommonName
}
