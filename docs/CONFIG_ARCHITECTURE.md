# Nova Arquitetura de Configurações - Guia Completo

> **⚠️ ATENÇÃO**: Esta documentação refere-se à nova arquitetura de configurações implementada em agosto de 2025. 
> Para compatibilidade com código legado, consulte [CONFIG_REFACTORING.md](CONFIG_REFACTORING.md).

## 🎯 Nova Abordagem de Configurações

### Princípios da Nova Arquitetura

- 🔒 **Segurança**: Cada manager acessa apenas suas configurações específicas
- 🏗️ **Modularidade**: Configurações isoladas por domínio de responsabilidade
- ✅ **Validação**: Verificação específica por tipo de configuração
- 🧪 **Testabilidade**: Facilita mocking e testes unitários
- 🌍 **Flexibilidade**: Uso centralizado ou individual

## 📁 Nova Estrutura de Configurações

```text
internal/app/
├── config/
│   ├── manager.go          # 🎯 Manager coordenador principal
│   └── manager_test.go     # ✅ Testes de integração
├── database/
│   ├── config.go           # �️ Configurações de banco de dados
│   ├── manager.go          # 🗄️ Manager de database
│   └── config_test.go      # ✅ Testes específicos
├── payment/
│   ├── config.go           # 💳 Configurações de pagamento
│   └── config_test.go      # ✅ Testes específicos
├── queue/
│   ├── config.go           # � Configurações de fila
│   ├── payment_queue.go    # 📬 Implementação da fila
│   └── config_test.go      # ✅ Testes específicos
├── circuitbreaker/
│   ├── config.go           # ⚡ Configurações de circuit breaker
│   └── config_test.go      # ✅ Testes específicos
└── controller/
    ├── config.go           # 🌐 Configurações de controller
    └── config_test.go      # ✅ Testes específicos

# Arquivos legados (para compatibilidade)
config/
├── app_config.go          # 📛 DEPRECATED - Manter para compatibilidade
├── config.go              # 📋 Utilitários (ainda usado)
├── haproxy.cfg           # ⚖️ Configuração do balanceador
└── postgresql.conf       # 🗄️ Configuração do PostgreSQL
```

### 🧩 Novos Managers de Configuração

| Manager | Responsabilidade | Arquivo | Validações |
|---------|------------------|---------|------------|
| **config.Manager** | Coordenação geral | `internal/app/config/manager.go` | Orquestra todos os managers |
| **database.ConfigManager** | Configurações de DB | `internal/app/database/config.go` | Host, Port, SSL, Timezone |
| **payment.ConfigManager** | URLs de pagamento | `internal/app/payment/config.go` | URLs obrigatórias e válidas |
| **queue.ConfigManager** | Configurações de fila | `internal/app/queue/config.go` | Workers > 0, Buffer > 0 |
| **circuitbreaker.ConfigManager** | Circuit breaker | `internal/app/circuitbreaker/config.go` | Timeouts > 0, Limites > 0 |
| **controller.ConfigManager** | Configurações HTTP | `internal/app/controller/config.go` | Hostname válido |

## 🏗️ Como Usar os Novos Managers

### Uso Centralizado (Recomendado para aplicações completas)

```go
import "github.com/fabianoflorentino/mr-robot/internal/app/config"

// Carrega e valida todas as configurações
configManager := config.NewManager()
err := configManager.LoadConfiguration()
if err != nil {
    log.Fatalf("Failed to load configuration: %v", err)
}

err = configManager.ValidateConfiguration()
if err != nil {
    log.Fatalf("Invalid configuration: %v", err)
}

// Acesso às configurações específicas
dbConfig := configManager.GetDatabaseConfig()
paymentConfig := configManager.GetPaymentConfig()
queueConfig := configManager.GetQueueConfig()
cbConfig := configManager.GetCircuitBreakerConfig()
controllerConfig := configManager.GetControllerConfig()
```

### Uso Individual (Recomendado para microserviços)

```go
import (
    "github.com/fabianoflorentino/mr-robot/internal/app/database"
    "github.com/fabianoflorentino/mr-robot/internal/app/payment"
)

// Apenas configurações de database
dbConfigManager := database.NewConfigManager()
err := dbConfigManager.LoadConfig()
if err != nil {
    log.Fatalf("Failed to load database config: %v", err)
}

err = dbConfigManager.Validate()
if err != nil {
    log.Fatalf("Invalid database config: %v", err)
}

dbConfig := dbConfigManager.GetConfig()

// Apenas configurações de payment
paymentConfigManager := payment.NewConfigManager()
err = paymentConfigManager.LoadConfig()
// ... validação e uso
```

