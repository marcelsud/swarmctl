#!/bin/bash

# Script de setup para testes de integração
# Prepara o ambiente antes dos testes

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEST_DIR="$(dirname "$SCRIPT_DIR")"
CONFIG_DIR="$TEST_DIR/configs"

# Cores
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}[SETUP]${NC} Preparando ambiente de teste..."

# Verificar se o cluster Swarm está ativo
echo "Verificando cluster Swarm..."
if multipass exec swarm-manager -- docker node ls > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC} Cluster Swarm está ativo"
else
    echo -e "${RED}✗${NC} Cluster Swarm não está acessível"
    exit 1
fi

# Criar diretórios necessários
mkdir -p "$TEST_DIR/results"
mkdir -p "$TEST_DIR/logs"

# Copiar binário swarmctl se não existir
if [[ ! -f "$TEST_DIR/../swarmctl" ]]; then
    echo "Copiando binário swarmctl..."
    cp "$TEST_DIR/../../swarmctl" "$TEST_DIR/../swarmctl" 2> /dev/null || echo "Binário não encontrado, usará o do PATH"
fi

# Verificar conectividade SSH (simplificado)
echo "Verificando conectividade entre nodes..."
if multipass exec swarm-manager -- ping -c 1 worker-1 > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC} Rede entre nodes funcionando"
else
    echo -e "${RED}✗${NC} Rede entre nodes com problemas"
fi

# Criar arquivo de ambiente para secrets
echo "Criando arquivo de secrets..."
# Limpar stacks anteriores
echo "Limpando stacks de teste anteriores..."
multipass exec swarm-manager -- docker stack ls | grep "test-app\|integration" | awk '{print $1}' | while read stack; do
    echo "Removendo stack: $stack"
    multipass exec swarm-manager -- docker stack rm "$stack" || true
done

sleep 5

# Limpar secrets anteriores
echo "Limpando secrets anteriores..."
multipass exec swarm-manager -- docker secret ls | grep "test-app\|integration" | awk '{print $1}' | while read secret; do
    echo "Removendo secret: $secret"
    multipass exec swarm-manager -- docker secret rm "$secret" || true
done

echo -e "${GREEN}[SETUP]${NC} Ambiente preparado com sucesso!"
echo ""
echo "Pronto para executar os testes:"
echo "  cd $TEST_DIR"
echo "  ./run-tests.sh"