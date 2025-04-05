package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// --- Constants ---
const (
	helmReleaseName = "gochat"
	// helmChartPath is now dynamic, determined after cloning
	gochatRepoURL = "https://github.com/FlameInTheDark/gochat.git"

	// Configurable fields indices
	namespaceInput   = 0
	minioPassInput   = 1
	grafanaPassInput = 2
	domainNameInput  = 3
	tlsSecretInput   = 4 // New index for TLS Secret
	numConfigFields  = 5 // Increment total text inputs
)

var ( // Define globally for easy use in View
	titleStyle        = lipgloss.NewStyle().MarginLeft(2).Bold(true)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("205"))
	helpStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).MarginTop(1)
	focusedStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	errorStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	doneStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
)

type state int

const (
	checkingPrerequisites state = iota // 0 - Check git, docker, compose
	prerequisitesFailed                // 1
	selectInstallTarget                // 2
	checkingKubePrereqs                // 3 - Check helm, kubectl
	kubePrereqsFailed                  // 4
	fetchingContexts                   // 5
	fetchContextError                  // 6
	promptNamespace                    // 7
	promptMinioPass                    // 8
	promptGrafanaPass                  // 9
	promptDomainName                   // 10
	promptTlsSecret                    // 11 (New State)
	selectContext                      // 12 (Renumbered)
	cloningRepo                        // 13 (Renumbered)
	cloneError                         // 14 (Renumbered)
	installing                         // 15 (Renumbered) - K8s specific
	installFinished                    // 16 (Renumbered) - K8s specific
	installError                       // 17 (Renumbered) - K8s specific
	runningCompose                     // 18 (Renumbered)
	composeFinished                    // 19 (Renumbered)
	composeError                       // 20 (Renumbered)
	noKubeContextsWarning              // 21 (Renumbered)
)

// --- List Item Delegate ---
// item implements list.Item interface
type item struct {
	title       string
	desc        string
	filterValue string // Used for install target ("kubernetes"/"docker")
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.filterValue }

// --- Messages ---
type checkMsg struct {
	name string
	ok   bool
	err  error
}

type installResultMsg struct {
	output string
	err    error
}

type gitCloneResultMsg struct {
	path string
	err  error
}

type kubeContextsMsg struct {
	contexts []string
	err      error
}

type composeResultMsg struct {
	output string
	err    error
}

// --- Model ---

type model struct {
	state   state
	spinner spinner.Model
	list    list.Model // Reused for target, context
	inputs  []textinput.Model
	// Prereqs
	gitOk        bool
	dockerOk     bool
	composeOk    bool
	helmOk       bool
	kubectlOk    bool
	errorMessage string
	prereqChecks map[string]bool
	// Target
	installTarget string // "kubernetes" or "docker"

	// Kube Contexts
	kubeContexts    []list.Item
	selectedContext string

	// Stored selections (K8s specific)
	namespace       string
	minioPassword   string
	grafanaPassword string
	domainName      string
	tlsSecretName   string // New field for TLS secret

	// Execution Details
	helmChartPath  string
	clonedRepoPath string
	helmOutput     string
	composeOutput  string
	finalError     error
}

// --- Helper Functions ---

func checkCommand(name string) tea.Cmd {
	return func() tea.Msg {
		_, err := exec.LookPath(name)
		if err != nil {
			return checkMsg{name: name, ok: false, err: fmt.Errorf("command '%s' not found in PATH", name)}
		}
		return checkMsg{name: name, ok: true}
	}
}

// runGitCloneCmd clones the repository
func runGitCloneCmd(repoURL string) tea.Cmd {
	return func() tea.Msg {
		destPath, err := os.MkdirTemp("", "gochat-installer-*")
		if err != nil {
			return gitCloneResultMsg{err: fmt.Errorf("failed to create temp dir: %w", err)}
		}
		cmd := exec.Command("git", "clone", "--depth=1", repoURL, destPath)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		err = cmd.Run()
		if err != nil {
			_ = os.RemoveAll(destPath)
			return gitCloneResultMsg{err: fmt.Errorf("git clone failed: %w\nStderr: %s", err, stderr.String())}
		}
		return gitCloneResultMsg{path: destPath, err: nil}
	}
}