### Fluxo de Carregamento

```mermaid
graph TD
    A[LoadAppConfig()] --> B[LoadEnv()]
    B --> C[Ler variáveis de ambiente]
    C --> D[Aplicar valores padrão]
    D --> E[Validar configurações]
    E --> F[Retornar AppConfig]
```

### Precedência de Configuração

1. **🔴 Variáveis de Ambiente**: Maior prioridade
2. **🟡 Arquivo .env**: Prioridade média
3. **🟢 Valores Padrão**: Menor prioridade (fallback)

## ➕ Como Adicionar Nova Configuracao

### Passo 1: Definir Estrutura da Configuração

Adicione em `app_config.go`:

```go
// Nova estrutura de configuração
type NovaConfig struct {
    Endpoint     string        // URL do serviço
    Timeout      time.Duration // Timeout das requisições
    MaxRetries   int          // Máximo de tentativas
    EnableCache  bool         // Habilitar cache
    CacheExpiry  time.Duration // Tempo de expiração do cache
}

// Adicionar na AppConfig principal
type AppConfig struct {
    Database DatabaseConfig
    Payment  PaymentConfig
    Queue    QueueConfig
    Nova     NovaConfig  // ⬅️ Nova configuração aqui
}
```

### Passo 2: Implementar Carregamento

Na função `LoadAppConfig()`:

```go
func LoadAppConfig() (*AppConfig, error) {
    if err := LoadEnv(); err != nil {
        return nil, fmt.Errorf("failed to load environment: %w", err)
    }

    // ... configurações existentes ...

    // Carregar configurações da nova área
    timeout, err := time.ParseDuration(getEnvOrDefault("NOVA_TIMEOUT", "30s"))
    if err != nil {
        timeout = 30 * time.Second
    }

    maxRetries, err := strconv.Atoi(getEnvOrDefault("NOVA_MAX_RETRIES", "3"))
    if err != nil {
        maxRetries = 3
    }

    enableCache, err := strconv.ParseBool(getEnvOrDefault("NOVA_ENABLE_CACHE", "true"))
    if err != nil {
        enableCache = true
    }

    cacheExpiry, err := time.ParseDuration(getEnvOrDefault("NOVA_CACHE_EXPIRY", "1h"))
    if err != nil {
        cacheExpiry = 1 * time.Hour
    }

    return &AppConfig{
        // ... configurações existentes ...
        Nova: NovaConfig{
            Endpoint:     getEnvOrDefault("NOVA_ENDPOINT", "http://localhost:8080"),
            Timeout:      timeout,
            MaxRetries:   maxRetries,
            EnableCache:  enableCache,
            CacheExpiry:  cacheExpiry,
        },
    }, nil
}
```

### Passo 3: Adicionar Variáveis de Ambiente

Adicione no arquivo `.env`:

```bash
# Nova configuração
NOVA_ENDPOINT=http://nova-service:8080
NOVA_TIMEOUT=45s
NOVA_MAX_RETRIES=5
NOVA_ENABLE_CACHE=true
NOVA_CACHE_EXPIRY=2h
```

### Passo 4: Documentar Configuração

Adicione comentários detalhados:

```go
type NovaConfig struct {
    // Endpoint é a URL base do serviço externo
    // Exemplo: http://nova-service:8080
    // Variável: NOVA_ENDPOINT
    Endpoint string

    // Timeout para requisições HTTP ao serviço
    // Formato: "30s", "1m", "1h30m"
    // Variável: NOVA_TIMEOUT
    // Padrão: 30s
    Timeout time.Duration

    // MaxRetries define o número máximo de tentativas
    // em caso de falha na requisição
    // Variável: NOVA_MAX_RETRIES
    // Padrão: 3
    MaxRetries int

    // EnableCache habilita o sistema de cache
    // Variável: NOVA_ENABLE_CACHE
    // Padrão: true
    EnableCache bool

    // CacheExpiry define o tempo de expiração do cache
    // Formato: "1h", "30m", "24h"
    // Variável: NOVA_CACHE_EXPIRY
    // Padrão: 1h
    CacheExpiry time.Duration
}
```

