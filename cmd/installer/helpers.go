package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// Command to force a redraw cycle
func forceRedrawCmd() tea.Cmd {
	return func() tea.Msg {
		return redrawMsg{}
	}
}

// generatePassword creates a secure random hex string (32 chars long)
func generatePassword() (string, error) {
	bytes := make([]byte, 16) // 16 bytes = 32 hex characters
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

func buildHelmCommandArgs(m model) []string {
	args := []string{
		// Base args from stored values
		"--set", fmt.Sprintf("minio.auth.rootPassword=%s", m.minioPassword),
		"--set", fmt.Sprintf("grafana.adminPassword=%s", m.grafanaPassword),
	}
	// Domain/Ingress related args
	useDomain := m.domainName != "" && m.domainName != "gochat.local"
	if useDomain {
		args = append(args, "--set", fmt.Sprintf("ingress.hostOverride=%s", m.domainName))
		if m.tlsSecretName != "" {
			args = append(args, "--set", fmt.Sprintf("ingress.tlsSecretName=%s", m.tlsSecretName))
		}
	}
	if m.ingressClassName != "" {
		args = append(args, "--set", fmt.Sprintf("ingress.className=%s", m.ingressClassName))
	}

	if m.devBranch {
		devTag := "dev"
		args = append(args,
			"--set", fmt.Sprintf("api.image.tag=%s", devTag),
			"--set", fmt.Sprintf("ui.image.tag=%s", devTag),
			"--set", fmt.Sprintf("scylla.migrate.image.tag=%s", devTag),
		)
	}

	return args
}
