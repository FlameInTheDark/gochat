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
	nginxSvcList     = 3
	grafanaSvcList   = 4
	numConfigFields  = 5
)

var ( // Define globally for easy use in View
	serviceTypeChoices = []list.Item{
		item{title: "LoadBalancer", desc: "Expose via cloud load balancer (Default)"},
		item{title: "NodePort", desc: "Expose on each node's IP at a static port"},
		item{title: "ClusterIP", desc: "Expose only within the cluster (use Ingress/port-forward)"},
	}

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
	selectInstallTarget                // 2 (New)
	checkingKubePrereqs                // 3 (New) - Check helm, kubectl
	kubePrereqsFailed                  // 4 (New)
	fetchingContexts                   // 5
	fetchContextError                  // 6
	promptNamespace                    // 7
	promptMinioPass                    // 8
	promptGrafanaPass                  // 9
	selectNginxSvc                     // 10
	selectGrafanaSvc                   // 11
	selectContext                      // 12
	cloningRepo                        // 13 - Now used by both paths
	cloneError                         // 14
	installing                         // 15 - K8s specific
	installFinished                    // 16 - K8s specific
	installError                       // 17 - K8s specific
	runningCompose                     // 18 (New)
	composeFinished                    // 19 (New)
	composeError                       // 20 (New)
	noKubeContextsWarning              // 21 (New)
)

// --- List Item Delegate ---
// item implements list.Item interface
type item struct {
	title       string
	desc        string
	filterValue string
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

type kubeContextsMsg struct { // New message
	contexts []string
	err      error
}

type composeResultMsg struct { // New message
	output string
	err    error
}

// --- Model ---

type model struct {
	state   state
	spinner spinner.Model
	list    list.Model // Reused for services, context, target
	inputs  []textinput.Model
	// Prereqs
	gitOk        bool
	dockerOk     bool // New
	composeOk    bool // New
	helmOk       bool
	kubectlOk    bool
	errorMessage string
	prereqChecks map[string]bool
	// Target
	installTarget string // "kubernetes" or "docker" (New)

	// Kube Contexts
	kubeContexts    []list.Item
	selectedContext string

	// Stored selections (K8s specific)
	namespace       string
	minioPassword   string
	grafanaPassword string
	nginxSvcType    string
	grafanaSvcType  string

	// Execution Details
	helmChartPath  string
	clonedRepoPath string // Store path after cloning (New)
	helmOutput     string
	composeOutput  string // New
	finalError     error  // Maybe rename or have separate K8s/Compose errors
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
		// Create a temporary directory
		destPath, err := os.MkdirTemp("", "gochat-installer-*")
		if err != nil {
			return gitCloneResultMsg{err: fmt.Errorf("failed to create temp dir: %w", err)}
		}

		cmd := exec.Command("git", "clone", "--depth=1", repoURL, destPath)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		err = cmd.Run()
		if err != nil {
			// Attempt cleanup on failure
			_ = os.RemoveAll(destPath)
			return gitCloneResultMsg{err: fmt.Errorf("git clone failed: %w\nStderr: %s", err, stderr.String())}
		}

		return gitCloneResultMsg{path: destPath, err: nil}
	}
}

// runHelmInstallCmd now gets values directly and uses selected context
func runHelmInstallCmd(m model) tea.Cmd { // Takes model to access stored values
	return func() tea.Msg {
		namespace := m.namespace
		nginxSvcType := m.nginxSvcType
		grafanaSvcType := m.grafanaSvcType
		selectedContext := m.selectedContext // Get selected context

		if namespace == "" {
			namespace = "default"
		}
		if nginxSvcType == "" {
			nginxSvcType = "LoadBalancer"
		}
		if grafanaSvcType == "" {
			grafanaSvcType = "LoadBalancer"
		}

		args := []string{
			"upgrade", "--install", helmReleaseName, m.helmChartPath,
			"--namespace", namespace, "--create-namespace",
			"--set", fmt.Sprintf("minio.auth.rootPassword=%s", m.minioPassword),
			"--set", fmt.Sprintf("grafana.adminPassword=%s", m.grafanaPassword),
			"--set", fmt.Sprintf("nginx.service.type=%s", nginxSvcType),
			"--set", fmt.Sprintf("grafana.service.type=%s", grafanaSvcType),
		}

		// Add context flag if a context was selected
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
func getKubeContextsCmd() tea.Cmd { // New function
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
			contexts = []string{} // Handle empty output
		}

		return kubeContextsMsg{contexts: contexts, err: nil}
	}
}

