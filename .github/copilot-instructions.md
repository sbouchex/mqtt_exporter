# Copilot Coding Agent Onboarding Instructions

## Repository Summary
This repository implements an MQTT Exporter written in Go. It receives metrics in various formats (JSON, collectd, raw) via MQTT, filters and transforms them, and exposes them for Prometheus consumption. Typical use cases include integrating metrics from sources like zigbee2mqtt or collectd, with flexible filtering and mapping to Prometheus metrics.

## High-Level Information
- **Project Type:** Go application, containerized for deployment
- **Languages:** Go (primary)
- **Frameworks/Libraries:**
  - MQTT: github.com/eclipse/paho.mqtt.golang
  - Prometheus: github.com/prometheus/client_golang
  - Logging: github.com/sirupsen/logrus
  - Config: github.com/spf13/viper, github.com/spf13/pflag
- **Repo Size:** Small, single main Go file (`main.go`), config samples, Dockerfile
- **Target Runtime:** Go 1.25, runs as a container or standalone binary

## Build, Run, and Validation Steps
### Environment Setup
- **Go Version:** Always use Go 1.25 (see `go.mod` and Dockerfile)
- **Dependencies:** All dependencies are managed via Go modules. No manual install required if using `go build` or Docker.

### Build
- **Container Build:**
  - `docker build -t mqtt_exporter .`
  - This builds the binary and container image. Always use this for production.
- **Local Build:**
  - `go build -o mqtt_exporter main.go`
  - Produces the binary for local testing.

### Run
- **Container Run:**
  - `docker run -d -p 9393:9393 --name=mqtt_exporter --network <network> --restart=always -v mqtt_exporter:/mqtt_exporter_data mqtt_exporter:latest /mqtt_exporter`
  - Always mount configuration files and set up the correct network for MQTT broker access.
- **Local Run:**
  - `./mqtt_exporter -c mqtt_exporter.json` (or use default config path)

### Configuration
- **Main config:** `mqtt_exporter.json` (MQTT broker, listen address, config file path)
- **Mapping config:** `configuration.json` (sensor/topic mapping, filters, labels)
- **Sample configs:** See `*.sample` files for templates.

### Validation
- **Prometheus Endpoint:** After starting, metrics are available at `/metrics` (default port 9393).
- **Check logs:** Verbose mode with `-v` flag for debugging.
- **CI/CD:**
  - GitHub Actions build on push/PR to `main` (`.github/workflows/build.yml`).
  - Container images are built and pushed to GHCR.
  - Old images are purged by scheduled workflow (`purge.yml`).

### Common Issues & Workarounds
- **Go version mismatch:** Always use Go 1.25. Other versions may fail to build.
- **Config file errors:** Ensure config files are present and valid JSON. Missing or invalid config will cause startup failure.
- **MQTT broker access:** Container must be able to reach the broker. Check network settings.
- **Port conflicts:** Default port is 9393. Change in config if needed.

## Project Layout & Key Files
- **Root files:**
  - `main.go`: All source code (entry point, MQTT/Prometheus logic)
  - `Dockerfile`: Build and run instructions for container
  - `go.mod`, `go.sum`: Dependency management
  - `README.md`: Usage, config, and build instructions
  - `mqtt_exporter.json`, `configuration.json`: Main and mapping configs
  - `*.sample` files: Config templates
  - `.editorconfig`, `.gitignore`, `LICENSE`
- **GitHub Actions:** `.github/workflows/build.yml` (build/push), `purge.yml` (cleanup)

## Architecture Overview
- **Single binary:** All logic in `main.go`
- **Config-driven:** Behavior controlled by JSON config files
- **Metrics exposed via HTTP for Prometheus**
- **MQTT client subscribes to topics, parses messages, and updates metrics**

## Validation Steps Before Check-In
- Always build and run the container locally before submitting changes.
- Validate config files with sample data.
- Check Prometheus endpoint for expected metrics.
- Review CI build status after push/PR.

## Agent Guidance
- Trust these instructions for build, run, and validation steps.
- Only search the codebase if information here is incomplete or found to be in error.
- For changes, prefer updating `main.go` and config files in the repo root.
- For new features, follow the config-driven architecture and update documentation as needed.
