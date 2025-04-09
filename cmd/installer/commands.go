package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/otiai10/copy"
)

func checkCommand(name string) tea.Cmd {
	return func() tea.Msg {
		_, err := exec.LookPath(name)

		if err != nil {
			// *** Fallback Logic for Docker Compose ***
			if name == "docker compose" {
				_, fallbackErr := exec.LookPath("docker-compose")
				if fallbackErr != nil {
					// Both failed
					return checkMsg{name: "compose", ok: false, err: fmt.Errorf("neither 'docker compose' nor 'docker-compose' found in PATH")}
				}
				return checkMsg{name: "compose", ok: true}
			}
			// *** End Fallback Logic ***

			return checkMsg{name: name, ok: false, err: fmt.Errorf("command '%s' not found in PATH", name)}
		}

		// Original command succeeded
		finalName := name
		if name == "docker compose" {
			finalName = "compose" // Use consistent name "compose"
		}
		return checkMsg{name: finalName, ok: true}
	}
}

func runGitCloneCmd(repoURL string, branch string) tea.Cmd {
	return func() tea.Msg {
		destPath, err := os.MkdirTemp("", "gochat-installer-*")
		if err != nil {
			return cloneResultMsg{err: fmt.Errorf("failed to create temp dir: %w", err)}
		}

		// --- Git Clone ---
		// Clone default branch first
		cloneArgs := []string{"clone", repoURL, destPath}
		cmdClone := exec.Command("git", cloneArgs...)
		var stderrClone bytes.Buffer
		cmdClone.Stderr = &stderrClone
		err = cmdClone.Run()
		if err != nil {
			_ = os.RemoveAll(destPath)
			return cloneResultMsg{err: fmt.Errorf("git clone failed: %w\nStderr: %s", err, stderrClone.String())}
		}

		// If a specific branch is requested, check it out
		if branch != "" {
			checkoutArgs := []string{"-C", destPath, "checkout", branch}
			cmdCheckout := exec.Command("git", checkoutArgs...)
			var stderrCheckout bytes.Buffer
			cmdCheckout.Stderr = &stderrCheckout
			err = cmdCheckout.Run()
			if err != nil {
				_ = os.RemoveAll(destPath)
				return cloneResultMsg{err: fmt.Errorf("git checkout branch '%s' failed: %w\nStderr: %s", branch, err, stderrCheckout.String())}
			}
		}

		// --- Get Git Info (from cloned repo) ---
		var branchName, tagName, shortHash string
		var gitErr error

		// Get Branch
		cmdBranch := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
		cmdBranch.Dir = destPath // Run in cloned dir
		outBranch, err := cmdBranch.Output()
		if err != nil {
			gitErr = fmt.Errorf("failed to get branch: %w", err)
		} else {
			branchName = strings.TrimSpace(string(outBranch))
		}

		// Get Tag (if on exact tag) - Redirect stderr to avoid noise if not on tag
		cmdTag := exec.Command("git", "describe", "--tags", "--exact-match")
		cmdTag.Dir = destPath
		cmdTag.Stderr = nil // Suppress "fatal: no tag exactly matches..." error output
		outTag, err := cmdTag.Output()
		// We only care if the command succeeded (exit code 0), ignore error otherwise
		if err == nil {
			tagName = strings.TrimSpace(string(outTag))
		}

		// Get Short Hash
		cmdHash := exec.Command("git", "rev-parse", "--short", "HEAD")
		cmdHash.Dir = destPath
		outHash, err := cmdHash.Output()
		if err != nil {
			if gitErr != nil { // Append error if previous one exists
				gitErr = fmt.Errorf("%w; failed to get hash: %v", gitErr, err)
			} else {
				gitErr = fmt.Errorf("failed to get hash: %w", err)
			}
		} else {
			shortHash = strings.TrimSpace(string(outHash))
		}

		// Return result, including any error from getting git info
		return cloneResultMsg{
			path:   destPath,
			branch: branchName,
			tag:    tagName,
			hash:   shortHash,
			err:    gitErr, // Return error from git info commands if clone succeeded but info failed
		}
	}
}

// Modify to accept a pointer receiver *model
func runHelmInstallCmd(m *model) tea.Cmd {
	return func() tea.Msg {
		// Access fields directly via the pointer m.fieldName
		namespace := m.namespace
		selectedContext := m.selectedContext

		// Set defaults
		if namespace == "" {
			namespace = "default"
		}

		// Build the base command arguments
		args := []string{
			"upgrade", "--install", helmReleaseName, m.helmChartPath,
			"--namespace", namespace, "--create-namespace",
			"--force",
			"--timeout", "10m",
			"--wait",
		}

		// Append the dynamically calculated arguments (includes passwords, domain, tls, dev tags etc.)
		args = append(args, m.helmArgs...)

		// Add context flag (must be added *after* potentially stored args from helmArgs)
		if selectedContext != "" {
			args = append(args, "--kube-context", selectedContext)
		}

		// --- Copy Migrations Before Helm (Keep this logic) ---
		migrationsSrcPath := filepath.Join(m.clonedRepoPath, "db", "migrations")
		migrationsDestPath := filepath.Join(m.helmChartPath, "db", "migrations")
		// Ensure parent dir exists for destination within chart
		if err := os.MkdirAll(filepath.Dir(migrationsDestPath), 0755); err != nil {
			return helmInstallMsg{output: "", err: fmt.Errorf("failed to create temp migration dir in chart: %w", err)}
		}
		if err := copy.Copy(migrationsSrcPath, migrationsDestPath); err != nil {
			// Don't fail fatally if migrations aren't found in source repo, just proceed without them
			if !os.IsNotExist(err) {
				// Clean up potentially partially copied dir on other errors
				_ = os.RemoveAll(migrationsDestPath)
				return helmInstallMsg{output: "", err: fmt.Errorf("failed to copy migrations to chart: %w", err)}
			}
			// If migrations dir doesn't exist in source, migrationsDestPath won't be created by copy, so no need to clean up
		} else {
			// Ensure cleanup happens after Helm command runs
			defer os.RemoveAll(migrationsDestPath)
		}
		// --- End Copy Migrations ---

		cmd := exec.Command("helm", args...)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		output := stdout.String() + "\n" + stderr.String()

		return helmInstallMsg{output: strings.TrimSpace(output), err: err}
	}
}

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

