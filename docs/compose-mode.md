# Compose Mode

swarmctl supports Docker Compose as an alternative to Docker Swarm for deployments. This is useful when:

- Docker Swarm is not available or not desired
- You want simpler single-node deployments
- You're developing locally before deploying to Swarm

## Configuration

Enable compose mode by setting `mode: compose` in your swarm.yaml:

```yaml
stack: myapp
mode: compose
compose_file: docker-compose.yaml

# SSH is optional - omit for local deployment
ssh:
  host: server.example.com
  user: deploy
```

## How It Works

In compose mode, swarmctl uses `docker compose` (v2 plugin) instead of Docker Swarm:

| Operation | Swarm Mode | Compose Mode |
|-----------|------------|--------------|
| Deploy | `docker stack deploy` | `docker compose up -d` |
| Remove | `docker stack rm` | `docker compose down` |
| Status | `docker service ls` | `docker compose ps` |
| Logs | `docker service logs` | `docker compose logs` |
| Exec | Find task container | Find compose container |

## Commands

All swarmctl commands work in compose mode:

```bash
swarmctl setup      # Verifies Docker and compose plugin
swarmctl deploy     # Deploys with docker compose
swarmctl status     # Shows containers instead of tasks
swarmctl logs       # Uses docker compose logs
swarmctl exec       # Executes in compose container
swarmctl rollback   # Rolls back via history (see below)
```

## Rollback Support

Compose mode supports rollback through a history sidecar container that stores previous deployments.

### How Rollback Works

1. On each deploy, the compose file and image metadata are recorded
2. The history is stored in a SQLite database inside a container
3. Rollback retrieves the previous compose file and redeploys

### History Container

The history container (`{stack}-history`) is automatically started during deploy. It uses the image `docker.io/marcelsud/swarmctl-history`.

If the history container is not available, rollback will show a warning but deploy will continue.

```bash
# Rollback to previous version
swarmctl rollback
```

**Note:** In compose mode, rollback affects all services at once (individual service rollback is not supported).

## Limitations

### No Dynamic Scaling

Compose mode does not support the `scale` operation:

```bash
swarmctl scale web=3  # Error: scale not supported in compose mode
```

To scale in compose mode, update your docker-compose.yaml with `deploy.replicas` and redeploy.

### All-or-Nothing Rollback

Unlike Swarm mode where individual services can be rolled back, compose mode rolls back the entire deployment at once.

## Setup Requirements

The `swarmctl setup` command in compose mode verifies:

1. Docker is installed
2. Docker Compose v2 plugin is available (`docker compose version`)
3. Registry login (if configured)

It does **not** initialize Swarm or create overlay networks.

## Example

### Local Development

```yaml
# swarm.yaml
stack: myapp
mode: compose
compose_file: docker-compose.yaml
```

```bash
swarmctl setup   # Verify docker compose
swarmctl deploy  # Deploy locally
swarmctl status  # Check status
swarmctl logs web -f  # Follow logs
```

### Remote Deployment

```yaml
# swarm.yaml
stack: myapp
mode: compose
compose_file: docker-compose.yaml

ssh:
  host: server.example.com
  user: deploy
  port: 22
```

```bash
swarmctl setup   # Verify remote docker compose
swarmctl deploy  # Deploy to remote server
swarmctl status  # Check remote status
```

## Switching Between Modes

You can switch between swarm and compose modes by changing the `mode` field:

```yaml
# swarm.yaml for compose
mode: compose
```

```yaml
# swarm.yaml for swarm (default)
mode: swarm
# or simply omit the mode field
```

**Note:** Switching modes requires removing the existing deployment first, as Swarm stacks and Compose projects are managed separately.
