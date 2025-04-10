package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Define lipgloss styles globally for the view
var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2).Bold(true)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("205"))
	helpStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).MarginTop(1)
	focusedStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	errorStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	doneStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
)

func (m *model) View() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("GoChat Universal Installer") + "\n\n")

	// --- Main State Switch --- //
	switch m.state {
	case checkingBasePrereqs:
		b.WriteString(m.spinner.View() + " Checking base prerequisites...")

	case prereqsFailed:
		b.WriteString("Base Prerequisites Check Failed:\n")
		checkView := func(name string, ok bool) string {
			if ok {
				return doneStyle.Render("✓ " + name + " found")
			}
			if name == "compose" {
				return errorStyle.Render("✗ docker compose / docker-compose not found")
			}
			return errorStyle.Render("✗ " + name + " not found")
		}
		b.WriteString("  " + checkView("git", m.gitOk) + "\n")
		b.WriteString("  " + checkView("docker", m.dockerOk) + "\n")
		b.WriteString("  " + checkView("compose", m.composeOk) + "\n\n")
		if m.errorMessage != "" {
			b.WriteString(errorStyle.Render("Errors:\n"+m.errorMessage) + "\n")
		}
		b.WriteString(helpStyle.Render("Please install missing commands and try again. Press q to quit."))

	case selectInstallTarget:
		// Show only the target list itself, no preceding text
		b.WriteString(m.targetList.View())
		b.WriteString("\n" + helpStyle.Render("Use arrow keys, Enter to select, q to quit."))

	// --- Kubernetes Path States --- //
	case checkingKubePrereqs,
		fetchingKubeContexts,
		promptNamespace,
		promptDomain,
		promptTLSSecret,
		fetchingIngressClasses,
		promptIngressClass,
		promptMinioPassword,
		promptMinioAccessKey,
		promptMinioSecretKey,
		promptMinioConsoleDomain,
		promptMinioConsoleIngressClass,
		promptGrafanaPassword,
		installingHelm:
		if m.installTarget == "kubernetes" {
			m.viewKubernetesSteps(&b)
		} else {
			b.WriteString(errorStyle.Render("Internal Error: Unexpected K8s state for non-K8s target."))
		}

	// NEW: Dedicated case for Context Selection List
	case selectKubeContext:
		b.WriteString(m.contextList.View())
		b.WriteString("\n" + helpStyle.Render("Use arrow keys, Enter to select. q to quit."))

	// NEW: Dedicated case for Ingress Class Selection List
	case selectIngressClass:
		b.WriteString(m.ingressClassList.View())
		b.WriteString("\n" + helpStyle.Render("Use arrows, Enter to select (or skip). q to quit."))

	// --- Docker Path States --- //
	case checkingDockerPrereqs:
		// This state might be removed later if base checks are sufficient
		b.WriteString(m.spinner.View() + " Checking Docker prerequisites...")

	case runningCompose:
		if m.installTarget == "docker" {
			b.WriteString(m.spinner.View() + " Starting services with Docker Compose...")
			// TODO: Add git info helper call if needed
		} else {
			b.WriteString(errorStyle.Render("Internal Error: Unexpected runningCompose state for non-Docker target."))
		}

	// --- Shared States --- //
	case cloningRepo:
		// Show spinner, target determined later for installation step
		b.WriteString(m.spinner.View() + " Cloning repository...")
		// TODO: Add git info helper call if needed

	case installationComplete:
		if !m.debugMode {
			// Display normal success message based on target
			if m.installTarget == "kubernetes" {
				b.WriteString(doneStyle.Render("✓ GoChat Helm Chart Installation Successful!") + "\n\n")
				b.WriteString(fmt.Sprintf("Namespace: %s\n", m.namespace))
				if m.domainName != "gochat.local" {
					b.WriteString(fmt.Sprintf("Access URL: http://%s\n", m.domainName))
				} else {
					b.WriteString(fmt.Sprintf("Access via: kubectl port-forward svc/gochat-gochat-chart-ui 8080:80 -n %s\n", m.namespace))
				}
				if m.minioPassGenerated {
					b.WriteString(fmt.Sprintf("Generated MinIO Password: %s\n", m.minioPassword))
				}
				if m.grafanaPassGenerated {
					b.WriteString(fmt.Sprintf("Generated Grafana Password: %s\n", m.grafanaPassword))
				}
				if m.helmOutput != "" {
					b.WriteString("\nHelm Output:\n" + m.helmOutput + "\n")
				}
			} else if m.installTarget == "docker" {
				b.WriteString(doneStyle.Render("✓ GoChat Docker Compose Setup Successful!") + "\n\n")
				b.WriteString("Access UI at http://localhost:8080\n")
				if m.composeOutput != "" {
					b.WriteString("\nCompose Output:\n" + m.composeOutput + "\n")
				}
			} else {
				b.WriteString(doneStyle.Render("✓ Installation Complete (Unknown Target?)"))
			}
			b.WriteString("\n" + helpStyle.Render("Press any key to quit."))
		} else {
			// In debug mode, just indicate the planned exit
			b.WriteString(focusedStyle.Render("Debug Mode: Exiting to print command..."))
			// The actual command print happens in main.go after Run() exits
		}

	case showError:
		b.WriteString(errorStyle.Render("✗ Error Encountered!") + "\n\n")
		if m.error != nil {
			b.WriteString(errorStyle.Render(m.error.Error()) + "\n")
		} else if m.errorMessage != "" {
			b.WriteString(errorStyle.Render(m.errorMessage) + "\n")
		}
		b.WriteString("\n" + helpStyle.Render("Press any key to quit."))

	default:
		b.WriteString(fmt.Sprintf("Unknown state: %v", m.state))
	}

	return b.String()
}

