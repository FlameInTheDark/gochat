# GoChat

GoChat is a real-time chat application built with Go.

## Installation

There are two primary ways to install and run GoChat:

1.  **Universal Installer (Recommended):** A command-line tool that guides you through installing GoChat on either Kubernetes (using Helm) or locally with Docker Compose.
2.  **Manual Setup:** For development or custom environments.

### Universal Installer (Quick Start)

This interactive TUI installer simplifies the setup process.

**Prerequisites:**

*   **Always Required:** `git`
*   **For Docker Compose Target:** `docker` and `docker compose` (or `docker-compose`)
*   **For Kubernetes Target:** `helm` and `kubectl`, plus access to a Kubernetes cluster.

**One-Liner (Linux/macOS - amd64):**

This command downloads the latest installer for Linux (amd64), makes it executable, and runs it:

```bash
curl -L -o gochat-installer https://github.com/FlameInTheDark/gochat/releases/latest/download/installer-linux-amd64 && \
chmod +x gochat-installer && \
./gochat-installer
```

*(Note: You might need `sudo` before `./gochat-installer` depending on your download location and permissions.)*

**Other Platforms:**

Pre-built binaries for other platforms (Windows, macOS arm64, etc.) may be available on the [**GitHub Releases**](https://github.com/FlameInTheDark/gochat/releases) page. Download the appropriate binary, make it executable (if necessary), and run it.

**Using the Installer:**

The installer will:

1.  Check for necessary prerequisites based on your chosen target (Kubernetes or Docker Compose).
2.  Ask you to select the installation target.
3.  If Kubernetes is chosen, guide you through configuration (namespace, passwords, service types, cluster context).
4.  Clone the GoChat repository to a temporary location.
5.  Run either `helm upgrade --install ...` (for Kubernetes) or `docker compose up -d --build` (for Docker).
6.  Display the results.

### Manual Setup / Development

Before starting the backend manually, you need to prepare the environment.

#### ScyllaDB

*(Add any necessary manual ScyllaDB setup steps here, e.g., creating keyspaces or tables if not handled by the application startup)*

## Usage

*   **Docker Compose:** Access via `http://localhost` (or the relevant Nginx port).
*   **Kubernetes:** Access depends on the Nginx service type chosen (LoadBalancer IP, NodePort, or Ingress setup). Check `kubectl get svc -n <namespace>` for details. Grafana and MinIO consoles might also be exposed depending on service types.