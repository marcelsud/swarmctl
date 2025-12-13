# Configuração

Referência completa de configuração do swarmctl.

## Arquivos de Configuração

| Arquivo | Descrição |
|---------|-----------|
| `swarm.yaml` | Configuração principal do swarmctl |
| `docker-compose.yaml` | Definição dos serviços (formato Docker padrão) |
| `.env` | Valores dos secrets (não commitar!) |

## swarm.yaml

Arquivo principal de configuração do swarmctl.

```yaml
# Nome do stack no Docker Swarm
stack: myapp-production

# Conexão SSH ao manager node
ssh:
  host: manager.example.com    # Hostname ou IP
  user: deploy                 # Usuário SSH
  port: 22                     # Porta SSH (default: 22)
  key: ~/.ssh/id_ed25519       # Chave privada (opcional, usa ssh-agent por padrão)

# Registry de containers (opcional)
registry:
  url: ghcr.io                 # URL do registry
  username: myuser             # Usuário
  # Password via variável de ambiente SWARMCTL_REGISTRY_PASSWORD

# Secrets a serem criados no Swarm
secrets:
  - DATABASE_URL
  - API_KEY
  - REDIS_URL

# Serviços auxiliares (podem ser parados/iniciados independentemente)
accessories:
  - redis
  - postgres
  - elasticsearch

# Caminho para o docker-compose.yaml
compose_file: docker-compose.yaml
```

## Campos

### stack (obrigatório)

Nome do stack no Docker Swarm. Será usado como prefixo para todos os serviços e secrets.

```yaml
stack: myapp-production
```

Resultado: serviços serão `myapp-production_web`, `myapp-production_api`, etc.

### ssh (obrigatório)

Configuração de conexão SSH ao manager node do Swarm.

| Campo | Obrigatório | Default | Descrição |
|-------|-------------|---------|-----------|
| host | Sim | - | Hostname ou IP do servidor |
| user | Sim | - | Usuário SSH |
| port | Não | 22 | Porta SSH |
| key | Não | - | Caminho para chave privada |

**Autenticação SSH:**

1. **ssh-agent** (recomendado): Se `key` não for especificado, usa o ssh-agent
2. **Chave privada**: Especifique o caminho em `key`
3. **Chaves padrão**: Tenta `~/.ssh/id_ed25519` e `~/.ssh/id_rsa`

### registry (opcional)

Configuração do registry de containers para pull de imagens privadas.

```yaml
registry:
  url: ghcr.io
  username: myuser
```

**Password**: Defina a variável de ambiente `SWARMCTL_REGISTRY_PASSWORD`.

```bash
export SWARMCTL_REGISTRY_PASSWORD=ghp_xxxxxxxxxxxx
swarmctl deploy
```

### secrets (opcional)

Lista de secrets a serem criados no Docker Swarm. Os valores são lidos de:

1. Arquivo `.env` no diretório atual
2. Variáveis de ambiente

```yaml
secrets:
  - DATABASE_URL
  - API_KEY
```

**.env:**
```
DATABASE_URL=postgres://user:pass@localhost/db
API_KEY=secret-key-123
```

Os secrets são criados com o formato `{stack}_{secret_name}`:
- `myapp_database_url`
- `myapp_api_key`

### accessories (opcional)

Lista de serviços que podem ser gerenciados independentemente (start/stop/restart).

```yaml
accessories:
  - redis
  - postgres
```

Útil para:
- Parar banco de dados para manutenção
- Reiniciar cache
- Debugging

### compose_file (opcional)

Caminho para o arquivo docker-compose.yaml. Default: `docker-compose.yaml`.

```yaml
compose_file: docker/production.yaml
```

## docker-compose.yaml

Use o formato padrão do Docker Compose com a seção `deploy` para configurações do Swarm.

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
      restart_policy:
        condition: on-failure
    ports:
      - "80:3000"
    environment:
      - RAILS_ENV=production
    secrets:
      - myapp_database_url

  worker:
    image: myapp:latest
    command: bundle exec sidekiq
    deploy:
      replicas: 2

  redis:
    image: redis:7-alpine
    deploy:
      replicas: 1

secrets:
  myapp_database_url:
    external: true
```

## Variáveis de Ambiente

| Variável | Descrição |
|----------|-----------|
| `SWARMCTL_REGISTRY_PASSWORD` | Password do registry de containers |

## Multi-ambiente

Veja [Multi-ambiente](./multi-environment.md) para configurar staging/production.
