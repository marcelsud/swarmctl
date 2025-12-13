# Multi-ambiente

Guia para configurar e gerenciar múltiplos ambientes (staging, production, etc).

O swarmctl suporta múltiplos ambientes através da flag `-d` (destination).

## Estrutura de Arquivos

```
myapp/
├── swarm.yaml              # Default / desenvolvimento
├── swarm.staging.yaml      # Staging
├── swarm.production.yaml   # Production
├── docker-compose.yaml     # Compose compartilhado
├── .env                    # Secrets locais
├── .env.staging            # Secrets staging
└── .env.production         # Secrets production
```

## Configuração por Ambiente

### swarm.yaml (default)

```yaml
stack: myapp-dev

ssh:
  host: dev.example.com
  user: deploy

compose_file: docker-compose.yaml
```

### swarm.staging.yaml

```yaml
stack: myapp-staging

ssh:
  host: staging.example.com
  user: deploy

registry:
  url: ghcr.io
  username: myuser

secrets:
  - DATABASE_URL
  - API_KEY

compose_file: docker-compose.yaml
```

### swarm.production.yaml

```yaml
stack: myapp-production

ssh:
  host: prod.example.com
  user: deploy

registry:
  url: ghcr.io
  username: myuser

secrets:
  - DATABASE_URL
  - API_KEY
  - SENTRY_DSN

accessories:
  - redis
  - elasticsearch

compose_file: docker-compose.yaml
```

## Uso

```bash
# Default (usa swarm.yaml)
swarmctl setup
swarmctl deploy

# Staging (usa swarm.staging.yaml)
swarmctl setup -d staging
swarmctl deploy -d staging
swarmctl status -d staging
swarmctl logs web -d staging

# Production (usa swarm.production.yaml)
swarmctl setup -d production
swarmctl deploy -d production
swarmctl status -d production
```

## Secrets por Ambiente

```bash
# Staging
swarmctl secrets push -d staging -e .env.staging

# Production
swarmctl secrets push -d production -e .env.production
```

## Docker Compose por Ambiente

Se precisar de compose files diferentes por ambiente:

### swarm.staging.yaml
```yaml
compose_file: docker-compose.staging.yaml
```

### swarm.production.yaml
```yaml
compose_file: docker-compose.production.yaml
```

## Aliases (opcional)

Adicione aliases no seu `.bashrc` ou `.zshrc`:

```bash
alias deploy-staging="swarmctl deploy -d staging"
alias deploy-prod="swarmctl deploy -d production"
alias logs-staging="swarmctl logs -d staging"
alias logs-prod="swarmctl logs -d production"
```

## CI/CD

### GitHub Actions

```yaml
name: Deploy

on:
  push:
    branches: [main]

jobs:
  deploy-staging:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Install swarmctl
        run: go install github.com/marcelsud/swarmctl/cmd/swarmctl@latest

      - name: Setup SSH
        run: |
          mkdir -p ~/.ssh
          echo "${{ secrets.SSH_KEY }}" > ~/.ssh/id_ed25519
          chmod 600 ~/.ssh/id_ed25519
          ssh-keyscan staging.example.com >> ~/.ssh/known_hosts

      - name: Deploy to staging
        env:
          SWARMCTL_REGISTRY_PASSWORD: ${{ secrets.GHCR_TOKEN }}
        run: swarmctl deploy -d staging

  deploy-production:
    runs-on: ubuntu-latest
    needs: deploy-staging
    if: github.ref == 'refs/heads/main'
    steps:
      # Similar ao staging...
      - name: Deploy to production
        env:
          SWARMCTL_REGISTRY_PASSWORD: ${{ secrets.GHCR_TOKEN }}
        run: swarmctl deploy -d production
```