// --- Refactored viewKubernetesSteps Helper --- //
func (m *model) viewKubernetesSteps(b *strings.Builder) {
	b.WriteString("Kubernetes Configuration:\n")
	steps := []string{
		"Check Kube Prereqs",          // checkingKubePrereqs
		"Fetch Kube Contexts",         // fetchingKubeContexts
		"Select Kube Context",         // ADDED: selectKubeContext
		"Namespace",                   // promptNamespace
		"Domain Name",                 // promptDomain
		"TLS Secret Name",             // promptTLSSecret
		"Fetch Ingress Classes",       // fetchingIngressClasses
		"Select Ingress Class",        // selectIngressClass / promptIngressClass
		"MinIO Root Password",         // promptMinioPassword
		"MinIO API Access Key",        // ADDED: promptMinioAccessKey
		"MinIO API Secret Key",        // ADDED: promptMinioSecretKey
		"MinIO Console Domain",        // ADDED: 14
		"MinIO Console Ingress Class", // ADDED: 15
		"Grafana Password",            // promptGrafanaPassword (now 16)
		"Install Helm Chart",          // installingHelm (now 17)
	}

	currentStepIndex := -1
	switch m.state {
	case checkingKubePrereqs:
		b.WriteString(m.spinner.View() + " Checking Kubernetes prerequisites (helm, kubectl)...")

	case fetchingKubeContexts:
		currentStepIndex = 1
	case selectKubeContext:
		currentStepIndex = 2
	case promptNamespace:
		currentStepIndex = 3
	case promptDomain:
		currentStepIndex = 4
	case promptTLSSecret:
		currentStepIndex = 5
	case fetchingIngressClasses:
		currentStepIndex = 6
	case selectIngressClass, promptIngressClass:
		currentStepIndex = 7
	case promptMinioPassword:
		currentStepIndex = 8
	case promptMinioAccessKey:
		currentStepIndex = 9
	case promptMinioSecretKey:
		currentStepIndex = 10
	case promptMinioConsoleDomain:
		currentStepIndex = 11
	case promptMinioConsoleIngressClass:
		currentStepIndex = 12
	case promptGrafanaPassword:
		currentStepIndex = 13
	case installingHelm:
		currentStepIndex = 14
	}

	// Render the steps list...
	for i, step := range steps {
		style := itemStyle
		status := " "
		value := ""

		if currentStepIndex > i {
			status = doneStyle.Render("✓")
			// Add captured value based on step index
			switch i {
			case 2:
				value = fmt.Sprintf(" (%s)", m.selectedContext)
			case 3:
				value = fmt.Sprintf(" (%s)", m.namespace)
			case 4:
				value = fmt.Sprintf(" (%s)", m.domainName)
			case 5:
				if m.tlsSecretName != "" {
					value = fmt.Sprintf(" (%s)", m.tlsSecretName)
				} else {
					value = " (None)"
				}
			case 7:
				if m.ingressClassName != "" {
					value = fmt.Sprintf(" (%s)", m.ingressClassName)
				} else {
					value = " (None)"
				}
			case 8, 9, 10: // MinIO Pass, AccessKey, SecretKey
				value = " (set)"
			case 11:
				if m.minioConsoleDomain != "" {
					value = fmt.Sprintf(" (%s)", m.minioConsoleDomain)
				} else {
					value = " (Disabled)"
				}
			case 12: // MinIO Console Ingress Class
				if m.minioConsoleIngressClass != "" {
					value = fmt.Sprintf(" (%s)", m.minioConsoleIngressClass)
				} else {
					value = " (Default)" // Indicate default will be used
				}
			case 13: // Grafana Pass (shifted)
				value = " (set)"
			}
		} else if currentStepIndex == i {
			style = selectedItemStyle
			// Show spinner for async operations
			if m.state == checkingKubePrereqs || m.state == fetchingKubeContexts || m.state == fetchingIngressClasses || m.state == installingHelm {
				status = m.spinner.View()
			} else {
				status = ">"
			}
		}

		b.WriteString(style.Render(fmt.Sprintf("%s %s%s", status, step, value)) + "\n")
	}
	b.WriteString("\n")

	// --- Render Active Component for Current K8s State --- //
	switch m.state {
	// Prompt states use textInput
	case promptNamespace, promptDomain, promptTLSSecret, promptMinioPassword, promptMinioAccessKey, promptMinioSecretKey, promptMinioConsoleDomain, promptMinioConsoleIngressClass, promptGrafanaPassword, promptIngressClass:
		b.WriteString(m.textInput.View())
		b.WriteString("\n" + helpStyle.Render("Enter value, press Enter to confirm. Esc to quit."))

		// Selection states handled in main view switch
	}
}
