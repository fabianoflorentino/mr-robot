# Arquitetura do Diretório Config - Guia de Manutenção

> **Consulte também**: [📖 ARCHITECTURE_GUIDE.md](ARCHITECTURE_GUIDE.md) para padrões gerais e convenções consolidadas.

Este documento foca especificamente no **diretório `config`** e seu sistema de gerenciamento de configurações.

## 🎯 Responsabilidades Específicas das Configurações

- ⚙️ **Carregamento de Variáveis**: Variáveis de ambiente e arquivos `.env`
- 🏗️ **Estruturas Tipadas**: Configurações organizadas por domínio
- 🔒 **Validação**: Verificação de configurações obrigatórias
- 🌍 **Multi-ambiente**: Suporte a desenvolvimento, teste e produção
- 📋 **Valores Padrão**: Fallbacks para configurações opcionais

## 📁 Estrutura do Diretorio Config

```text
config/
├── .env                    # 🔒 Variáveis de ambiente locais (não commitado)
├── app_config.go          # 🏗️ Estruturas principais de configuração
├── config.go              # 📋 Utilitários e carregamento de .env
├── haproxy.cfg           # ⚖️ Configuração do balanceador de carga
└── postgresql.conf       # 🗄️ Configuração específica do PostgreSQL
```

### 🧩 Componentes Principais

| Componente | Responsabilidade | Arquivo Principal | Tipo |
|------------|------------------|-------------------|------|
| **AppConfig** | Estrutura principal de config | `app_config.go` | Estrutura |
| **LoadAppConfig()** | Carregamento das configurações | `app_config.go` | Função |
| **LoadEnv()** | Carregamento de arquivos .env | `config.go` | Função |
| **getEnvOrDefault()** | Utilitário para variáveis | `app_config.go` | Função |

## 🏗️ Sistema de Configuracao

### Estrutura Principal (AppConfig)

```go
type AppConfig struct {
    Database DatabaseConfig  // 🗄️ Configurações de banco
    Payment  PaymentConfig   // 💳 Configurações de pagamento
    Queue    QueueConfig     // 📬 Configurações de fila
}

type DatabaseConfig struct {
    Host     string  // Endereço do banco
    Port     string  // Porta do banco
    User     string  // Usuário
    Password string  // Senha
    Database string  // Nome do banco
    SSLMode  string  // Modo SSL
    Timezone string  // Timezone
}

type PaymentConfig struct {
    DefaultProcessorURL  string  // URL do processador principal
    FallbackProcessorURL string  // URL do processador fallback
}

type QueueConfig struct {
    Workers               int  // Número de workers
    BufferSize            int  // Tamanho do buffer
    MaxEnqueueRetries     int  // Máximo de tentativas
    MaxSimultaneousWrites int  // Escritas simultâneas no DB
}
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
GIN_MODE=debug

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
GIN_MODE=test

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
GIN_MODE=release

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
GIN_MODE=debug
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
GIN_MODE=release
```

---

**📝 Nota**: Para padrões gerais, convenções de nomenclatura e troubleshooting consolidado, consulte o [📖 ARCHITECTURE_GUIDE.md](ARCHITECTURE_GUIDE.md).
