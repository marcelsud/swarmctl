# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- `init-llm` command for creating LLM/AI assistant configuration files (CLAUDE.md, .cursorrules, AGENTS.md, copilot-instructions.md, Claude skill)
- `exec` support for containers running on worker nodes with automatic SSH hop through manager
- Compose mode support for deploying with docker-compose
- Local mode support (no SSH required) for running swarmctl locally
- `docs` command with embedded documentation
- Entry point for `go install` support

### Fixed

- Include cmd/swarmctl entry point for go install

## [0.1.0] - Initial Release

### Added

- Core CLI structure with Cobra framework
- `setup` command to initialize Docker Swarm on server
- `deploy` command to deploy stacks
- `status` command to check service status
- `logs` command to view service logs
- `rollback` command to rollback to previous version
- `exec` command to shell into containers
- `secrets` commands (push, list) to manage Docker secrets
- `accessory` commands (start, stop, restart) to manage accessory services
- SSH-based remote execution
- Multi-environment support (-d staging, -d production)
- Configuration via swarm.yaml and docker-compose.yaml