// runHelmInstallCmd includes domain name and TLS secret settings
func runHelmInstallCmd(m model) tea.Cmd {
	return func() tea.Msg {
		namespace := m.namespace
		domainName := m.domainName
		tlsSecretName := m.tlsSecretName // Get TLS secret name
		selectedContext := m.selectedContext

		// Set defaults
		if namespace == "" {
			namespace = "default"
		}
		useDomain := domainName != "" && domainName != "gochat.local"

		args := []string{
			"upgrade", "--install", helmReleaseName, m.helmChartPath,
			"--namespace", namespace, "--create-namespace",
			"--set", fmt.Sprintf("minio.auth.rootPassword=%s", m.minioPassword),
			"--set", fmt.Sprintf("grafana.adminPassword=%s", m.grafanaPassword),
		}

		// Add the Ingress host override if a valid domain was provided
		if useDomain {
			args = append(args, "--set", fmt.Sprintf("ingress.hosts[0].host=%s", domainName))
		}

		// Configure TLS if a secret name and domain were provided
		if useDomain && tlsSecretName != "" {
			args = append(args, "--set", fmt.Sprintf("ingress.tls[0].hosts[0]=%s", domainName))
			args = append(args, "--set", fmt.Sprintf("ingress.tls[0].secretName=%s", tlsSecretName))
		} else {
			// Explicitly disable TLS in values if not configured via prompt
			// This overrides any potential default in values.yaml if user leaves secret blank
			args = append(args, "--set", "ingress.tls={}") // Set tls to an empty list/object
		}

		// Add context flag
		if selectedContext != "" {
			args = append(args, "--kube-context", selectedContext)
		}

		cmd := exec.Command("helm", args...)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		output := stdout.String() + "\n" + stderr.String()

		return installResultMsg{output: strings.TrimSpace(output), err: err}
	}
}

// getKubeContextsCmd fetches available Kubernetes contexts
func getKubeContextsCmd() tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("kubectl", "config", "get-contexts", "-o", "name")
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err != nil {
			return kubeContextsMsg{err: fmt.Errorf("failed to get kube contexts: %w\nStderr: %s", err, stderr.String())}
		}
		output := strings.TrimSpace(stdout.String())
		contexts := strings.Split(output, "\n")
		if len(contexts) == 1 && contexts[0] == "" {
			contexts = []string{}
		}
		return kubeContextsMsg{contexts: contexts, err: nil}
	}
}

// runDockerComposeCmd runs docker compose up
func runDockerComposeCmd(repoPath string) tea.Cmd {
	return func() tea.Msg {
		composeCmd := "docker"
		composeArgs := []string{"compose", "up", "-d", "--build", "--remove-orphans"}
		if _, err := exec.LookPath("docker compose"); err != nil {
			composeCmd = "docker-compose"
			composeArgs = []string{"up", "-d", "--build", "--remove-orphans"}
		}

		cmd := exec.Command(composeCmd, composeArgs...)
		cmd.Dir = repoPath
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		output := stdout.String() + "\n" + stderr.String()
		if err != nil {
			return composeResultMsg{output: strings.TrimSpace(output), err: fmt.Errorf("docker compose failed: %w", err)}
		}

		return composeResultMsg{output: strings.TrimSpace(output), err: nil}
	}
}

func initialModel() model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = focusedStyle

	// --- Text Inputs ---
	inputs := make([]textinput.Model, numConfigFields) // Now 5 inputs
	var t textinput.Model

	t = textinput.New() // Namespace (0)
	t.Placeholder = "default"
	t.Focus()
	t.CharLimit = 63
	t.Width = 30
	inputs[namespaceInput] = t

	t = textinput.New() // Minio Pass (1)
	t.Placeholder = "password (required)"
	t.EchoMode = textinput.EchoPassword
	t.CharLimit = 64
	t.Width = 30
	inputs[minioPassInput] = t

	t = textinput.New() // Grafana Pass (2)
	t.Placeholder = "password (required)"
	t.EchoMode = textinput.EchoPassword
	t.CharLimit = 64
	t.Width = 30
	inputs[grafanaPassInput] = t

	t = textinput.New() // Domain Name (3)
	t.Placeholder = "gochat.yourdomain.com (required for Ingress)"
	t.CharLimit = 253
	t.Width = 50
	inputs[domainNameInput] = t

	t = textinput.New() // TLS Secret Name (4 - NEW)
	t.Placeholder = "my-tls-secret (leave blank for HTTP)"
	t.CharLimit = 253 // Max K8s name length
	t.Width = 50
	inputs[tlsSecretInput] = t

	// --- List (reused for target, context) ---
	delegate := list.NewDefaultDelegate()
	delegate.SetHeight(1)
	delegate.SetSpacing(0)
	delegate.Styles.NormalTitle = itemStyle.Copy().Foreground(lipgloss.Color("240"))
	delegate.Styles.NormalDesc = lipgloss.NewStyle()
	delegate.Styles.SelectedTitle = selectedItemStyle.Copy().Foreground(lipgloss.Color("42")).Bold(true)
	delegate.Styles.SelectedDesc = lipgloss.NewStyle()
	delegate.Styles.DimmedTitle = delegate.Styles.NormalTitle.Copy()
	delegate.Styles.DimmedDesc = delegate.Styles.NormalDesc.Copy()

	listView := list.New([]list.Item{}, delegate, 40, 5)
	listView.SetShowHelp(false)
	listView.SetShowStatusBar(false)
	listView.SetFilteringEnabled(false)
	listView.SetShowPagination(false)
	listView.Title = "Initializing..."
	listView.Styles.Title = blurredStyle

	return model{
		state:        checkingPrerequisites,
		spinner:      s,
		prereqChecks: make(map[string]bool),
		inputs:       inputs,
		list:         listView,
		kubeContexts: []list.Item{},
		domainName:   "gochat.local",
		// tlsSecretName initialized to empty string by default
	}
}

// --- Bubbletea Core ---

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		checkCommand("git"),
		checkCommand("docker"),
		checkCommand("docker compose"), // Check preferred command first
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width / 2)
		// Resize inputs too if needed
		// for i := range m.inputs {
		//     m.inputs[i].Width = msg.Width / 3
		// }
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			// Cleanup temp dir on quit if it exists
			if m.clonedRepoPath != "" {
				_ = os.RemoveAll(m.clonedRepoPath)
			}
			return m, tea.Quit

		case "enter":
			switch m.state {
			case selectInstallTarget:
				if i, ok := m.list.SelectedItem().(item); ok {
					m.installTarget = i.FilterValue()
				} else {
					return m, nil // Should not happen
				}
				m.list.Styles.Title = blurredStyle

				if m.installTarget == "kubernetes" {
					m.state = checkingKubePrereqs
					cmds = append(cmds, m.spinner.Tick, checkCommand("helm"), checkCommand("kubectl"))
				} else if m.installTarget == "docker" {
					m.state = cloningRepo
					cmds = append(cmds, m.spinner.Tick, runGitCloneCmd(gochatRepoURL))
				} else {
					m.state = prerequisitesFailed
					m.errorMessage = "Invalid installation target selected."
				}
				return m, tea.Batch(cmds...)

			case checkingKubePrereqs, fetchingContexts:
				// Spinner states, transitions happen on msg receipt
				return m, nil

			case promptNamespace:
				m.namespace = m.inputs[namespaceInput].Value()
				m.state = promptMinioPass
				m.inputs[namespaceInput].Blur()
				return m, m.inputs[minioPassInput].Focus()

			case promptMinioPass:
				if m.inputs[minioPassInput].Value() == "" {
					return m, nil // Require password
				}
				m.minioPassword = m.inputs[minioPassInput].Value()
				m.state = promptGrafanaPass
				m.inputs[minioPassInput].Blur()
				return m, m.inputs[grafanaPassInput].Focus()

			case promptGrafanaPass:
				if m.inputs[grafanaPassInput].Value() == "" {
					return m, nil // Require password
				}
				m.grafanaPassword = m.inputs[grafanaPassInput].Value()
				m.state = promptDomainName // Go to domain name prompt
				m.inputs[grafanaPassInput].Blur()
				return m, m.inputs[domainNameInput].Focus()

			case promptDomainName:
				enteredDomain := m.inputs[domainNameInput].Value()
				if enteredDomain == "" {
					// Use default
				} else {
					m.domainName = enteredDomain
				}
				m.state = promptTlsSecret // Next: TLS Secret (NEW)
				m.inputs[domainNameInput].Blur()
				return m, m.inputs[tlsSecretInput].Focus() // Focus TLS secret input

			case promptTlsSecret: // NEW state handler
				// Store secret name (blank means no TLS)
				m.tlsSecretName = m.inputs[tlsSecretInput].Value()

				m.state = selectContext // Proceed to context selection
				m.inputs[tlsSecretInput].Blur()
				m.list.Title = "Select Kubernetes Context:"
				m.list.SetItems(m.kubeContexts)
				m.list.Styles.Title = focusedStyle
				return m, nil // Focus list

			case selectContext:
				if len(m.kubeContexts) > 0 { // Only select if list isn't empty
					if i, ok := m.list.SelectedItem().(item); ok {
						m.selectedContext = i.Title()
					}
				}
				m.state = cloningRepo
				m.list.Styles.Title = blurredStyle
				cmds = append(cmds, m.spinner.Tick, runGitCloneCmd(gochatRepoURL))
				return m, tea.Batch(cmds...)

			case noKubeContextsWarning:
				m.state = promptNamespace // Proceed to config without context selection
				// No need to prep list items here
				return m, m.inputs[namespaceInput].Focus()
			}
		}
		// Enter key was handled above for state changes
		if msg.String() == "enter" && len(cmds) == 0 {
			return m, nil // Avoid passing enter to list/input if handled
		}

	// --- Other messages ---
	case checkMsg:
		m.prereqChecks[msg.name] = msg.ok
		if !msg.ok {
			// Add specific error only if it's not the 'docker compose' fallback scenario
			if !(msg.name == "docker compose" && m.state == checkingPrerequisites) {
				m.errorMessage += msg.err.Error() + "\n"
			}
		}

		switch m.state {
		case checkingPrerequisites:
			if msg.name == "docker compose" && !msg.ok {
				// Trigger check for 'docker-compose' if 'docker compose' fails
				delete(m.prereqChecks, "docker compose") // Remove failed check
				return m, checkCommand("docker-compose")
			}

			// Update specific ok flags
			if msg.name == "git" {
				m.gitOk = msg.ok
			}
			if msg.name == "docker" {
				m.dockerOk = msg.ok
			}
			if msg.name == "docker compose" || msg.name == "docker-compose" {
				m.composeOk = msg.ok
			}

			// Check if all initial checks are done
			_, gitDone := m.prereqChecks["git"]
			_, dockerDone := m.prereqChecks["docker"]
			_, composePrimaryDone := m.prereqChecks["docker compose"]
			_, composeFallbackDone := m.prereqChecks["docker-compose"]
			initialComposeDone := composePrimaryDone || composeFallbackDone

			if gitDone && dockerDone && initialComposeDone {
				if m.gitOk && m.dockerOk && m.composeOk {
					m.state = selectInstallTarget
					targetItems := []list.Item{
						item{title: "Kubernetes (Helm)", desc: "Install into a Kubernetes cluster using Helm", filterValue: "kubernetes"},
						item{title: "Docker Compose", desc: "Run locally using Docker Compose", filterValue: "docker"},
					}
					m.list.Title = "Select Installation Target:"
					m.list.SetItems(targetItems)
					m.list.Styles.Title = focusedStyle
				} else {
					m.state = prerequisitesFailed
				}
			}

		case checkingKubePrereqs:
			if msg.name == "helm" {
				m.helmOk = msg.ok
			}
			if msg.name == "kubectl" {
				m.kubectlOk = msg.ok
			}

			_, helmDone := m.prereqChecks["helm"]
			_, kubectlDone := m.prereqChecks["kubectl"]

			if helmDone && kubectlDone {
				if m.helmOk && m.kubectlOk {
					m.state = fetchingContexts
					cmds = append(cmds, m.spinner.Tick, getKubeContextsCmd())
				} else {
					m.state = kubePrereqsFailed
				}
			}
		}
		return m, tea.Batch(cmds...)

	case kubeContextsMsg:
		if msg.err != nil {
			m.state = fetchContextError
			m.finalError = msg.err
		} else {
			items := []list.Item{}
			if len(msg.contexts) == 0 {
				m.state = noKubeContextsWarning
			} else {
				items = make([]list.Item, len(msg.contexts))
				for i, ctx := range msg.contexts {
					items[i] = item{title: ctx, desc: ""}
				}
			}
			m.kubeContexts = items
			// Don't change state here, wait for config steps
			// Prepare list model for later use in selectContext state
			m.list.SetItems(m.kubeContexts)

			// Start K8s config flow
			m.state = promptNamespace
			cmds = append(cmds, m.inputs[namespaceInput].Focus())
		}
		return m, tea.Batch(cmds...)

	case gitCloneResultMsg:
		if msg.err != nil {
			m.state = cloneError
			m.finalError = msg.err
		} else {
			m.clonedRepoPath = msg.path // Store path
			// Decide next step based on target
			if m.installTarget == "kubernetes" {
				m.state = installing
				m.helmChartPath = filepath.Join(m.clonedRepoPath, "gochat-chart")
				cmds = append(cmds, m.spinner.Tick, runHelmInstallCmd(m))
			} else if m.installTarget == "docker" {
				m.state = runningCompose
				cmds = append(cmds, m.spinner.Tick, runDockerComposeCmd(m.clonedRepoPath))
			} else {
				m.state = cloneError // Should not happen
				m.finalError = fmt.Errorf("invalid install target '%s' after clone", m.installTarget)
			}
		}
		// Cleanup function in case of error during install/compose
		if m.state == installError || m.state == composeError || m.state == cloneError {
			if m.clonedRepoPath != "" {
				_ = os.RemoveAll(m.clonedRepoPath)
			}
		}
		return m, tea.Batch(cmds...)

	case installResultMsg:
		m.helmOutput = msg.output
		if msg.err != nil {
			m.state = installError
			m.finalError = msg.err
		} else {
			m.state = installFinished
		}
		// Cleanup successful clone dir
		if m.clonedRepoPath != "" {
			_ = os.RemoveAll(m.clonedRepoPath)
		}
		return m, nil

	case composeResultMsg:
		m.composeOutput = msg.output
		if msg.err != nil {
			m.state = composeError
			m.finalError = msg.err
		} else {
			m.state = composeFinished
		}
		// Cleanup successful clone dir
		if m.clonedRepoPath != "" {
			_ = os.RemoveAll(m.clonedRepoPath)
		}
		return m, nil

	case spinner.TickMsg:
		// Only tick spinner in active spinner states
		switch m.state {
		case checkingPrerequisites, checkingKubePrereqs, fetchingContexts, cloningRepo, installing, runningCompose:
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)
	}

	// Handle updates for the active component (input field or list)
	cmd = m.updateActiveComponent(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// updateActiveComponent focuses and updates the currently active UI element
func (m *model) updateActiveComponent(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	switch m.state {
	case selectInstallTarget:
		m.list.Styles.Title = focusedStyle
		m.list, cmd = m.list.Update(msg)
	case promptNamespace:
		m.inputs[namespaceInput].PromptStyle = focusedStyle
		m.inputs[namespaceInput].TextStyle = focusedStyle
		m.inputs[namespaceInput], cmd = m.inputs[namespaceInput].Update(msg)
	case promptMinioPass:
		m.inputs[minioPassInput].PromptStyle = focusedStyle
		m.inputs[minioPassInput].TextStyle = focusedStyle
		m.inputs[minioPassInput], cmd = m.inputs[minioPassInput].Update(msg)
	case promptGrafanaPass:
		m.inputs[grafanaPassInput].PromptStyle = focusedStyle
		m.inputs[grafanaPassInput].TextStyle = focusedStyle
		m.inputs[grafanaPassInput], cmd = m.inputs[grafanaPassInput].Update(msg)
	case promptDomainName:
		m.inputs[domainNameInput].PromptStyle = focusedStyle
		m.inputs[domainNameInput].TextStyle = focusedStyle
		m.inputs[domainNameInput], cmd = m.inputs[domainNameInput].Update(msg)
	case promptTlsSecret: // NEW state
		m.inputs[tlsSecretInput].PromptStyle = focusedStyle
		m.inputs[tlsSecretInput].TextStyle = focusedStyle
		m.inputs[tlsSecretInput], cmd = m.inputs[tlsSecretInput].Update(msg)
	case selectContext:
		m.list.Styles.Title = focusedStyle
		m.list, cmd = m.list.Update(msg)
	}
	return cmd
}

func (m model) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("GoChat Universal Installer") + "\n\n")

	switch m.state {
	case checkingPrerequisites:
		b.WriteString(m.spinner.View() + " Checking base prerequisites (git, docker, compose)...")

	case prerequisitesFailed:
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
		b.WriteString("Base Prerequisites Met:\n")
		b.WriteString("  " + doneStyle.Render("✓ git found") + "\n")
		b.WriteString("  " + doneStyle.Render("✓ docker found") + "\n")
		b.WriteString("  " + doneStyle.Render("✓ docker compose / docker-compose found") + "\n\n")
		b.WriteString("Select Installation Target:\n")
		b.WriteString(m.list.View())
		b.WriteString("\n" + helpStyle.Render("Use arrow keys to navigate, Enter to select, q to quit."))

	case checkingKubePrereqs:
		b.WriteString(m.spinner.View() + " Checking Kubernetes prerequisites (helm, kubectl)...")

	case kubePrereqsFailed:
		b.WriteString("Kubernetes Prerequisites Check Failed:\n")
		checkView := func(name string, ok bool) string {
			if ok {
				return doneStyle.Render("✓ " + name + " found")
			}
			return errorStyle.Render("✗ " + name + " not found")
		}
		b.WriteString("  " + checkView("helm", m.helmOk) + "\n")
		b.WriteString("  " + checkView("kubectl", m.kubectlOk) + "\n\n")
		if m.errorMessage != "" {
			b.WriteString(errorStyle.Render("Errors:\n"+m.errorMessage) + "\n")
		}
		b.WriteString(helpStyle.Render("Please install missing commands or choose Docker Compose target. Press q to quit."))

	case fetchingContexts:
		b.WriteString(m.spinner.View() + " Fetching Kubernetes contexts...\n")

	case fetchContextError:
		b.WriteString(errorStyle.Render("✗ Context Fetch Failed!\n\n"))
		if m.finalError != nil {
			b.WriteString("Error: " + m.finalError.Error() + "\n\n")
		}
		b.WriteString(helpStyle.Render("Context fetch failed. Check error above. Press q to quit."))

	case noKubeContextsWarning:
		b.WriteString(errorStyle.Render("Warning: No Kubernetes contexts found!") + "\n\n")
		b.WriteString("The installer could not find any configured Kubernetes contexts using 'kubectl config get-contexts'.\n")
		b.WriteString("This usually means your KUBECONFIG environment variable is not set or the default ~/.kube/config file is missing/empty.\n\n")
		b.WriteString("You can configure kubectl access following the Kubernetes documentation for your cluster provider.\n\n")
		b.WriteString(helpStyle.Render("Press Enter to proceed with the installation using the default context (which might fail if not configured), or q to quit."))

	// --- K8s Config Steps (Updated) ---
	case promptNamespace, promptMinioPass, promptGrafanaPass, promptDomainName, promptTlsSecret, selectContext: // Added promptTlsSecret
		if m.installTarget != "kubernetes" {
			b.WriteString(errorStyle.Render("Internal state error: Unexpected configuration step for Docker target."))
			break // Should not happen
		}
		b.WriteString("Kubernetes Configuration Steps:\n")
		// Updated list of steps
		steps := []string{"Namespace", "MinIO Password", "Grafana Password", "Domain Name", "TLS Secret Name", "Kube Context"} // Added TLS step
		currentStateIndex := -1

		switch m.state { // Map state to step index
		case promptNamespace:
			currentStateIndex = 0
		case promptMinioPass:
			currentStateIndex = 1
		case promptGrafanaPass:
			currentStateIndex = 2
		case promptDomainName:
			currentStateIndex = 3
		case promptTlsSecret:
			currentStateIndex = 4 // New mapping
		case selectContext:
			currentStateIndex = 5 // Renumbered
		}

		for i, step := range steps {
			style := itemStyle
			status := " "
			value := ""
			if i < currentStateIndex {
				status = doneStyle.Render("✓")
				// Show stored value for completed steps
				switch i {
				case 0:
					value = fmt.Sprintf(" (%s)", m.namespace)
					if m.namespace == "" {
						value = " (default)"
					}
				case 1:
					value = " (***)" // MinIO Pass
				case 2:
					value = " (***)" // Grafana Pass
				case 3:
					value = fmt.Sprintf(" (%s)", m.domainName)
					if m.domainName == "" || m.domainName == "gochat.local" {
						value = " (default: gochat.local)"
					} // Domain Name
				case 4:
					value = fmt.Sprintf(" (%s)", m.tlsSecretName)
					if m.tlsSecretName == "" {
						value = " (HTTP only)"
					} // TLS Secret (NEW)
				case 5:
					value = fmt.Sprintf(" (%s)", m.selectedContext)
					if m.selectedContext == "" {
						value = " (default)"
					} // Kube Context
				}
			} else if i == currentStateIndex {
				style = selectedItemStyle
			}
			b.WriteString(style.Render(fmt.Sprintf("%s %s%s", status, step, value)) + "\n")
		}
		b.WriteString("\n")

		// Render the active component view
		switch m.state {
		case promptNamespace:
			b.WriteString("Enter Kubernetes Namespace (leave blank for 'default'):\n")
			b.WriteString(m.inputs[namespaceInput].View())
		case promptMinioPass:
			b.WriteString("Enter MinIO Root Password (required):\n")
			b.WriteString(m.inputs[minioPassInput].View())
		case promptGrafanaPass:
			b.WriteString("Enter Grafana Admin Password (required):\n")
			b.WriteString(m.inputs[grafanaPassInput].View())
		case promptDomainName:
			b.WriteString("Enter Domain Name for Ingress (e.g., gochat.example.com):\n")
			b.WriteString(m.inputs[domainNameInput].View())
		case promptTlsSecret: // NEW view case
			b.WriteString("Enter Kubernetes TLS Secret Name (leave blank to disable HTTPS):\n")
			b.WriteString(m.inputs[tlsSecretInput].View())
		case selectContext:
			b.WriteString(m.list.View()) // Show context list
		}
		b.WriteString("\n" + helpStyle.Render("Press Enter to confirm, Ctrl+C or q to quit."))

	case cloningRepo:
		b.WriteString(m.spinner.View() + " Cloning GoChat repository...")

	case cloneError:
		b.WriteString(errorStyle.Render("✗ Repository Cloning Failed!\n\n"))
		if m.finalError != nil {
			b.WriteString("Error: " + m.finalError.Error() + "\n\n")
		}
		// Attempt cleanup on error view
		if m.clonedRepoPath != "" {
			_ = os.RemoveAll(m.clonedRepoPath)
			m.clonedRepoPath = "" // Prevent repeat attempts
		}
		b.WriteString(helpStyle.Render("Cloning failed. Check error above. Press q to quit."))

	// --- K8s Install States ---
	case installing:
		b.WriteString(m.spinner.View() + " Installing GoChat Helm chart...\n")

	case installFinished:
		b.WriteString(doneStyle.Render("✓ Installation Successful!\n\n"))
		b.WriteString("Helm Output:\n" + m.helmOutput + "\n\n")
		// Cleanup on successful view
		if m.clonedRepoPath != "" {
			_ = os.RemoveAll(m.clonedRepoPath)
			m.clonedRepoPath = ""
		}
		b.WriteString(helpStyle.Render("Installation complete. Check Helm output for details. Press q to quit."))

	case installError:
		b.WriteString(errorStyle.Render("✗ Installation Failed!\n\n"))
		if m.finalError != nil {
			b.WriteString("Error: " + m.finalError.Error() + "\n\n")
		}
		b.WriteString("Helm Output:\n" + m.helmOutput + "\n\n")
		// Attempt cleanup on error view
		if m.clonedRepoPath != "" {
			_ = os.RemoveAll(m.clonedRepoPath)
			m.clonedRepoPath = ""
		}
		b.WriteString(helpStyle.Render("Installation failed. Check output for details. Press q to quit."))

	// --- Docker Compose States ---
	case runningCompose:
		b.WriteString(m.spinner.View() + " Starting services with Docker Compose...")

	case composeFinished:
		b.WriteString(doneStyle.Render("✓ Docker Compose Services Started Successfully!\n\n"))
		b.WriteString("Compose Output:\n" + m.composeOutput + "\n\n")
		// Cleanup on successful view
		if m.clonedRepoPath != "" {
			_ = os.RemoveAll(m.clonedRepoPath)
			m.clonedRepoPath = ""
		}
		b.WriteString(helpStyle.Render("Services should be running in the background. Use 'docker compose logs -f' or 'docker ps' to check status. Press q to quit."))

	case composeError:
		b.WriteString(errorStyle.Render("✗ Docker Compose Failed!\n\n"))
		if m.finalError != nil {
			b.WriteString("Error: " + m.finalError.Error() + "\n\n")
		}
		b.WriteString("Compose Output:\n" + m.composeOutput + "\n\n")
		// Attempt cleanup on error view
		if m.clonedRepoPath != "" {
			_ = os.RemoveAll(m.clonedRepoPath)
			m.clonedRepoPath = ""
		}
		b.WriteString(helpStyle.Render("Docker Compose failed. Check output for details. Press q to quit."))
	}

	return b.String()
}

// --- Main ---

func main() {
	// Use a model pointer
	m := initialModel()
	p := tea.NewProgram(&m) // Pass pointer to model

	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
