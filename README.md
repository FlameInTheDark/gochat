# GoChat

GoChat is a real-time chat application built with Go.

## Installation

There are two primary ways to install and run GoChat:

1.  **Universal Installer (Recommended):** A command-line tool that guides you through installing GoChat on either Kubernetes (using Helm) or locally with Docker Compose.
2.  **Manual Setup:** For development or custom environments.

### Universal Installer

Located in `cmd/installer`, this tool provides a terminal UI to guide users through installing GoChat using either Docker Compose (for local development/testing) or Helm (for Kubernetes deployment).

**Quick Start (Recommended):**

If pre-built binaries are available on the [**GitHub Releases**](https://github.com/FlameInTheDark/gochat/releases) page, you can use a one-liner (example for Linux amd64):

```bash
# Download latest installer, make executable, run
curl -L -o gochat-installer https://github.com/FlameInTheDark/gochat/releases/latest/download/installer-linux-amd64 && \
chmod +x gochat-installer && \
./gochat-installer

# To install from the dev branch (if available as a pre-built binary or if building manually):
# ./gochat-installer --dev
```
*(Note: Check the Releases page for binaries for other OS/Architectures. You might need `sudo` depending on permissions.)*

**Manual Build and Run:**

```bash
# Clone the repo (if you haven't already)
# git clone https://github.com/FlameInTheDark/gochat.git
# cd gochat

# Build the installer
cd cmd/installer
go build -o installer

# Run the installer (optionally add --dev flag)
./installer
# or
./installer --dev
```

**Using the Installer:**

Follow the on-screen prompts to:

1.  Select installation target (Docker Compose or Kubernetes).
2.  Provide necessary configuration (passwords, domain name for Kubernetes Ingress, TLS secret etc.).
3.  Choose the Kubernetes context (if applicable).

The installer performs prerequisite checks (git, docker, helm, kubectl) and handles cloning the repository (using the default or `dev` branch based on the flag) and running the appropriate commands (`docker compose up` or `helm upgrade --install`).

**Flags:**

*   `--dev`: If provided, the installer will clone and deploy using the `dev` branch of the repository instead of the default branch.

### Manual Setup / Development

Before starting the backend manually, you need to prepare the environment.

#### ScyllaDB

*(Add any necessary manual ScyllaDB setup steps here, e.g., creating keyspaces or tables if not handled by the application startup)*

## Usage

*   **Docker Compose:** Access via `http://localhost` (or the relevant Nginx port).
*   **Kubernetes:** Access depends on the Nginx service type chosen (LoadBalancer IP, NodePort, or Ingress setup). Check `kubectl get svc -n <namespace>` for details. Grafana and MinIO consoles might also be exposed depending on service types.