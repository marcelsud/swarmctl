# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Security

- **CRITICAL**: Fix command injection vulnerability in accessories manager
  - Add strict input validation for accessory names (alphanumeric + underscore only)
  - Fix incorrect shellquote usage that could allow shell metacharacter injection
  - Affects: `swarmctl accessory start/stop/restart` commands
  - Impact: Remote code execution via malicious accessory names
  
- **CRITICAL**: Fix SSH host key verification bypass
  - Replace `InsecureIgnoreHostKey()` with proper host key validation
  - Use system's `known_hosts` file for verification
  - Add fallback confirmation prompt for unknown hosts
  - Add `SWARMCTL_INSECURE_SSH` escape hatch with explicit warning
  - Impact: MITM attacks against SSH connections
  
- **CRITICAL**: Fix SSH parameter injection in worker node connections
  - Add strict validation for targetHost and targetUser parameters
  - Use proper shell escaping for SSH hop commands
  - Remove insecure SSH options (`StrictHostKeyChecking=no`)
  - Impact: Command injection on Swarm manager via malicious node configuration

### Added

- Comprehensive integration test suite with Multipass and Docker Swarm
- `--verbose` flag for detailed command execution output
- Automatic Docker secrets push during `deploy`
- Real-time health check monitoring during `deploy`
- Support for `deploy --service` to deploy specific services
- Support for `deploy --skip-accessories` to skip auxiliary services
- `logs` command now shows logs for all services when no service is specified
- `init-llm` command for creating LLM/AI assistant configuration files (CLAUDE.md, .cursorrules, AGENTS.md, copilot-instructions.md, Claude skill)
- `exec` support for containers running on worker nodes with automatic SSH hop through manager
- Compose mode support for deploying with docker-compose
- Local mode support (no SSH required) for running swarmctl locally
- `docs` command with embedded documentation
- Entry point for `go install` support
- Comprehensive security tests for all injection vulnerabilities

### Fixed

- Include cmd/swarmctl entry point for go install
- Fix shellquote usage in accessories manager (intermediate variables)

### Breaking Changes

- Accessory names, SSH hosts, and usernames now restricted to alphanumeric characters and underscores only (no dots or hyphens allowed)

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
