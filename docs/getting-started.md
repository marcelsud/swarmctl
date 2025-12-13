# Getting Started

Guia de instalação e primeiro deploy com swarmctl.

## Pré-requisitos

O swarmctl pode rodar em dois modos: **local** ou **remoto (SSH)**.

### Modo Local

- Docker 24.0+ instalado
- Docker Swarm inicializado (`docker swarm init`)
- Usuário com permissão para executar Docker

### Modo Remoto (SSH)

**Na sua máquina local:**
- Go 1.22+ (para build)
- SSH client configurado
- Chave SSH (recomendado: ed25519)

**No servidor de destino:**
- Docker 24.0+
- Porta 22 (SSH) acessível
- Usuário com permissão para executar Docker

## Instalação

### Via Go Install

```bash
go install github.com/marcelsud/swarmctl/cmd/swarmctl@latest
```

### Build from Source

```bash
git clone https://github.com/marcelsud/swarmctl.git
cd swarmctl
go build -o swarmctl ./cmd/swarmctl
sudo mv swarmctl /usr/local/bin/
```

### Verificar instalação

```bash
swarmctl --version
```

## Primeiro Deploy (Modo Local)

Ideal para desenvolvimento ou quando você está no próprio manager node.

### 1. Inicializar Swarm

```bash
docker swarm init
```

### 2. Criar configuração

No diretório do seu projeto, crie `swarm.yaml`:

```yaml
stack: myapp

compose_file: docker-compose.yaml
# Sem seção ssh = modo local
```

### 3. Criar docker-compose.yaml

```yaml
version: "3.8"

services:
  web:
    image: nginx:alpine
    deploy:
      replicas: 2
    ports:
      - "80:80"
```

### 4. Deploy

```bash
swarmctl setup
swarmctl deploy
swarmctl status
```

---

## Primeiro Deploy (Modo Remoto via SSH)

Para deploy em servidores remotos.

### 1. Preparar o servidor

Certifique-se de que o servidor tem Docker instalado:

```bash
ssh deploy@seu-servidor.com "docker --version"
```

Se não tiver, instale o Docker:

```bash
ssh deploy@seu-servidor.com
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER
# Reconecte para aplicar o grupo
```

### 2. Criar configuração

No diretório do seu projeto, crie `swarm.yaml`:

```yaml
stack: myapp

ssh:
  host: seu-servidor.com
  user: deploy
  port: 22

compose_file: docker-compose.yaml
```

### 3. Criar docker-compose.yaml

```yaml
version: "3.8"

services:
  web:
    image: nginx:alpine
    deploy:
      replicas: 2
      update_config:
        parallelism: 1
        delay: 5s
        failure_action: rollback
    ports:
      - "80:80"
```

### 4. Setup do cluster

```bash
swarmctl setup
```

Este comando:
- Conecta via SSH ao servidor
- Verifica se Docker está instalado
- Inicializa o Swarm (se necessário)
- Cria a network overlay para o stack
- Faz login no registry (se configurado)

**Output esperado:**
```
→ Loading configuration...
  Stack: myapp
  Host:  deploy@seu-servidor.com:22
→ Connecting to seu-servidor.com...
  ✓ Connected
→ Checking Docker installation...
  ✓ Docker version 24.0.7
→ Checking Swarm status...
  ✓ Swarm already initialized
→ Creating network myapp-network...
  ✓ Network ready

✓ Setup complete! Run 'swarmctl deploy' to deploy your stack.
```

### 5. Deploy

```bash
swarmctl deploy
```

**Output esperado:**
```
→ Loading configuration...
  Stack: myapp
→ Connecting to seu-servidor.com...
  ✓ Connected
→ Deploying stack myapp...
  ✓ Stack deployed
→ Waiting for services to start...

→ Services:
  NAME                  MODE         REPLICAS
  myapp_web             replicated   2/2

✓ Deploy completed
```

### 6. Verificar status

```bash
swarmctl status
```

### 7. Ver logs

```bash
swarmctl logs web
swarmctl logs web -f    # Follow mode
```

## Estrutura de Arquivos Recomendada

```
myapp/
├── swarm.yaml              # Configuração do swarmctl
├── swarm.staging.yaml      # Config para staging (opcional)
├── swarm.production.yaml   # Config para production (opcional)
├── docker-compose.yaml     # Definição dos serviços
├── .env                    # Secrets locais (não commitar!)
├── .env.staging            # Secrets staging
└── .env.production         # Secrets production
```

## Configuração de SSH

### Usando ssh-agent (recomendado)

```bash
# Adicionar chave ao agent
ssh-add ~/.ssh/id_ed25519

# Verificar
ssh-add -l

# Deploy
swarmctl deploy
```

### Usando chave específica

```yaml
# swarm.yaml
ssh:
  host: seu-servidor.com
  user: deploy
  key: ~/.ssh/id_ed25519_deploy
```

### Testando conexão

```bash
ssh -i ~/.ssh/id_ed25519 deploy@seu-servidor.com "docker info"
```

## Próximos Passos

- [Configuração](./configuration.md) - Referência completa do swarm.yaml
- [Modo Local](./local-mode.md) - Executar sem SSH
- [Comandos](./commands.md) - Todos os comandos disponíveis
- [Multi-ambiente](./multi-environment.md) - Configurar staging/production
