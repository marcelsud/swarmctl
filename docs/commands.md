# Comandos

Referência completa de todos os comandos do swarmctl.

## Flags Globais

```bash
-c, --config string        # Arquivo de configuração (default: swarm.yaml)
-d, --destination string   # Ambiente de destino (staging, production)
-v, --verbose              # Output detalhado
    --version              # Versão do swarmctl
```

---

## swarmctl setup

Configura o ambiente de deployment.

```bash
swarmctl setup
swarmctl setup -d production
```

**Ações (Swarm mode):**
1. Conecta via SSH ao manager (se configurado)
2. Verifica se Docker está instalado
3. Inicializa Swarm (`docker swarm init`) se necessário
4. Cria network overlay para o stack
5. Faz login no registry (se configurado)

**Ações (Compose mode):**
1. Conecta via SSH (se configurado)
2. Verifica se Docker está instalado
3. Verifica se docker compose plugin está disponível
4. Faz login no registry (se configurado)

**Output:**
```
→ Loading configuration...
  Stack: myapp
  Host:  deploy@manager.example.com:22
→ Connecting to manager.example.com...
  ✓ Connected
→ Checking Docker installation...
  ✓ Docker version 24.0.7
→ Checking Swarm status...
  ✓ Swarm already initialized
→ Creating network myapp-network...
  ✓ Network ready

→ Swarm nodes:
HOSTNAME     STATUS    AVAILABILITY   MANAGER STATUS
manager      Ready     Active         Leader

✓ Setup complete! Run 'swarmctl deploy' to deploy your stack.
```

---

## swarmctl deploy

Faz deploy do stack.

```bash
swarmctl deploy
swarmctl deploy -d staging
swarmctl deploy --service web        # Deploy apenas do serviço web
swarmctl deploy --skip-accessories   # Não atualiza accessories
```

**Flags:**
```
-s, --service string     # Deploy apenas este serviço
    --skip-accessories   # Não atualiza serviços auxiliares
```

**Ações (Swarm mode):**
1. Carrega e valida configuração
2. Conecta via SSH (se configurado)
3. Login no registry
4. Executa `docker stack deploy`
5. Aguarda serviços iniciarem
6. Mostra status final

**Ações (Compose mode):**
1. Carrega e valida configuração
2. Conecta via SSH (se configurado)
3. Login no registry
4. Executa `docker compose up -d`
5. Registra deploy no histórico (para rollback)
6. Aguarda serviços iniciarem
7. Mostra status final

**Output:**
```
→ Loading configuration...
  Stack: myapp
→ Loading compose file...
  ✓ docker-compose.yaml
→ Connecting to manager.example.com...
  ✓ Connected
→ Deploying stack myapp...
  ✓ Stack deployed
→ Waiting for services to start...

→ Services:
  NAME                  MODE         REPLICAS        IMAGE
  myapp_web             replicated   3/3             myapp:latest
  myapp_worker          replicated   2/2             myapp:latest

✓ Deploy completed in 8.234s
```

---

## swarmctl status

Mostra status do stack e serviços.

```bash
swarmctl status
swarmctl status web    # Status detalhado do serviço web
```

**Output (Swarm mode):**
```
→ Stack: myapp (swarm mode)

→ Services:
  NAME                  MODE         REPLICAS        PORTS
  myapp_web             replicated   3/3             *:80->3000/tcp
  myapp_worker          replicated   2/2

→ Tasks:
  ID              NAME              NODE       STATE              ERROR
  abc123def456    myapp_web.1       node-1     Running 2 hours ago
  def456ghi789    myapp_web.2       node-2     Running 2 hours ago
  ghi789jkl012    myapp_web.3       node-1     Running 2 hours ago
```

**Output (Compose mode):**
```
→ Stack: myapp (compose mode)

→ Services:
  NAME                  MODE         REPLICAS        PORTS
  myapp_web             replicated   1/1             0.0.0.0:80->3000/tcp

→ Containers:
  ID              NAME              SERVICE          STATE
  f74a766bfdde    myapp-web-1       web              running
```

---

## swarmctl logs

Visualiza logs agregados dos serviços.

```bash
swarmctl logs web              # Logs do serviço web
swarmctl logs web -f           # Follow mode (tempo real)
swarmctl logs web --tail 50    # Últimas 50 linhas
swarmctl logs web --since 1h   # Última hora
```

**Flags:**
```
-f, --follow         # Acompanhar logs em tempo real
-n, --tail int       # Número de linhas (default: 100)
    --since string   # Mostrar logs desde (ex: 1h, 30m, 2h30m)
```

---

## swarmctl rollback

Volta serviços para a versão anterior.

```bash
swarmctl rollback           # Rollback de todos os serviços
swarmctl rollback web       # Rollback apenas do web (apenas swarm mode)
```

**Swarm mode:**
- Executa `docker service update --rollback` para cada serviço
- Suporta rollback de serviços individuais
- Mostra status após rollback

**Compose mode:**
- Usa o histórico de deploys armazenado no container sidecar
- Rollback sempre afeta todos os serviços de uma vez
- Requer o container de histórico (`{stack}-history`)

**Output (Swarm mode):**
```
→ Stack: myapp (swarm mode)
→ Rolling back 3 service(s)...
  → web... ✓
  → worker... ✓
  → redis... ✓

→ Services after rollback:
  myapp_web                    3/3
  myapp_worker                 2/2

✓ Rollback completed
```

**Output (Compose mode):**
```
→ Stack: myapp (compose mode)
→ Rolling back to previous deploy...

→ Services after rollback:
  myapp_web                    1/1

✓ Rollback completed
```

---

## swarmctl exec

Executa comando em um container do serviço.

```bash
swarmctl exec web                  # Shell interativo (sh)
swarmctl exec web bash             # Shell bash
swarmctl exec web -- ls -la        # Comando específico
swarmctl exec api -- rails console # Rails console
```

**Output:**
```
→ Finding container for service web...
→ Executing: ls -la

total 64
drwxr-xr-x 1 root root 4096 Dec 12 10:00 .
drwxr-xr-x 1 root root 4096 Dec 12 10:00 ..
-rw-r--r-- 1 root root  220 Dec 12 10:00 Gemfile
...
```

---

## swarmctl secrets

Gerencia secrets do Docker Swarm.

### secrets push

Envia secrets do `.env` ou variáveis de ambiente para o Swarm.

```bash
swarmctl secrets push
swarmctl secrets push -e .env.production
```

**Flags:**
```
-e, --env-file string   # Arquivo .env (default: .env)
```

**Output:**
```
→ Loading secrets for: [DATABASE_URL API_KEY]
  Loaded from .env
→ Connecting to manager.example.com...
→ Pushing 2 secret(s)...
  → DATABASE_URL... ✓
  → API_KEY... ✓

✓ Secrets pushed successfully
```

### secrets list

Lista secrets existentes para o stack.

```bash
swarmctl secrets list
```

**Output:**
```
→ Secrets for stack myapp:
  - myapp_database_url
  - myapp_api_key
```

---

## swarmctl accessory

Gerencia serviços auxiliares (Redis, PostgreSQL, etc).

### accessory (list)

Lista status dos accessories.

```bash
swarmctl accessory
```

**Output:**
```
→ Accessories for stack myapp:

  NAME                 REPLICAS        STATUS
  redis                1/1             running
  postgres             1/1             running
```

### accessory start

Inicia um accessory (scale para 1).

```bash
swarmctl accessory start redis
swarmctl accessory start all    # Inicia todos
```

### accessory stop

Para um accessory (scale para 0).

```bash
swarmctl accessory stop redis
swarmctl accessory stop all
```

### accessory restart

Reinicia um accessory (force update).

```bash
swarmctl accessory restart redis
swarmctl accessory restart all
```
