# Modo Local

Executar swarmctl diretamente na máquina local, sem SSH.

## Quando usar

O modo local é útil quando:

- Você está rodando Docker Swarm na própria máquina de desenvolvimento
- O swarmctl está instalado diretamente no manager node
- Você quer testar stacks localmente antes de fazer deploy remoto
- Ambientes de CI/CD onde o runner já tem acesso ao Docker

## Como ativar

Para usar o modo local, simplesmente **omita a seção `ssh`** do `swarm.yaml`:

```yaml
# swarm.yaml - Modo Local
stack: myapp

compose_file: docker-compose.yaml

# Sem seção ssh = execução local
```

Comparado com o modo remoto:

```yaml
# swarm.yaml - Modo Remoto (via SSH)
stack: myapp

ssh:
  host: manager.example.com
  user: deploy
  port: 22

compose_file: docker-compose.yaml
```

## Pré-requisitos

### Modo Local

- Docker instalado e rodando
- Docker Swarm inicializado (`docker swarm init`)
- Usuário com permissão para executar comandos Docker

### Modo Remoto (SSH)

- SSH client configurado
- Chave SSH ou ssh-agent
- Docker instalado no servidor remoto
- Usuário SSH com permissão Docker

## Exemplos

### Deploy local

```yaml
# swarm.yaml
stack: myapp-dev

compose_file: docker-compose.yaml

secrets:
  - DATABASE_URL
  - API_KEY
```

```bash
# Inicializar Swarm (primeira vez)
docker swarm init

# Setup e deploy
swarmctl setup
swarmctl deploy

# Status
swarmctl status
```

### Desenvolvimento local + Deploy remoto

Use arquivos de configuração diferentes para cada ambiente:

```yaml
# swarm.yaml - Desenvolvimento local
stack: myapp-dev
compose_file: docker-compose.yaml
```

```yaml
# swarm.production.yaml - Produção remota
stack: myapp-production

ssh:
  host: prod.example.com
  user: deploy
  port: 22

compose_file: docker-compose.yaml

registry:
  url: ghcr.io
  username: myuser
```

```bash
# Deploy local
swarmctl deploy

# Deploy produção
swarmctl deploy -d production
```

## Comportamento

O modo local executa os mesmos comandos Docker que o modo remoto, apenas diretamente via shell ao invés de SSH:

| Comando | Modo Local | Modo Remoto |
|---------|------------|-------------|
| `swarmctl setup` | `docker swarm init` local | SSH + `docker swarm init` |
| `swarmctl deploy` | `docker stack deploy` local | SSH + `docker stack deploy` |
| `swarmctl status` | `docker service ls` local | SSH + `docker service ls` |
| `swarmctl logs` | `docker service logs` local | SSH + `docker service logs` |
| `swarmctl exec` | `docker exec` local | SSH + `docker exec` |

## Detecção automática

O swarmctl detecta automaticamente o modo baseado na presença da configuração SSH:

```
swarm.yaml
    │
    ├── ssh.host definido?
    │       │
    │       ├── Sim → Modo Remoto (SSH)
    │       │         Conecta via SSH ao host
    │       │
    │       └── Não → Modo Local
    │                 Executa comandos diretamente
    │
```

## Verificando o modo

O output do swarmctl indica qual modo está sendo usado:

**Modo Local:**
```
→ Loading configuration...
  Stack: myapp
  Mode:  local
→ Checking Docker installation...
  ✓ Docker version 24.0.7
```

**Modo Remoto:**
```
→ Loading configuration...
  Stack: myapp
  Host:  deploy@manager.example.com:22
→ Connecting to manager.example.com...
  ✓ Connected
→ Checking Docker installation...
  ✓ Docker version 24.0.7
```

## Limitações

- **Sem ssh-agent**: No modo local, não há conexão SSH
- **Mesma máquina**: Todos os comandos rodam na máquina onde o swarmctl está
- **Permissões**: O usuário atual precisa ter acesso ao Docker

## Migração

Para migrar de local para remoto, adicione a seção `ssh`:

```yaml
# Antes (local)
stack: myapp
compose_file: docker-compose.yaml

# Depois (remoto)
stack: myapp

ssh:
  host: manager.example.com
  user: deploy

compose_file: docker-compose.yaml
```
