# Modo Compose

O swarmctl suporta Docker Compose como alternativa ao Docker Swarm para deployments. Útil quando:

- Docker Swarm não está disponível ou não é desejado
- Você quer deployments mais simples em um único node
- Você está desenvolvendo localmente antes de fazer deploy para Swarm

## Configuração

Ative o modo compose definindo `mode: compose` no seu swarm.yaml:

```yaml
stack: myapp
mode: compose
compose_file: docker-compose.yaml

# SSH é opcional - omita para deploy local
ssh:
  host: server.example.com
  user: deploy
```

## Como Funciona

No modo compose, o swarmctl usa `docker compose` (plugin v2) ao invés do Docker Swarm:

| Operação | Modo Swarm | Modo Compose |
|----------|------------|--------------|
| Deploy | `docker stack deploy` | `docker compose up -d` |
| Remove | `docker stack rm` | `docker compose down` |
| Status | `docker service ls` | `docker compose ps` |
| Logs | `docker service logs` | `docker compose logs` |
| Exec | Encontra container da task | Encontra container do compose |

## Comandos

Todos os comandos do swarmctl funcionam no modo compose:

```bash
swarmctl setup      # Verifica Docker e plugin compose
swarmctl deploy     # Faz deploy com docker compose
swarmctl status     # Mostra containers ao invés de tasks
swarmctl logs       # Usa docker compose logs
swarmctl exec       # Executa no container do compose
swarmctl rollback   # Rollback via histórico (veja abaixo)
```

## Suporte a Rollback

O modo compose suporta rollback através de um container sidecar de histórico que armazena deployments anteriores.

### Como o Rollback Funciona

1. A cada deploy, o arquivo compose e metadados das imagens são registrados
2. O histórico é armazenado em um banco SQLite dentro de um container
3. O rollback recupera o arquivo compose anterior e faz redeploy

### Container de Histórico

O container de histórico (`{stack}-history`) é iniciado automaticamente durante o deploy. Ele usa a imagem `docker.io/marcelsud/swarmctl-history`.

Se o container de histórico não estiver disponível, o rollback mostrará um aviso, mas o deploy continuará.

```bash
# Rollback para versão anterior
swarmctl rollback
```

**Nota:** No modo compose, o rollback afeta todos os serviços de uma vez (rollback de serviço individual não é suportado).

## Limitações

### Sem Scale Dinâmico

O modo compose não suporta a operação `scale`:

```bash
swarmctl scale web=3  # Erro: scale não suportado no modo compose
```

Para escalar no modo compose, atualize seu docker-compose.yaml com `deploy.replicas` e faça redeploy.

### Rollback Tudo-ou-Nada

Diferente do modo Swarm onde serviços individuais podem sofrer rollback, o modo compose faz rollback de todo o deployment de uma vez.

## Requisitos do Setup

O comando `swarmctl setup` no modo compose verifica:

1. Docker está instalado
2. Plugin Docker Compose v2 está disponível (`docker compose version`)
3. Login no registry (se configurado)

Ele **não** inicializa Swarm ou cria overlay networks.

## Exemplos

### Desenvolvimento Local

```yaml
# swarm.yaml
stack: myapp
mode: compose
compose_file: docker-compose.yaml
```

```bash
swarmctl setup   # Verifica docker compose
swarmctl deploy  # Deploy local
swarmctl status  # Verificar status
swarmctl logs web -f  # Acompanhar logs
```

### Deploy Remoto

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
swarmctl setup   # Verifica docker compose remoto
swarmctl deploy  # Deploy no servidor remoto
swarmctl status  # Verificar status remoto
```

## Alternando Entre Modos

Você pode alternar entre os modos swarm e compose mudando o campo `mode`:

```yaml
# swarm.yaml para compose
mode: compose
```

```yaml
# swarm.yaml para swarm (default)
mode: swarm
# ou simplesmente omita o campo mode
```

**Nota:** Alternar entre modos requer remover o deployment existente primeiro, já que stacks Swarm e projetos Compose são gerenciados separadamente.