// runDockerComposeCmd runs docker compose up
func runDockerComposeCmd(repoPath string) tea.Cmd { // New function
	return func() tea.Msg {
		// Determine the compose command (docker compose vs docker-compose)
		// We assume one of them exists based on prerequisite checks.
		// A more robust solution might store the detected command name in the model.
		composeCmd := "docker"
		composeArgs := []string{"compose", "up", "-d", "--build", "--remove-orphans"}
		if _, err := exec.LookPath("docker compose"); err != nil {
			// If 'docker compose' not found, assume 'docker-compose' exists
			composeCmd = "docker-compose"
			composeArgs = []string{"up", "-d", "--build", "--remove-orphans"}
		}

		cmd := exec.Command(composeCmd, composeArgs...)
		cmd.Dir = repoPath // Run command inside the cloned repository
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
	s.Style = focusedStyle // Use focusedStyle for spinner

	// --- Text Inputs ---
	inputs := make([]textinput.Model, 3)
	var t textinput.Model
	t = textinput.New()
	t.Placeholder = "default"
	t.Focus()
	t.CharLimit = 63
	t.Width = 30
	inputs[namespaceInput] = t
	t = textinput.New()
	t.Placeholder = "password (required)"
	t.EchoMode = textinput.EchoPassword
	t.CharLimit = 64
	t.Width = 30
	inputs[minioPassInput] = t
	t = textinput.New()
	t.Placeholder = "password (required)"
	t.EchoMode = textinput.EchoPassword
	t.CharLimit = 64
	t.Width = 30
	inputs[grafanaPassInput] = t

	// --- List (reused) ---
	delegate := list.NewDefaultDelegate()
	delegate.SetHeight(1)  // Each item takes one line
	delegate.SetSpacing(0) // No extra vertical spacing

	// Custom styling
	delegate.Styles.NormalTitle = itemStyle.Copy().Foreground(lipgloss.Color("240"))                     // Gray for normal title
	delegate.Styles.NormalDesc = lipgloss.NewStyle()                                                     // Clear description style
	delegate.Styles.SelectedTitle = selectedItemStyle.Copy().Foreground(lipgloss.Color("42")).Bold(true) // Green, Bold (Removed SetString)
	delegate.Styles.SelectedDesc = lipgloss.NewStyle()                                                   // Clear description style
	delegate.Styles.DimmedTitle = delegate.Styles.NormalTitle.Copy()
	delegate.Styles.DimmedDesc = delegate.Styles.NormalDesc.Copy()

	listView := list.New([]list.Item{}, delegate, 40, 5)
	listView.SetShowHelp(false)
	listView.SetShowStatusBar(false)
	listView.SetFilteringEnabled(false)
	listView.SetShowPagination(false)
	listView.Title = "Initializing..." // Placeholder title
	listView.Styles.Title = blurredStyle

	return model{
		state:        checkingPrerequisites,
		spinner:      s,
		prereqChecks: make(map[string]bool),
		inputs:       inputs,
		list:         listView,
		kubeContexts: []list.Item{},
		// Initialize selections with defaults
		nginxSvcType:   "LoadBalancer",
		grafanaSvcType: "LoadBalancer",
	}
}

// --- Bubbletea Core ---

func (m model) Init() tea.Cmd {
	// Initial checks for commands needed by *either* path initially
	return tea.Batch(
		m.spinner.Tick,
		checkCommand("git"),
		checkCommand("docker"),
		checkCommand("docker compose"), // Use 'docker compose' first
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg: // Handle window resize
		m.list.SetWidth(msg.Width / 2) // Adjust layout as needed
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "enter":
			switch m.state {
			// Prerequisites check is handled by checkMsg now
			case selectInstallTarget: // New: Handle target selection
				if i, ok := m.list.SelectedItem().(item); ok {
					m.installTarget = i.FilterValue() // Use FilterValue ("kubernetes"/"docker")
				}
				m.list.Styles.Title = blurredStyle // Blur list

				// Branch logic based on target
				if m.installTarget == "kubernetes" {
					m.state = checkingKubePrereqs
					cmds = append(cmds, m.spinner.Tick)
					// Check K8s specific commands
					cmds = append(cmds, checkCommand("helm"))
					cmds = append(cmds, checkCommand("kubectl"))
				} else if m.installTarget == "docker" {
					// Docker path: skip K8s checks/config, go to cloning
					m.state = cloningRepo
					cmds = append(cmds, m.spinner.Tick, runGitCloneCmd(gochatRepoURL))
				} else {
					// Should not happen with current list items
					m.state = prerequisitesFailed // Or some other error state
					m.errorMessage = "Invalid installation target selected."
				}
				return m, tea.Batch(cmds...)

			case checkingKubePrereqs:
				// This state is just for showing the spinner,
				// transition happens in checkMsg when helm/kubectl results arrive.
				return m, nil // No action on Enter here

			case fetchingContexts:
				// This state is just for showing the spinner,
				// transition happens in kubeContextsMsg
				return m, nil // No action on Enter here

			case promptNamespace:
				m.namespace = m.inputs[namespaceInput].Value() // Store value
				m.state = promptMinioPass
				m.inputs[namespaceInput].Blur()
				return m, m.inputs[minioPassInput].Focus()
			case promptMinioPass:
				if m.inputs[minioPassInput].Value() != "" {
					m.minioPassword = m.inputs[minioPassInput].Value() // Store value
					m.state = promptGrafanaPass
					m.inputs[minioPassInput].Blur()
					return m, m.inputs[grafanaPassInput].Focus()
				}
			case promptGrafanaPass:
				if m.inputs[grafanaPassInput].Value() != "" {
					m.grafanaPassword = m.inputs[grafanaPassInput].Value() // Store value
					m.state = selectNginxSvc
					m.inputs[grafanaPassInput].Blur()
					m.list.Title = "Select Nginx Service Type:"
					m.list.Styles.Title = focusedStyle // Focus list title
					return m, nil
				}
			case selectNginxSvc:
				if i, ok := m.list.SelectedItem().(item); ok {
					m.nginxSvcType = i.Title() // Store selection
				}
				m.state = selectGrafanaSvc
				m.list.Title = "Select Grafana Service Type:"
				// Keep list focused
				return m, nil
			case selectGrafanaSvc:
				if i, ok := m.list.SelectedItem().(item); ok {
					m.grafanaSvcType = i.Title() // Store selection
				}
				m.state = selectContext // Transition to context selection
				m.list.Title = "Select Kubernetes Context:"
				m.list.SetItems(m.kubeContexts) // Load fetched contexts into the list
				// Keep list focused
				return m, nil
			case selectContext: // New: Handle context selection
				// Store selection (handle if no contexts found?)
				if len(m.kubeContexts) > 0 {
					if i, ok := m.list.SelectedItem().(item); ok {
						m.selectedContext = i.Title()
					}
				} // If no contexts, selectedContext remains ""
				m.state = cloningRepo              // Proceed to cloning
				m.list.Styles.Title = blurredStyle // Blur list title
				cmds = append(cmds, m.spinner.Tick, runGitCloneCmd(gochatRepoURL))
				return m, tea.Batch(cmds...)
			case noKubeContextsWarning: // New: Proceed from warning
				m.state = promptNamespace
				// Prepare the list model for later use (service types)
				m.list.SetItems(serviceTypeChoices)
				return m, m.inputs[namespaceInput].Focus()
			}
		}
		// Don't fall through if Enter was handled for state change
		if msg.String() == "enter" {
			return m, tea.Batch(cmds...) // Return whatever cmds were queued (Focus changes etc)
		}

	// --- Other messages ---
	case checkMsg:
		m.prereqChecks[msg.name] = msg.ok
		if !msg.ok {
			m.errorMessage += msg.err.Error() + "\n"
		}

		switch m.state {
		case checkingPrerequisites:
			// Check if we got 'docker compose' result
			// If 'docker compose' failed, try 'docker-compose'
			if msg.name == "docker compose" && !msg.ok {
				// Clear the error from the failed 'docker compose' check
				// This is simplistic; ideally, store errors per command
				m.errorMessage = strings.Replace(m.errorMessage, msg.err.Error()+"\n", "", 1)
				delete(m.prereqChecks, "docker compose")
				// Trigger check for 'docker-compose'
				return m, checkCommand("docker-compose")
			}
			// Update specific flags
			if msg.name == "git" {
				m.gitOk = msg.ok
			}
			if msg.name == "docker" {
				m.dockerOk = msg.ok
			}
			if msg.name == "docker compose" || msg.name == "docker-compose" {
				m.composeOk = msg.ok
				// Store the successful command name for later use if needed
				// m.composeCmdName = msg.name
			}

			// Check if all initial checks are done
			_, gitDone := m.prereqChecks["git"]
			_, dockerDone := m.prereqChecks["docker"]
			// Check if either 'docker compose' or 'docker-compose' is done
			_, composeDone := m.prereqChecks["docker compose"]
			_, composeFallbackDone := m.prereqChecks["docker-compose"]
			initialComposeDone := composeDone || composeFallbackDone

			if gitDone && dockerDone && initialComposeDone {
				if m.gitOk && m.dockerOk && m.composeOk {
					m.state = selectInstallTarget
					// Prepare list for target selection
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

			// Check if K8s checks are done
			_, helmDone := m.prereqChecks["helm"]
			_, kubectlDone := m.prereqChecks["kubectl"]

			if helmDone && kubectlDone {
				if m.helmOk && m.kubectlOk {
					// K8s prereqs met, proceed to fetch contexts
					m.state = fetchingContexts
					cmds = append(cmds, m.spinner.Tick, getKubeContextsCmd())
				} else {
					m.state = kubePrereqsFailed // New specific failure state
				}
			}
		}
		return m, tea.Batch(cmds...)

	case kubeContextsMsg:
		if msg.err != nil {
			m.state = fetchContextError
			m.finalError = msg.err
		} else {
			if len(msg.contexts) == 0 { // Check if context list is empty
				m.state = noKubeContextsWarning // Transition to warning state
				m.kubeContexts = []list.Item{}  // Ensure it's empty
			} else {
				// Convert string contexts to list.Item
				items := make([]list.Item, len(msg.contexts))
				for i, ctx := range msg.contexts {
					items[i] = item{title: ctx, desc: ""} // No description for contexts
				}
				m.kubeContexts = items

				// Now ready to start config prompts
				m.state = promptNamespace
				// Prepare the list model for later use (service types)
				m.list.SetItems(serviceTypeChoices)
				// Start by focusing the first input field
				cmds = append(cmds, m.inputs[namespaceInput].Focus())
			}
		}
		return m, tea.Batch(cmds...)

	case gitCloneResultMsg:
		if msg.err != nil {
			m.state = cloneError
			m.finalError = msg.err
		} else {
			m.clonedRepoPath = msg.path // Store the path
			// Decide next step based on install target
			if m.installTarget == "kubernetes" {
				m.state = installing
				m.helmChartPath = filepath.Join(m.clonedRepoPath, "gochat-chart")
				cmds = append(cmds, m.spinner.Tick)
				cmds = append(cmds, runHelmInstallCmd(m))
			} else if m.installTarget == "docker" {
				m.state = runningCompose
				cmds = append(cmds, m.spinner.Tick)
				cmds = append(cmds, runDockerComposeCmd(m.clonedRepoPath))
			} else {
				m.state = cloneError
				m.finalError = fmt.Errorf("invalid install target '%s' after clone", m.installTarget)
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
		return m, nil

	case composeResultMsg:
		m.composeOutput = msg.output
		if msg.err != nil {
			m.state = composeError
			m.finalError = msg.err
		} else {
			m.state = composeFinished
		}
		return m, nil

	case spinner.TickMsg:
		if m.state == checkingPrerequisites || m.state == checkingKubePrereqs || m.state == fetchingContexts || m.state == cloningRepo || m.state == installing || m.state == runningCompose {
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)
	}

	// Handle updates for the active component
	cmd = m.updateActiveComponent(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// updateActiveComponent needs to handle selectInstallTarget
func (m *model) updateActiveComponent(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	switch m.state {
	case selectInstallTarget: // Added
		m.list, cmd = m.list.Update(msg)
		m.list.Styles.Title = focusedStyle
	case promptNamespace:
		m.inputs[namespaceInput], cmd = m.inputs[namespaceInput].Update(msg)
		m.inputs[namespaceInput].Focus()
		m.inputs[namespaceInput].PromptStyle = focusedStyle
		m.inputs[namespaceInput].TextStyle = focusedStyle
	case promptMinioPass:
		m.inputs[minioPassInput], cmd = m.inputs[minioPassInput].Update(msg)
		m.inputs[minioPassInput].Focus()
		m.inputs[minioPassInput].PromptStyle = focusedStyle
		m.inputs[minioPassInput].TextStyle = focusedStyle
	case promptGrafanaPass:
		m.inputs[grafanaPassInput], cmd = m.inputs[grafanaPassInput].Update(msg)
		m.inputs[grafanaPassInput].Focus()
		m.inputs[grafanaPassInput].PromptStyle = focusedStyle
		m.inputs[grafanaPassInput].TextStyle = focusedStyle
	case selectNginxSvc, selectGrafanaSvc, selectContext:
		m.list, cmd = m.list.Update(msg)
		m.list.Styles.Title = focusedStyle
	}
	return cmd
}

func (m model) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("GoChat Universal Installer") + "\n\n") // New title

	switch m.state {
	case checkingPrerequisites:
		b.WriteString(m.spinner.View() + " Checking base prerequisites (git, docker, compose)...")

	case prerequisitesFailed:
		b.WriteString("Base Prerequisites Check Failed:\n")
		checkView := func(name string, ok bool) string {
			if ok {
				return doneStyle.Render("✓ " + name + " found")
			}
			// Try to be specific about compose command
			if name == "compose" {
				return errorStyle.Render("✗ docker compose / docker-compose not found")
			}
			return errorStyle.Render("✗ " + name + " not found")
		}
		b.WriteString("  " + checkView("git", m.gitOk) + "\n")
		b.WriteString("  " + checkView("docker", m.dockerOk) + "\n")
		b.WriteString("  " + checkView("compose", m.composeOk) + "\n\n")
		b.WriteString(errorStyle.Render("Errors:\n"+m.errorMessage) + "\n")
		b.WriteString(helpStyle.Render("Please install missing commands and try again. Press q to quit."))

	case selectInstallTarget:
		b.WriteString("Base Prerequisites Met:\n")
		b.WriteString("  " + doneStyle.Render("✓ git found") + "\n")
		b.WriteString("  " + doneStyle.Render("✓ docker found") + "\n")
		b.WriteString("  " + doneStyle.Render("✓ docker compose / docker-compose found") + "\n\n")
		b.WriteString("Select Installation Target:\n")
		b.WriteString(m.list.View()) // Show the target selection list
		b.WriteString("\n" + helpStyle.Render("Use arrow keys to navigate, Enter to select, q to quit."))

	case checkingKubePrereqs:
		b.WriteString(m.spinner.View() + " Checking Kubernetes prerequisites (helm, kubectl)...")

	case kubePrereqsFailed: // New View
		b.WriteString("Kubernetes Prerequisites Check Failed:\n")
		checkView := func(name string, ok bool) string {
			if ok {
				return doneStyle.Render("✓ " + name + " found")
			}
			return errorStyle.Render("✗ " + name + " not found")
		}
		b.WriteString("  " + checkView("helm", m.helmOk) + "\n")
		b.WriteString("  " + checkView("kubectl", m.kubectlOk) + "\n\n")
		b.WriteString(errorStyle.Render("Errors:\n"+m.errorMessage) + "\n") // Display specific helm/kubectl errors
		b.WriteString(helpStyle.Render("Please install missing commands or choose Docker Compose target. Press q to quit."))

	case fetchingContexts:
		b.WriteString(m.spinner.View() + " Fetching Kubernetes contexts...\n")

	case fetchContextError:
		b.WriteString(errorStyle.Render("✗ Context Fetch Failed!\n\n"))
		if m.finalError != nil {
			b.WriteString("Error: " + m.finalError.Error() + "\n\n")
		}
		b.WriteString(helpStyle.Render("Context fetch failed. Check error above. Press q to quit."))

	case noKubeContextsWarning: // New View
		b.WriteString(errorStyle.Render("Warning: No Kubernetes contexts found!") + "\n\n")
		b.WriteString("The installer could not find any configured Kubernetes contexts using 'kubectl config get-contexts'.\n")
		b.WriteString("This usually means your KUBECONFIG environment variable is not set or the default ~/.kube/config file is missing/empty.\n\n")
		b.WriteString("You can configure kubectl access following the Kubernetes documentation for your cluster provider.\n\n")
		b.WriteString(helpStyle.Render("Press Enter to proceed with the installation using the default context (which might fail if not configured), or q to quit."))

	// --- K8s Config Steps (Only if m.installTarget == "kubernetes") ---
	case promptNamespace, promptMinioPass, promptGrafanaPass, selectNginxSvc, selectGrafanaSvc, selectContext:
		if m.installTarget != "kubernetes" {
			// Should not be in these states if target isn't K8s
			b.WriteString(errorStyle.Render("Internal state error: Unexpected configuration step for Docker target."))
			break
		}
		b.WriteString("Kubernetes Configuration Steps:\n")
		steps := []string{"Namespace", "MinIO Password", "Grafana Password", "Nginx Service", "Grafana Service", "Kube Context"}
		currentStateIndex := int(m.state) - int(promptNamespace) // Base index offset

		for i, step := range steps {
			style := itemStyle
			status := " "
			value := ""
			if i < currentStateIndex {
				status = doneStyle.Render("✓")
				// Show stored value for completed steps
				switch i { // Use step index (0, 1, 2, 3, 4)
				case 0:
					if m.namespace == "" {
						value = " (default)" // Show default if empty
					} else {
						value = fmt.Sprintf(" (%s)", m.namespace) // Namespace
					}
				case 1:
					value = " (***)" // MinIO Pass
				case 2:
					value = " (***)" // Grafana Pass
				case 3:
					value = fmt.Sprintf(" (%s)", m.nginxSvcType) // Nginx Svc
				case 4:
					value = fmt.Sprintf(" (%s)", m.grafanaSvcType) // Grafana Svc
				case 5:
					if m.selectedContext == "" {
						value = " (default)" // Show default context if none selected
					} else {
						value = fmt.Sprintf(" (%s)", m.selectedContext) // Kube Context
					}
				}
			} else if i == currentStateIndex {
				style = selectedItemStyle
			}
			b.WriteString(style.Render(fmt.Sprintf("%s %s%s", status, step, value)) + "\n")
		}
		b.WriteString("\n")

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
		case selectNginxSvc, selectGrafanaSvc:
			// Only render the active list
			b.WriteString(m.list.View())
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
		b.WriteString(helpStyle.Render("Cloning failed. Check error above. Press q to quit."))

	// --- K8s Install States ---
	case installing:
		b.WriteString(m.spinner.View() + " Installing GoChat Helm chart...\n")

	case installFinished:
		b.WriteString(doneStyle.Render("✓ Installation Successful!\n\n"))
		b.WriteString("Helm Output:\n" + m.helmOutput + "\n\n")
		b.WriteString(helpStyle.Render("Installation complete. Check Helm output for details. Press q to quit."))

	case installError:
		b.WriteString(errorStyle.Render("✗ Installation Failed!\n\n"))
		if m.finalError != nil {
			b.WriteString("Error: " + m.finalError.Error() + "\n\n")
		}
		b.WriteString("Helm Output:\n" + m.helmOutput + "\n\n")
		b.WriteString(helpStyle.Render("Installation failed. Check output for details. Press q to quit."))

	// --- Docker Compose States (New) ---
	case runningCompose:
		b.WriteString(m.spinner.View() + " Starting services with Docker Compose...")

	case composeFinished:
		b.WriteString(doneStyle.Render("✓ Docker Compose Services Started Successfully!\n\n"))
		b.WriteString("Compose Output:\n" + m.composeOutput + "\n\n")
		b.WriteString(helpStyle.Render("Services should be running in the background. Use 'docker compose logs -f' or 'docker ps' to check status. Press q to quit."))

	case composeError:
		b.WriteString(errorStyle.Render("✗ Docker Compose Failed!\n\n"))
		if m.finalError != nil {
			b.WriteString("Error: " + m.finalError.Error() + "\n\n")
		}
		b.WriteString("Compose Output:\n" + m.composeOutput + "\n\n")
		b.WriteString(helpStyle.Render("Docker Compose failed. Check output for details. Press q to quit."))
	}

	return b.String()
}

// --- Main ---

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
