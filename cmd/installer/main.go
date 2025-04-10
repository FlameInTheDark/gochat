package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// --- Constants ---
const (
	helmReleaseName               = "gochat"
	gochatRepoURL                 = "https://github.com/FlameInTheDark/gochat.git"
	namespaceInput                = 0
	minioPassInput                = 1
	grafanaPassInput              = 2
	domainNameInput               = 3
	tlsSecretInput                = 4
	minioConsoleDomainInput       = 5
	minioConsoleIngressClassInput = 6
	numConfigFields               = 7
)

// initialModel no longer needs temporary style redefinitions
func initialModel(dev bool, debug bool) model {
	s := spinner.New()
	s.Spinner = spinner.Dot

	inputs := make([]textinput.Model, numConfigFields)
	var t textinput.Model

	t = textinput.New() // Namespace (0)
	t.Placeholder = "gochat"
	t.Focus()
	t.CharLimit = 63
	t.Width = 30
	t.Validate = func(s string) error {
		if !k8sNameRegex.MatchString(s) && s != "" {
			return fmt.Errorf("invalid namespace format (must be lowercase letters, numbers, -)")
		}
		return nil
	}
	inputs[namespaceInput] = t

	t = textinput.New() // Minio Pass (1)
	t.Placeholder = "password (leave blank to generate)"
	t.EchoMode = textinput.EchoPassword
	t.CharLimit = 64
	t.Width = 30
	inputs[minioPassInput] = t

	t = textinput.New() // Grafana Pass (2)
	t.Placeholder = "password (leave blank to generate)"
	t.EchoMode = textinput.EchoPassword
	t.CharLimit = 64
	t.Width = 30
	inputs[grafanaPassInput] = t

	t = textinput.New() // Domain Name (3)
	t.Placeholder = "gochat.yourdomain.com (required for Ingress)"
	t.CharLimit = 253
	t.Width = 50
	t.Validate = func(s string) error {
		if s != "" && !strings.Contains(s, ".") {
			return fmt.Errorf("invalid domain name format")
		}
		return nil
	}
	inputs[domainNameInput] = t

	t = textinput.New() // TLS Secret Name (4)
	t.Placeholder = "my-tls-secret (leave blank for HTTP)"
	t.CharLimit = 253
	t.Width = 50
	t.Validate = func(s string) error {
		if s != "" && !k8sNameRegex.MatchString(s) {
			return fmt.Errorf("invalid secret name format (must be lowercase letters, numbers, -)")
		}
		return nil
	}
	inputs[tlsSecretInput] = t

	t = textinput.New()
	t.Placeholder = "minio.localhost (leave blank to disable console ingress)"
	t.CharLimit = 253
	t.Width = 50
	t.Validate = func(s string) error {
		if s != "" && !strings.Contains(s, ".") {
			return fmt.Errorf("invalid domain name format")
		}
		return nil
	}
	inputs[minioConsoleDomainInput] = t

	// ADDED: MinIO Console Ingress Class Input (6)
	t = textinput.New()
	t.Placeholder = "nginx (or leave blank to use main ingress class)"
	t.CharLimit = 253
	t.Width = 50
	inputs[minioConsoleIngressClassInput] = t

	delegate := list.NewDefaultDelegate()
	// REMOVED: delegate.SetHeight(1)
	// REMOVED: delegate.SetSpacing(0)
	// Revert to default multi-line delegate

	// Set excessive height for target list as a diagnostic step
	listView := list.New([]list.Item{}, delegate, 40, 10) // Height set to 10
	listView.SetShowHelp(false)
	listView.SetShowStatusBar(false)
	listView.SetFilteringEnabled(false)
	listView.SetShowPagination(false)
	listView.Title = "Initializing..."

	// Initialize the single text input model
	ti := textinput.New()
	ti.Placeholder = "Initializing..." // Set an initial placeholder
	ti.CharLimit = 256                 // Example limit
	ti.Width = 60                      // INCREASED width for longer placeholders

	return model{
		state:        checkingBasePrereqs,
		spinner:      s,
		prereqChecks: make(map[string]bool),
		inputs:       inputs,
		targetList:   listView, // Use the list with height 10
		textInput:    ti,
		// Use the default delegate, keep height 10 for these lists
		// Explicitly disable filter, pagination, status bar for consistency
		ingressClassList: func() list.Model {
			l := list.New([]list.Item{}, delegate, 40, 10)
			l.SetFilteringEnabled(false)
			l.SetShowPagination(false)
			l.SetShowStatusBar(false)
			return l
		}(),
		contextList: func() list.Model {
			l := list.New([]list.Item{}, delegate, 40, 10)
			l.SetFilteringEnabled(false)
			l.SetShowPagination(false)
			l.SetShowStatusBar(false)
			return l
		}(),
		domainName:               "gochat.local", // Default value
		minioPassGenerated:       false,
		grafanaPassGenerated:     false,
		devBranch:                dev,
		debugMode:                debug,
		k8sFormFocusIndex:        0,
		k8sFormValidationErrors:  make(map[int]string),
		debug:                    debug,
		minioConsoleDomain:       "", // Default empty
		minioConsoleIngressClass: "", // Default empty
	}
}

// --- Bubbletea Core ---

// Change receiver to pointer
func (m *model) Init() tea.Cmd {
	// Return spinner tick AND dispatch initial checks
	return tea.Batch(
		m.spinner.Tick,
		checkCommand("git"),
		checkCommand("docker"),
		checkCommand("docker compose"), // Fallback handled internally
	)
}

// --- Main ---

func main() {
	var devFlag bool
	var debugFlag bool
	flag.BoolVar(&devFlag, "dev", false, "Install from the dev branch instead of default")
	flag.BoolVar(&debugFlag, "debug", false, "Print Helm command instead of executing")
	flag.Parse()

	m := initialModel(devFlag, debugFlag)
	p := tea.NewProgram(&m, tea.WithOutput(os.Stderr))

	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Alas, there's been an error: %v\n", err)
		os.Exit(1)
	}

	if finalM, ok := finalModel.(*model); ok {
		if finalM.state == installationComplete && finalM.debugMode && finalM.helmOutput != "" {
			fmt.Println("\n--- DEBUG MODE: Generated Helm Command ---")
			outputCmd := finalM.helmOutput
			if finalM.helmChartPath == "" {
				outputCmd = strings.Replace(outputCmd, "<chart-path>", "[Chart Path Pending Clone]", 1)
			} else {
				outputCmd = strings.Replace(outputCmd, "<chart-path>", "[Chart Path Pending Clone]", 1)
			}
			fmt.Println(outputCmd)
			fmt.Println("------------------------------------------")
		}
	} else if err == nil {
		fmt.Fprintf(os.Stderr, "Error: Could not cast final model type after TUI exit.\n")
		os.Exit(1)
	}
}
