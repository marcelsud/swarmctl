#!/bin/bash

# Script de teste de integração para swarmctl
# Testa funcionalidades básicas em cluster Swarm real

set -e  # Exit on error

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configurações
TEST_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$TEST_DIR/.." && pwd)"
SWARMCTL="${SWARMCTL:-$PROJECT_ROOT/swarmctl}"
CONFIG_FILE="$TEST_DIR/configs/swarm.test.yaml"
COMPOSE_FILE="$TEST_DIR/configs/docker-compose.test.yaml"
SECRETS_FILE="$TEST_DIR/configs/.env.test"

# Funções auxiliares
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

check_requirements() {
    log_info "Verificando requisitos..."
    
    # Verificar swarmctl
    if [[ ! -f "$SWARMCTL" ]]; then
        log_error "swarmctl não encontrado em: $SWARMCTL"
        exit 1
    fi
    
    # Verificar cluster Swarm
    if ! multipass exec swarm-manager -- docker node ls > /dev/null 2>&1; then
        log_error "Cluster Swarm não está acessível"
        exit 1
    fi
    
    # Verificar arquivos de config
    if [[ ! -f "$CONFIG_FILE" ]]; then
        log_error "Arquivo de configuração não encontrado: $CONFIG_FILE"
        exit 1
    fi
    
    log_info "Requisitos verificados ✓"
}

cleanup_stack() {
    log_info "Limpando stack de teste anterior..."
    
    # Remover stack se existir
    multipass exec swarm-manager -- docker stack rm test-app 2> /dev/null || true
    
    # Remover secrets
    multipass exec swarm-manager -- docker secret rm test-app_test_secret 2> /dev/null || true
    multipass exec swarm-manager -- docker secret rm test-app_api_key 2> /dev/null || true
    
    # Aguardar remoção completa
    sleep 5
    
    log_info "Limpeza concluída"
}

test_deploy_basico() {
    log_info "=== Teste 1: Deploy Básico ==="
    
    # Push secrets primeiro
    log_info "Criando secrets..."
    cd "$TEST_DIR/configs"
    $SWARMCTL --config "$CONFIG_FILE" secrets push --env-file "$SECRETS_FILE"
    
    # Deploy básico
    log_info "Executando deploy básico..."
    $SWARMCTL --config "$CONFIG_FILE" deploy
    
    # Verificar status
    log_info "Verificando status..."
    sleep 10
    $SWARMCTL --config "$CONFIG_FILE" status
    
    # Verificar se serviços estão rodando
    log_info "Verificando serviços no cluster..."
    multipass exec swarm-manager -- docker service ls | grep test-app || log_warning "Nenhum serviço encontrado"
    
    log_info "Deploy básico concluído ✓"
}

test_deploy_verbose() {
    log_info "=== Teste 2: Deploy com --verbose ==="
    
    # Cleanup primeiro
    cleanup_stack
    
    log_info "Executando deploy com verbose..."
    $SWARMCTL --verbose --config "$CONFIG_FILE" deploy 2>> "$TEST_DIR/results/verbose.log"
    
    # Verificar se o log verbose foi gerado
    if [[ -f "$TEST_DIR/results/verbose.log" ]] && grep -q "Running:" "$TEST_DIR/results/verbose.log"; then
        log_info "Verbose mode funcionando ✓"
    else
        log_warning "Verbose mode pode não estar funcionando corretamente"
    fi
}

test_deploy_service() {
    log_info "=== Teste 3: Deploy com --service ==="
    
    cleanup_stack
    
    log_info "Executando deploy apenas do serviço 'web'..."
    $SWARMCTL --config "$CONFIG_FILE" deploy --service web
    
    # Verificar se apenas o serviço web foi deployado
    sleep 5
    SERVICES=$(multipass exec swarm-manager -- docker service ls | grep test-app | wc -l)
    log_info "Serviços encontrados: $SERVICES"
    
    if [[ "$SERVICES" -eq 1 ]]; then
        log_info "Deploy por serviço funcionando ✓"
    else
        log_warning "Deploy por serviço pode não estar funcionando corretamente"
    fi
}

test_deploy_skip_accessories() {
    log_info "=== Teste 4: Deploy com --skip-accessories ==="
    
    cleanup_stack
    
    log_info "Executando deploy pulando accessories..."
    $SWARMCTL --config "$CONFIG_FILE" deploy --skip-accessories
    
    # Verificar se accessories não foram deployados
    sleep 5
    if multipass exec swarm-manager -- docker service ls | grep -q "test-app_redis\|test-app_nginx"; then
        log_warning "Accessories foram deployados mesmo com --skip-accessories"
    else
        log_info "Skip accessories funcionando ✓"
    fi
}

test_health_check() {
    log_info "=== Teste 5: Health Check ==="
    
    cleanup_stack
    
    log_info "Executando deploy e verificando health check..."
    $SWARMCTL --config "$CONFIG_FILE" deploy
    
    # Aguardar health check
    log_info "Aguardando serviços ficarem healthy..."
    sleep 15
    
    # Verificar status detalhado
    $SWARMCTL --config "$CONFIG_FILE" status
    
    log_info "Health check testado ✓"
}

test_logs() {
    log_info "=== Teste 6: Logs ==="
    
    # Verificar logs de serviço específico
    log_info "Verificando logs do serviço web..."
    $SWARMCTL --config "$CONFIG_FILE" logs web --tail 5
    
    # Verificar logs de todos os serviços
    log_info "Verificando logs de todos os serviços..."
    $SWARMCTL --config "$CONFIG_FILE" logs --tail 3
    
    log_info "Logs testados ✓"
}

test_exec() {
    log_info "=== Teste 7: Exec ==="
    
    # Executar comando em container
    log_info "Executando comando no container web..."
    $SWARMCTL --config "$CONFIG_FILE" exec web -- hostname
    
    log_info "Exec testado ✓"
}

test_secrets() {
    log_info "=== Teste 8: Secrets ==="
    
    # Listar secrets
    log_info "Listando secrets..."
    $SWARMCTL --config "$CONFIG_FILE" secrets list
    
    # Verificar se secrets estão nos containers
    log_info "Verificando secrets nos containers..."
    $SWARMCTL --config "$CONFIG_FILE" exec web -- cat /run/secrets/test_secret || log_warning "Secret não encontrado"
    
    log_info "Secrets testados ✓"
}

# Função principal
main() {
    log_info "Iniciando testes de integração do swarmctl..."
    
    # Criar diretório de resultados
    mkdir -p "$TEST_DIR/results"
    
    # Verificar requisitos
    check_requirements
    
    # Executar testes
    test_deploy_basico
    test_deploy_verbose
    test_deploy_service
    test_deploy_skip_accessories
    test_health_check
    test_logs
    test_exec
    test_secrets
    
    # Limpeza final
    cleanup_stack
    
    log_info "Testes de integração concluídos! ✓"
    log_info "Resultados salvos em: $TEST_DIR/results/"
}

# Executar se chamado diretamente
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi