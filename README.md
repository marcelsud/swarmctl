# swarmctl

CLI para deploy e gerenciamento de stacks Docker Swarm via SSH, inspirada no [Kamal](https://kamal-deploy.org/).

## Features

- **Deploy simplificado** - Deploy de stacks Docker Swarm com um único comando
- **Multi-ambiente** - Suporte a staging, production e outros ambientes via `-d` flag
- **Gerenciamento de secrets** - Push de secrets do `.env` para Docker Swarm
- **Accessories** - Gerenciamento independente de serviços auxiliares (Redis, Postgres, etc)
- **Rollback** - Rollback rápido para versão anterior
- **Logs agregados** - Visualização de logs em tempo real
- **Exec remoto** - Shell interativo em containers
- **SSH nativo** - Conexão via SSH com suporte a ssh-agent

## Requisitos

- Go 1.22+ (para build)
- Docker Swarm configurado no servidor de destino
- Acesso SSH ao manager node

## Instalação

```bash
go install github.com/marcelsud/swarmctl/cmd/swarmctl@latest
```

Ou build local:

```bash
git clone https://github.com/marcelsud/swarmctl.git
cd swarmctl
go build -o swarmctl ./cmd/swarmctl
sudo mv swarmctl /usr/local/bin/
```

## Quick Start

1. Crie `swarm.yaml`:

```yaml
stack: myapp

ssh:
  host: manager.example.com
  user: deploy
  port: 22

secrets:
  - DATABASE_URL
  - API_KEY

accessories:
  - redis
  - postgres

compose_file: docker-compose.yaml
```

2. Crie `docker-compose.yaml`:

```yaml
version: "3.8"

services:
  web:
    image: myapp:latest
    deploy:
      replicas: 3
      update_config:
        parallelism: 1
        delay: 10s
        failure_action: rollback
    ports:
      - "80:3000"

  redis:
    image: redis:7-alpine

  postgres:
    image: postgres:16-alpine
```

3. Deploy:

```bash
swarmctl setup    # Inicializa o cluster Swarm
swarmctl deploy   # Faz deploy do stack
```

## Comandos

| Comando | Descrição |
|---------|-----------|
| `swarmctl setup` | Inicializa Swarm, cria network, login no registry |
| `swarmctl deploy` | Faz deploy do stack no Swarm |
| `swarmctl status` | Mostra status do stack e serviços |
| `swarmctl logs <service>` | Visualiza logs do serviço |
| `swarmctl rollback [service]` | Rollback para versão anterior |
| `swarmctl exec <service> [cmd]` | Executa comando no container |
| `swarmctl secrets push` | Envia secrets do .env para o Swarm |
| `swarmctl secrets list` | Lista secrets existentes |
| `swarmctl accessory` | Lista status dos accessories |
| `swarmctl accessory start <name>` | Inicia accessory |
| `swarmctl accessory stop <name>` | Para accessory |
| `swarmctl accessory restart <name>` | Reinicia accessory |

## Multi-ambiente

Use a flag `-d` para diferentes ambientes:

```bash
swarmctl deploy                  # Usa swarm.yaml
swarmctl deploy -d staging       # Usa swarm.staging.yaml
swarmctl deploy -d production    # Usa swarm.production.yaml
```

## Estrutura de Arquivos

```
myapp/
├── swarm.yaml              # Configuração principal
├── swarm.staging.yaml      # Config para staging
├── swarm.production.yaml   # Config para production
├── docker-compose.yaml     # Definição dos serviços
├── .env                    # Secrets locais (não commitar!)
├── .env.staging            # Secrets staging
└── .env.production         # Secrets production
```

## Documentação

- [Getting Started](docs/getting-started.md) - Instalação e primeiro deploy
- [Configuração](docs/configuration.md) - Referência completa do swarm.yaml
- [Comandos](docs/commands.md) - Documentação de todos os comandos
- [Multi-ambiente](docs/multi-environment.md) - Configuração de staging/production

## Desenvolvimento

### Build

```bash
go build -o swarmctl ./cmd/swarmctl
```

### Testes

```bash
go test ./...           # Executar todos os testes
go test ./... -v        # Modo verbose
go test ./... -cover    # Com cobertura
```

### Estrutura do Projeto

```
swarmctl/
├── cmd/swarmctl/           # Entry point
├── internal/
│   ├── accessories/        # Gerenciamento de accessories
│   ├── config/             # Parser de configuração
│   ├── secrets/            # Gerenciamento de secrets
│   ├── ssh/                # Cliente SSH
│   └── swarm/              # Operações Docker Swarm
├── pkg/cli/                # Comandos CLI (Cobra)
├── docs/                   # Documentação
└── e2e-test/               # Arquivos de teste E2E
```

## Configuração SSH

O swarmctl suporta múltiplos métodos de autenticação SSH:

1. **ssh-agent** (recomendado) - Usa chaves do ssh-agent automaticamente
2. **Chave específica** - Configure `ssh.key` no swarm.yaml
3. **Chaves padrão** - Tenta `~/.ssh/id_ed25519` e `~/.ssh/id_rsa`

## Variáveis de Ambiente

| Variável | Descrição |
|----------|-----------|
| `SWARMCTL_REGISTRY_PASSWORD` | Password do registry de containers |

## License

MIT