### Passo 5: Adicionar Validação (Opcional)

```go
func (c *NovaConfig) Validate() error {
    if c.Endpoint == "" {
        return fmt.Errorf("NOVA_ENDPOINT é obrigatório")
    }

    if c.Timeout <= 0 {
        return fmt.Errorf("NOVA_TIMEOUT deve ser positivo")
    }

    if c.MaxRetries < 0 {
        return fmt.Errorf("NOVA_MAX_RETRIES deve ser >= 0")
    }

    if c.CacheExpiry <= 0 && c.EnableCache {
        return fmt.Errorf("NOVA_CACHE_EXPIRY deve ser positivo quando cache está habilitado")
    }

    return nil
}

// Chamar validação no LoadAppConfig()
config := &AppConfig{ /* ... */ }
if err := config.Nova.Validate(); err != nil {
    return nil, fmt.Errorf("configuração Nova inválida: %w", err)
}
```

## 🌍 Configuracoes por Ambiente

### Desenvolvimento (.env)

```bash
# Banco de dados local
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=dev_password
POSTGRES_DB=mr_robot_dev

# Processadores locais
DEFAULT_PROCESSOR_URL=http://localhost:8080/process
FALLBACK_PROCESSOR_URL=http://localhost:8081/process

# Debug habilitado
DEBUG=true
LOG_LEVEL=debug

# Queue com poucos workers para debug
QUEUE_WORKERS=2
QUEUE_BUFFER_SIZE=10
```

### Teste (.env.test)

```bash
# Banco de dados de teste
POSTGRES_HOST=localhost
POSTGRES_PORT=5433
POSTGRES_USER=test_user
POSTGRES_PASSWORD=test_password
POSTGRES_DB=mr_robot_test

# Processadores mock
DEFAULT_PROCESSOR_URL=http://mock-processor:8080
FALLBACK_PROCESSOR_URL=http://mock-processor:8081

# Logs mínimos para testes
DEBUG=false
LOG_LEVEL=error

# Queue rápida para testes
QUEUE_WORKERS=1
QUEUE_BUFFER_SIZE=5
QUEUE_MAX_ENQUEUE_RETRIES=1
```

### Produção (via Docker/K8s)

```bash
# Banco de dados produção
POSTGRES_HOST=production-db.internal
POSTGRES_PORT=5432
POSTGRES_USER=app_user
POSTGRES_PASSWORD=${DB_SECRET}
POSTGRES_DB=mr_robot_prod

# Processadores externos
DEFAULT_PROCESSOR_URL=https://payment-processor.company.com/api/v1/process
FALLBACK_PROCESSOR_URL=https://backup-processor.company.com/api/v1/process

# Otimizado para produção
DEBUG=false
LOG_LEVEL=info

# Queue dimensionada para carga
QUEUE_WORKERS=10
QUEUE_BUFFER_SIZE=10000
QUEUE_MAX_SIMULTANEOUS_WRITES=50
```

### Comandos Úteis

```bash
# Validar sintaxe do .env
docker run --rm -v $(pwd)/.env:/tmp/.env \
  alpine/alpine:latest sh -c "source /tmp/.env && env"

# Testar configuração em container
docker-compose exec mr_robot_app env | grep POSTGRES

# Verificar se configuração foi aplicada
docker-compose exec mr_robot_app \
  sh -c "echo 'Config test' && curl localhost:8888/health"
```

## 📊 Configurações Específicas por Ambiente

### Desenvolvimento Local (.env)

```bash
# Banco de dados local
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=dev_password
POSTGRES_DB=mr_robot_dev

# Debug habilitado
DEBUG=true
LOG_LEVEL=debug
HTTP_TIMEOUT=30s
```

### Produção Cloud (via Docker/K8s)

```bash
# Banco de dados produção
POSTGRES_HOST=production-db.internal
POSTGRES_PORT=5432
POSTGRES_USER=app_user
POSTGRES_PASSWORD=${DB_SECRET}
POSTGRES_DB=mr_robot_prod

# Otimizado para produção
DEBUG=false
LOG_LEVEL=info
```

---

**📝 Nota**: Para padrões gerais, convenções de nomenclatura e troubleshooting consolidado, consulte o [📖 ARCHITECTURE_GUIDE.md](ARCHITECTURE_GUIDE.md).