func getKubeIngressClassesCmd(selectedContext string) tea.Cmd {
	return func() tea.Msg {
		args := []string{"get", "ingressclass", "-o", "jsonpath='{.items[*].metadata.name}'"}
		if selectedContext != "" {
			args = append(args, "--context", selectedContext)
		}
		cmd := exec.Command("kubectl", args...)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err != nil {
			// Don't treat "no resources found" as a fatal error for classes
			if strings.Contains(stderr.String(), "no resources found") {
				return kubeIngressClassesMsg{classes: []string{}, err: nil}
			}
			return kubeIngressClassesMsg{err: fmt.Errorf("failed to get ingress classes: %w\nStderr: %s", err, stderr.String())}
		}
		// Output might have single quotes around it from jsonpath, remove them
		output := strings.Trim(strings.TrimSpace(stdout.String()), "'")
		classes := strings.Fields(output) // Split by space
		if len(classes) == 1 && classes[0] == "" {
			classes = []string{}
		}
		return kubeIngressClassesMsg{classes: classes, err: nil}
	}
}

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
			return dockerComposeMsg{output: strings.TrimSpace(output), err: fmt.Errorf("docker compose failed: %w", err)}
		}

		return dockerComposeMsg{output: strings.TrimSpace(output), err: nil}
	}
}

// --- ADDED Placeholder Command Functions ---

func checkKubernetesPrerequisites() tea.Cmd {
	return func() tea.Msg {
		_, errHelm := exec.LookPath("helm")
		_, errKubectl := exec.LookPath("kubectl")

		helmFound := errHelm == nil
		kubectlFound := errKubectl == nil
		var combinedErr error
		if !helmFound {
			combinedErr = errHelm
		}
		if !kubectlFound {
			if combinedErr != nil {
				combinedErr = fmt.Errorf("%w; %w", combinedErr, errKubectl)
			} else {
				combinedErr = errKubectl
			}
		}

		return kubePrereqsResultMsg{kubectlFound: kubectlFound, helmFound: helmFound, err: combinedErr}
	}
}

func checkDockerPrerequisites() tea.Cmd {
	return func() tea.Msg {
		_, errDocker := exec.LookPath("docker")
		_, errCompose := exec.LookPath("docker compose")
		if errCompose != nil { // Try fallback
			_, errCompose = exec.LookPath("docker-compose")
		}

		dockerFound := errDocker == nil
		dockerComposeFound := errCompose == nil
		var combinedErr error
		if !dockerFound {
			combinedErr = errDocker
		}
		if !dockerComposeFound {
			if combinedErr != nil {
				combinedErr = fmt.Errorf("%w; %w", combinedErr, errCompose)
			} else {
				combinedErr = errCompose
			}
		}

		return dockerPrereqsResultMsg{dockerFound: dockerFound, dockerComposeFound: dockerComposeFound, err: combinedErr}
	}
}

func fetchKubeContexts() tea.Cmd {
	return func() tea.Msg {
		cmdConfig := exec.Command("kubectl", "config", "view", "-o", "jsonpath={.current-context}")
		currCtxBytes, err := cmdConfig.Output()
		if err != nil {
			return kubeContextsMsg{err: fmt.Errorf("failed to get current context: %w", err)}
		}
		currentContext := strings.TrimSpace(string(currCtxBytes))

		cmdContexts := exec.Command("kubectl", "config", "get-contexts", "-o", "name")
		contextsBytes, err := cmdContexts.Output()
		if err != nil {
			return kubeContextsMsg{err: fmt.Errorf("failed to get contexts: %w", err)}
		}

		contexts := strings.Split(strings.TrimSpace(string(contextsBytes)), "\n")
		if len(contexts) == 1 && contexts[0] == "" {
			contexts = []string{}
		} // Handle empty output

		return kubeContextsMsg{contexts: contexts, currentContext: currentContext, err: nil}
	}
}

func fetchIngressClasses(context string) tea.Cmd {
	return func() tea.Msg {
		args := []string{"get", "ingressclass", "-o", "jsonpath={.items[*].metadata.name}"}
		if context != "" {
			args = append(args, "--context", context)
		}
		cmd := exec.Command("kubectl", args...)

		outputBytes, err := cmd.Output()
		if err != nil {
			// Check stderr for "no resources found" - not a fatal error for this
			if exitErr, ok := err.(*exec.ExitError); ok {
				if strings.Contains(string(exitErr.Stderr), "no resources found") {
					return kubeIngressClassesMsg{classes: []string{}, err: nil}
				}
			}
			return kubeIngressClassesMsg{err: fmt.Errorf("failed to get ingress classes: %w", err)}
		}

		output := strings.Trim(strings.TrimSpace(string(outputBytes)), "'")
		classes := strings.Fields(output)
		if len(classes) == 1 && classes[0] == "" {
			classes = []string{}
		}

		return kubeIngressClassesMsg{classes: classes, err: nil}
	}
}

// Renamed from runGitCloneCmd to avoid conflict if old one still exists
func cloneRepoCmd(branch string) tea.Cmd {
	return runGitCloneCmd(gochatRepoURL, branch)
}

// --- End ADDED Placeholder Command Functions ---
