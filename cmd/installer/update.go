package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd
	var inputFocused bool

	// Check if any text input is focused
	switch m.state {
	case promptNamespace, promptDomain, promptTLSSecret, promptIngressClass, promptMinioPassword, promptGrafanaPassword:
		if m.textInput.Focused() {
			inputFocused = true
		}
	}

	// Global Quit Logic (only Ctrl+C and Esc)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			// Cleanup before quitting
			if m.clonedRepoPath != "" {
				fmt.Printf("\nCleaning up cloned repository at %s...\n", m.clonedRepoPath)
				os.RemoveAll(m.clonedRepoPath) // Best effort removal
			}
			return m, tea.Quit
		}
	}

	// State-specific logic
	switch m.state {
	case checkingBasePrereqs:
		switch msg := msg.(type) {
		case checkMsg:
			m.prereqChecks[msg.name] = msg.ok
			if !msg.ok {
				if m.errorMessage != "" {
					m.errorMessage += "; "
				}
				m.errorMessage += msg.err.Error()
			}

			_, gitDone := m.prereqChecks["git"]
			_, dockerDone := m.prereqChecks["docker"]
			composeOk, composeDone := m.prereqChecks["compose"]

			if gitDone && dockerDone && composeDone {
				m.gitOk = m.prereqChecks["git"]
				m.dockerOk = m.prereqChecks["docker"]
				m.composeOk = composeOk
				if m.gitOk && m.dockerOk && m.composeOk {
					m.state = selectInstallTarget
					m.targetList.SetItems([]list.Item{
						item{title: "Kubernetes (Helm)", desc: "Deploy to a Kubernetes cluster using Helm", filterValue: "kubernetes"},
						item{title: "Docker Compose", desc: "Deploy locally using Docker Compose", filterValue: "docker"},
					})
					m.targetList.Title = "Choose Installation Target"
				} else {
					m.state = prereqsFailed
				}
			}
		case spinner.TickMsg:
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		default:
			if keyMsg, ok := msg.(tea.KeyMsg); ok && !inputFocused {
				if keyMsg.String() == "q" {
					return m, tea.Quit
				}
			}
		}

	case selectInstallTarget:
		m.targetList, cmd = m.targetList.Update(msg)
		cmds = append(cmds, cmd)
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				i, ok := m.targetList.SelectedItem().(item)
				if ok {
					switch i.FilterValue() {
					case "kubernetes":
						m.installTarget = "kubernetes"
						m.state = checkingKubePrereqs
						cmds = append(cmds, m.spinner.Tick, checkKubernetesPrerequisites())
					case "docker":
						m.installTarget = "docker"
						m.state = cloningRepo
						branchToClone := "main"
						if m.devBranch {
							branchToClone = "dev"
						}
						cmds = append(cmds, m.spinner.Tick, cloneRepoCmd(branchToClone))
					}
				}
			case "q":
				if !inputFocused {
					return m, tea.Quit
				}
			}
		}

	case checkingKubePrereqs:
		switch msg := msg.(type) {
		case kubePrereqsResultMsg:
			m.kubePrereqs = msg
			if !msg.kubectlFound || !msg.helmFound {
				m.error = fmt.Errorf("Missing Kubernetes prerequisites. Kubectl found: %v, Helm found: %v", msg.kubectlFound, msg.helmFound)
				m.state = showError
			} else {
				m.state = fetchingKubeContexts
				cmds = append(cmds, m.spinner.Tick, fetchKubeContexts())
			}
		case spinner.TickMsg:
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	case fetchingKubeContexts:
		switch msg := msg.(type) {
		case kubeContextsMsg:
			if msg.err != nil {
				m.error = fmt.Errorf("failed to fetch kube contexts: %w", msg.err)
				m.state = showError
			} else {
				m.availableContexts = msg.contexts
				m.currentContext = msg.currentContext
				items := make([]list.Item, len(msg.contexts))
				selectedIndex := 0
				for i, ctx := range msg.contexts {
					desc := ""
					if ctx == m.currentContext {
						desc = "(current)"
						selectedIndex = i
					}
					items[i] = item{title: ctx, desc: desc}
				}
				m.contextList.SetItems(items)
				m.contextList.Title = "Select Kubernetes Context"
				m.contextList.Select(selectedIndex)
				m.state = selectKubeContext
			}
		case spinner.TickMsg:
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	case selectKubeContext:
		m.contextList, cmd = m.contextList.Update(msg)
		cmds = append(cmds, cmd)
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				i, ok := m.contextList.SelectedItem().(item)
				if ok {
					m.selectedContext = i.title
					m.state = promptNamespace
					m.textInput.Placeholder = "Enter Kubernetes namespace (e.g., gochat-dev)"
					m.textInput.Focus()
					cmds = append(cmds, textinput.Blink)
				}
			case "q":
				if !inputFocused {
					return m, tea.Quit
				}
			}
		}

	case promptNamespace:
		m.textInput, cmd = m.textInput.Update(msg)
		cmds = append(cmds, cmd)
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyEnter:
				val := strings.TrimSpace(m.textInput.Value())
				if val == "" {
					m.textInput.Placeholder = "Namespace cannot be empty. Enter Kubernetes namespace"
				} else {
					m.namespace = val
					m.textInput.Reset()
					m.textInput.Placeholder = "Enter domain name (e.g., gochat.example.com or leave blank)"
					m.state = promptDomain
				}
			}
		}

	case promptDomain:
		m.textInput, cmd = m.textInput.Update(msg)
		cmds = append(cmds, cmd)
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyEnter:
				domain := strings.TrimSpace(m.textInput.Value())
				if domain == "" {
					m.domainName = "gochat.local"
					m.state = fetchingIngressClasses
					cmds = append(cmds, m.spinner.Tick, fetchIngressClasses(m.currentContext))
				} else {
					m.domainName = domain
					m.textInput.Reset()
					m.textInput.Placeholder = "Enter TLS secret name for domain (or leave blank)"
					m.state = promptTLSSecret
				}
			}
		}

	case promptTLSSecret:
		m.textInput, cmd = m.textInput.Update(msg)
		cmds = append(cmds, cmd)
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyEnter:
				m.tlsSecretName = strings.TrimSpace(m.textInput.Value())
				m.textInput.Reset()
				m.state = fetchingIngressClasses
				cmds = append(cmds, m.spinner.Tick, fetchIngressClasses(m.currentContext))
			}
		}

	case fetchingIngressClasses:
		switch msg := msg.(type) {
		case kubeIngressClassesMsg:
			if msg.err != nil {
				m.error = fmt.Errorf("failed to fetch ingress classes: %w", msg.err)
				m.state = showError
			} else {
				m.availableIngressClasses = msg.classes
				items := []list.Item{}
				if len(msg.classes) > 0 {
					items = append(items, item{title: "(None)", desc: "Do not set an Ingress Class"})
					for _, class := range msg.classes {
						items = append(items, item{title: class, desc: ""})
					}
					m.ingressClassList.SetItems(items)
					m.ingressClassList.Title = "Select Ingress Class (Optional)"
					m.ingressClassList.ResetSelected()
					m.state = selectIngressClass
				} else {
					m.ingressClassName = ""
					m.state = promptMinioPassword
					m.textInput.Placeholder = "Enter MinIO password (or press Enter to generate)"
					m.textInput.EchoMode = textinput.EchoPassword
					m.textInput.Focus()
					cmds = append(cmds, textinput.Blink)
				}
			}
		case spinner.TickMsg:
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	case selectIngressClass:
		m.ingressClassList, cmd = m.ingressClassList.Update(msg)
		cmds = append(cmds, cmd)
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				i, ok := m.ingressClassList.SelectedItem().(item)
				if ok {
					if i.title == "(None)" {
						m.ingressClassName = ""
					} else {
						m.ingressClassName = i.title
					}
					m.state = promptMinioPassword
					m.textInput.Reset()
					m.textInput.Placeholder = "Enter MinIO password (or press Enter to generate)"
					m.textInput.EchoMode = textinput.EchoPassword
					m.textInput.Focus()
					cmds = append(cmds, textinput.Blink)
				}
			case "q":
				if !inputFocused {
					return m, tea.Quit
				}
			}
		}

	case promptMinioPassword:
		m.textInput, cmd = m.textInput.Update(msg)
		cmds = append(cmds, cmd)
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyEnter:
				password := m.textInput.Value()
				if password == "" {
					genPass, err := generatePassword()
					if err != nil {
						m.error = fmt.Errorf("failed to generate MinIO password: %w", err)
						m.state = showError
					} else {
						m.minioPassword = genPass
						m.state = promptGrafanaPassword
						m.textInput.Reset()
						m.textInput.Placeholder = "Enter Grafana password (or press Enter to generate)"
					}
				} else {
					if len(password) < 8 {
						m.textInput.Placeholder = "Password too short (min 8 chars). Enter MinIO password"
						m.textInput.SetValue("")
					} else {
						m.minioPassword = password
						m.state = promptGrafanaPassword
						m.textInput.Reset()
						m.textInput.Placeholder = "Enter Grafana password (or press Enter to generate)"
					}
				}
			}
		}

	case promptGrafanaPassword:
		m.textInput, cmd = m.textInput.Update(msg)
		cmds = append(cmds, cmd)
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyEnter:
				password := m.textInput.Value()
				if password == "" {
					genPass, err := generatePassword()
					if err != nil {
						m.error = fmt.Errorf("failed to generate Grafana password: %w", err)
						m.state = showError
					} else {
						m.grafanaPassword = genPass
						m.textInput.Blur()

						// --- Debug Check --- //
						helmArgs := buildHelmCommandArgs(*m) // Calculate args regardless
						if m.debugMode {
							// Construct the command string for debug output
							m.helmOutput = fmt.Sprintf("helm upgrade --install %s <chart-path> --namespace %s [other args: %s]",
								helmReleaseName, m.namespace, strings.Join(helmArgs, " "))
							m.state = installationComplete
							return m, tea.Quit
						} else {
							// Store args for actual install
							m.helmArgs = helmArgs
						}
						// --- End Debug Check --- //

						m.state = cloningRepo // Proceed to clone if not debug
						branchToClone := "main"
						if m.devBranch {
							branchToClone = "dev"
						}
						cmds = append(cmds, m.spinner.Tick, cloneRepoCmd(branchToClone))
					}
				} else {
					if len(password) < 8 {
						m.textInput.Placeholder = "Password too short (min 8 chars). Enter Grafana password"
						m.textInput.SetValue("")
					} else {
						m.grafanaPassword = password
						m.textInput.Blur()

						// --- Debug Check --- //
						helmArgs := buildHelmCommandArgs(*m) // Calculate args regardless
						if m.debugMode {
							// Construct the command string for debug output
							m.helmOutput = fmt.Sprintf("helm upgrade --install %s <chart-path> --namespace %s [other args: %s]",
								helmReleaseName, m.namespace, strings.Join(helmArgs, " "))
							m.state = installationComplete
							return m, tea.Quit
						} else {
							// Store args for actual install
							m.helmArgs = helmArgs
						}
						// --- End Debug Check --- //

						m.state = cloningRepo // Proceed to clone if not debug
						branchToClone := "main"
						if m.devBranch {
							branchToClone = "dev"
						}
						cmds = append(cmds, m.spinner.Tick, cloneRepoCmd(branchToClone))
					}
				}
			}
		}

	case checkingDockerPrereqs:
		switch msg := msg.(type) {
		case dockerPrereqsResultMsg:
			m.dockerPrereqs = msg
			if !msg.dockerFound || !msg.dockerComposeFound {
				m.error = fmt.Errorf("Docker prerequisites failed again? %v, %v", msg.dockerFound, msg.dockerComposeFound)
				m.state = showError
			} else {
				m.state = cloningRepo
				branchToClone := "main"
				if m.devBranch {
					branchToClone = "dev"
				}
				cmds = append(cmds, m.spinner.Tick, cloneRepoCmd(branchToClone))
			}
		case spinner.TickMsg:
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	case cloningRepo:
		switch msg := msg.(type) {
		case cloneResultMsg:
			if msg.err != nil {
				m.error = fmt.Errorf("failed to clone repository: %w", msg.err)
				m.state = showError
			} else {
				m.clonedRepoPath = msg.path
				m.gitBranch = msg.branch
				m.gitTag = msg.tag
				m.gitHash = msg.hash
				if m.debug {
					fmt.Printf("Repository cloned to: %s (%s/%s)\n", msg.path, msg.branch, msg.hash)
					os.Stdout.Sync()
				}
				if m.installTarget == "kubernetes" {
					m.state = installingHelm
					m.helmChartPath = filepath.Join(msg.path, "gochat-chart")
					installCmd := runHelmInstallCmd(m)
					cmds = append(cmds, m.spinner.Tick, installCmd)
				} else if m.installTarget == "docker" {
					m.state = runningCompose
					composeFilePath := filepath.Join(msg.path, "docker-compose.yaml")
					composeCmd := runDockerComposeCmd(composeFilePath)
					cmds = append(cmds, m.spinner.Tick, composeCmd)
				}
			}
		case spinner.TickMsg:
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	case installingHelm:
		switch msg := msg.(type) {
		case helmInstallMsg:
			m.helmOutput = msg.output
			if msg.err != nil {
				m.error = fmt.Errorf("Helm install failed: %w", msg.err)
				m.state = showError
			} else {
				m.state = installationComplete
				m.successMessage = "GoChat Helm chart installed successfully!\nNamespace: " + m.namespace
				if m.domainName != "gochat.local" {
					m.successMessage += "\nDomain: http://" + m.domainName
				} else {
					m.successMessage += "\nAccess via: kubectl port-forward service/gochat-gochat-chart-ui 8080:80 -n " + m.namespace
				}
				if m.clonedRepoPath != "" {
					os.RemoveAll(m.clonedRepoPath)
					m.clonedRepoPath = ""
				}
				return m, tea.Quit
			}
		case spinner.TickMsg:
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	case runningCompose:
		switch msg := msg.(type) {
		case dockerComposeMsg:
			m.composeOutput = msg.output
			if msg.err != nil {
				m.error = fmt.Errorf("Docker Compose failed: %w", msg.err)
				m.state = showError
			} else {
				m.state = installationComplete
				m.successMessage = "GoChat Docker Compose setup complete!\nAccess UI at http://localhost:8080"
				if m.clonedRepoPath != "" {
					os.RemoveAll(m.clonedRepoPath)
					m.clonedRepoPath = ""
				}
				return m, tea.Quit
			}
		case spinner.TickMsg:
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	case showError:
		switch msg.(type) {
		case tea.KeyMsg:
			if m.clonedRepoPath != "" {
				fmt.Printf("\nCleaning up cloned repository at %s...\n", m.clonedRepoPath)
				os.RemoveAll(m.clonedRepoPath)
			}
			return m, tea.Quit
		}

	case installationComplete:
		// No specific handling needed here, View func displays success message and app quits soon after
	}

	// Global Quit Key ('q') - Should be after state switch
	if !inputFocused {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if keyMsg.String() == "q" {
				if m.clonedRepoPath != "" {
					fmt.Printf("\nCleaning up cloned repository at %s...\n", m.clonedRepoPath)
					os.RemoveAll(m.clonedRepoPath)
				}
				return m, tea.Quit
			}
		}
	}

	return m, tea.Batch(cmds...)
}
