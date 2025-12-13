# swarmctl

CLI para deploy e gerenciamento de stacks Docker Swarm, inspirada no Kamal. Suporta execução local ou remota via SSH.

## Documentação

- [Getting Started](docs/getting-started.md) - Instalação e primeiro deploy
- [Configuração](docs/configuration.md) - Referência do swarm.yaml
- [Modo Local](docs/local-mode.md) - Executar sem SSH
- [Comandos](docs/commands.md) - Todos os comandos disponíveis
- [Multi-ambiente](docs/multi-environment.md) - Staging/Production

## Comandos Rápidos

```bash
swarmctl setup              # Inicializa Swarm no servidor
swarmctl deploy             # Deploy do stack
swarmctl status             # Status dos serviços
swarmctl logs <service>     # Ver logs (-f para follow)
swarmctl rollback           # Rollback para versão anterior
swarmctl exec <service>     # Shell no container
swarmctl secrets push       # Push secrets do .env
swarmctl secrets list       # Lista secrets
swarmctl accessory          # Lista accessories
swarmctl accessory start    # Inicia accessory
swarmctl accessory stop     # Para accessory
```

## Multi-ambiente

```bash
swarmctl deploy -d staging      # Usa swarm.staging.yaml
swarmctl deploy -d production   # Usa swarm.production.yaml
```

## Estrutura do Projeto

```
swarmctl/
├── cmd/swarmctl/
│   └── main.go                 # Entry point
├── internal/
│   ├── accessories/
│   │   ├── manager.go          # Start/Stop/Restart accessories
│   │   └── manager_test.go
│   ├── config/
│   │   ├── config.go           # Structs de configuração
│   │   ├── config_test.go
│   │   ├── parser.go           # Load swarm.yaml
│   │   ├── validation.go       # Validação de campos
│   │   └── validation_test.go
│   ├── executor/
│   │   ├── executor.go         # Interface de execução
│   │   ├── local.go            # Execução local (os/exec)
│   │   └── ssh.go              # Execução remota (SSH)
│   ├── secrets/
│   │   ├── manager.go          # Push/List secrets
│   │   └── manager_test.go
│   ├── ssh/
│   │   ├── client.go           # Conexão SSH (ssh-agent, keys)
│   │   └── session.go          # Execução de comandos
│   └── swarm/
│       ├── service.go          # Service operations
│       ├── stack.go            # Deploy/Remove stack
│       ├── status.go           # Status queries
│       └── swarm.go            # Init, network, registry
├── pkg/cli/
│   ├── root.go                 # Flags globais (-c, -d)
│   ├── setup.go                # swarmctl setup
│   ├── deploy.go               # swarmctl deploy
│   ├── status.go               # swarmctl status
│   ├── logs.go                 # swarmctl logs
│   ├── rollback.go             # swarmctl rollback
│   ├── exec.go                 # swarmctl exec
│   ├── secrets.go              # swarmctl secrets
│   └── accessory.go            # swarmctl accessory
├── docs/                       # Documentação
├── go.mod
└── go.sum
```

## Arquivos de Configuração

| Arquivo | Descrição |
|---------|-----------|
| `swarm.yaml` | Configuração principal (SSH, registry, secrets, accessories) |
| `docker-compose.yaml` | Definição dos serviços (formato padrão Docker) |
| `.env` | Valores dos secrets (não commitar) |

## Build e Testes

```bash
# Build
go build -o swarmctl ./cmd/swarmctl

# Testes
go test ./...
go test ./... -v        # Verbose
go test ./... -cover    # Com cobertura

# Vet
go vet ./...
```

## Fluxo de Deploy

```
swarmctl deploy
     │
     ▼
┌─────────────────┐
│ Load config     │  ← swarm.yaml + docker-compose.yaml
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ ssh.host set?   │
└────────┬────────┘
         │
    ┌────┴────┐
    │         │
   Sim       Não
    │         │
    ▼         ▼
┌────────┐ ┌────────┐
│  SSH   │ │ Local  │  ← Modo de execução
│connect │ │executor│
└───┬────┘ └───┬────┘
    │          │
    └────┬─────┘
         │
         ▼
┌─────────────────┐
│ Registry login  │  ← SWARMCTL_REGISTRY_PASSWORD (opcional)
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ docker stack    │  ← docker stack deploy -c compose.yml
│ deploy          │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Show status     │  ← docker service ls
└─────────────────┘
```

## Dependências

```
github.com/spf13/cobra      # CLI framework
github.com/fatih/color      # Terminal colors
golang.org/x/crypto/ssh     # SSH client
gopkg.in/yaml.v3            # YAML parsing
```

## Variáveis de Ambiente

| Variável | Descrição |
|----------|-----------|
| `SWARMCTL_REGISTRY_PASSWORD` | Password do registry de containers |
