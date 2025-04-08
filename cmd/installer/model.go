package main

import (
	"regexp"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
)

// Define a regex for valid Kubernetes names (simple version)
var k8sNameRegex = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)

type state int

const (
	welcome             state = iota // 0
	checkingBasePrereqs              // 1 (git, docker, compose)
	prereqsFailed                    // 2
	selectInstallTarget              // 3
	// Kubernetes Path
	checkingKubePrereqs    // 4 (helm, kubectl)
	fetchingKubeContexts   // 5
	selectKubeContext      // ADDED: State for selecting Kube context
	promptNamespace        // 6 (Adjust iota)
	promptDomain           // 7
	promptTLSSecret        // 8
	fetchingIngressClasses // 9
	selectIngressClass     // 10
	promptIngressClass     // ADDED: Define state for this prompt (might be combined with select?)
	promptMinioPassword    // 11 (Adjust iota values)
	promptGrafanaPassword  // 12
	// Docker Path
	checkingDockerPrereqs // 13 (Redundant? Already checked in base)
	// Shared Path
	cloningRepo          // 14
	installingHelm       // 15 (K8s)
	runningCompose       // 16 (Docker)
	installationComplete // 17
	showError            // 18 (Generic error state)

	// Removed/Old states (keep for reference?)
	// kubePrereqsFailed
	// promptK8sConfig
	// fetchContextError
	// selectContext
	// noKubeContextsWarning
	// cloneError
	// installFinished
	// installError
	// composeFinished
	// composeError
)

// Explicitly define fetchingContexts again to see if it resolves the issue
// const fetchingContexts state = 7 // This should match the iota sequence

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

// Renamed from installResultMsg to avoid potential confusion
type helmInstallMsg struct {
	output string
	err    error
}

// Renamed from gitCloneResultMsg
type cloneResultMsg struct {
	path   string
	err    error
	branch string // New field
	tag    string // New field (empty if not on a tag)
	hash   string // New field (short hash)
}

type kubeContextsMsg struct {
	contexts       []string
	currentContext string // ADDED: Field for the current context
	err            error
}

// New message for ingress classes
type kubeIngressClassesMsg struct {
	classes []string
	err     error
}

// Renamed from composeResultMsg
type dockerComposeMsg struct {
	output string
	err    error
}

// ADDED: Define missing message types
type kubePrereqsResultMsg struct {
	kubectlFound bool
	helmFound    bool
	err          error
}

type dockerPrereqsResultMsg struct {
	dockerFound        bool
	dockerComposeFound bool
	err                error
}

// Message to signal a view redraw may be needed
type redrawMsg struct{}

// --- Model ---

type model struct {
	state   state
	spinner spinner.Model
	// list    list.Model // Removed
	// Use separate lists for each selection context
	targetList       list.Model
	ingressClassList list.Model
	contextList      list.Model
	textInput        textinput.Model
	errorViewport    viewport.Model // For scrollable errors
	ready            bool           // For initial sizing
	inputs           []textinput.Model
	// Prereqs
	gitOk         bool
	dockerOk      bool
	composeOk     bool
	helmOk        bool
	kubectlOk     bool
	kubePrereqs   kubePrereqsResultMsg   // ADDED: Store kube prereq results
	dockerPrereqs dockerPrereqsResultMsg // ADDED: Store docker prereq results
	errorMessage  string
	prereqChecks  map[string]bool
	// Target
	installTarget string // "kubernetes" or "docker"

	// Kube Contexts & Classes
	availableContexts       []string // ADDED: To store fetched contexts
	currentContext          string   // ADDED: To store current context
	selectedContext         string
	availableIngressClasses []string // ADDED: To store fetched ingress classes
	ingressClassName        string   // Store the selected class name

	// Stored selections (K8s specific)
	namespace            string
	minioPassword        string
	grafanaPassword      string
	domainName           string
	tlsSecretName        string
	minioPassGenerated   bool // Flag to indicate if MinIO pass was generated
	grafanaPassGenerated bool // Flag to indicate if Grafana pass was generated

	// Execution Details
	helmChartPath  string
	clonedRepoPath string
	// clonedUiRepoPath string // REMOVED
	// builtUiImageName string // REMOVED
	gitBranch      string
	gitTag         string // New field
	gitHash        string // New field
	helmOutput     string // Stored output for install/debug
	composeOutput  string
	finalError     error
	successMessage string // ADDED: For completion messages
	error          error  // ADDED: General error field

	// Flags passed from main
	devBranch bool
	debugMode bool // ADDED: Debug flag from main
	debug     bool // ADDED: Alias for debugMode? Or separate? Assuming alias for now.

	// New fields for K8s config form management
	k8sFormFocusIndex       int            // Index of the currently focused input in the K8s form
	k8sFormValidationErrors map[int]string // Map input index to validation error message
}
