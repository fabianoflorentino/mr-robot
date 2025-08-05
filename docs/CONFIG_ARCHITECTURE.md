# Arquitetura do Diretório Config - Guia de Manutenção

Este documento serve como guia para desenvolvedores que irão realizar manutenção e adicionar novas configurações na aplicação mr-robot.

## 📋 Índice

- [Visao Geral](#visao-geral)
- [Estrutura do Diretorio Config](#estrutura-do-diretorio-config)
- [Sistema de Configuracao](#sistema-de-configuracao)
- [Como Adicionar Nova Configuracao](#como-adicionar-nova-configuracao)
- [Configuracoes por Ambiente](#configuracoes-por-ambiente)
- [Padroes e Convencoes](#padroes-e-convencoes)
- [Testes](#testes)
- [Troubleshooting](#troubleshooting)

## 🎯 Visao Geral

O diretório `config/` é responsável por todo o **gerenciamento de configurações** da aplicação e implementa:

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

## 📏 Padroes e Convencoes

### ✅ Boas Práticas de Configuração

- **📋 Nomenclatura Consistente**: `{AREA}_{PROPRIEDADE}`
- **🔒 Segurança**: Nunca commitar senhas ou secrets
- **📖 Documentação**: Comentar cada configuração
- **✅ Validação**: Validar configurações críticas
- **🎯 Valores Padrão**: Sempre fornecer fallbacks sensatos

### 📋 Convenções de Nomenclatura

| Tipo | Padrão | Exemplo |
|------|---------|---------|
| **Variável de Ambiente** | `{AREA}_{PROPRIEDADE}` | `DATABASE_HOST`, `PAYMENT_URL` |
| **Struct de Config** | `{Area}Config` | `DatabaseConfig`, `PaymentConfig` |
| **Campo de Struct** | `PascalCase` | `MaxRetries`, `EnableCache` |
| **Arquivo .env** | `.env.{ambiente}` | `.env`, `.env.test`, `.env.prod` |

### 🔒 Configurações Sensíveis

```bash
# ❌ Nunca fazer isso (commitar senhas)
POSTGRES_PASSWORD=super_secret_password

# ✅ Usar referências a secrets
POSTGRES_PASSWORD=${DB_PASSWORD}
POSTGRES_PASSWORD_FILE=/run/secrets/db_password

# ✅ Ou deixar vazio para ser definido no ambiente
POSTGRES_PASSWORD=
```

### 📊 Tipos de Dados Suportados

```go
// Tipos básicos
Host     string
Port     int
Timeout  time.Duration
Enabled  bool

// Conversão automática
timeout := getEnvOrDefault("TIMEOUT", "30s")          // string
timeoutDuration, _ := time.ParseDuration(timeout)     // time.Duration

retries := getEnvOrDefault("RETRIES", "3")            // string
retriesInt, _ := strconv.Atoi(retries)                // int

enabled := getEnvOrDefault("ENABLED", "true")         // string
enabledBool, _ := strconv.ParseBool(enabled)          // bool
```

## 🧪 Testes

### Testando Carregamento de Configuração

```go
func TestLoadAppConfig_Success(t *testing.T) {
    // Setup environment
    envVars := map[string]string{
        "POSTGRES_HOST":     "test-host",
        "POSTGRES_PORT":     "5432",
        "POSTGRES_USER":     "test-user",
        "POSTGRES_PASSWORD": "test-pass",
        "POSTGRES_DB":       "test-db",
        "QUEUE_WORKERS":     "5",
    }

    for key, value := range envVars {
        os.Setenv(key, value)
        defer os.Unsetenv(key)
    }

    // Act
    config, err := LoadAppConfig()

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, config)
    assert.Equal(t, "test-host", config.Database.Host)
    assert.Equal(t, "5432", config.Database.Port)
    assert.Equal(t, 5, config.Queue.Workers)
}

func TestLoadAppConfig_DefaultValues(t *testing.T) {
    // Setup - limpar todas as env vars relacionadas
    envVars := []string{"POSTGRES_HOST", "POSTGRES_PORT", "QUEUE_WORKERS"}
    for _, key := range envVars {
        os.Unsetenv(key)
    }

    // Act
    config, err := LoadAppConfig()

    // Assert
    assert.NoError(t, err)
    assert.Equal(t, "localhost", config.Database.Host)
    assert.Equal(t, "5432", config.Database.Port)
    assert.Equal(t, 10, config.Queue.Workers) // valor padrão
}
```

### Testando Validação

```go
func TestNovaConfig_Validate(t *testing.T) {
    tests := []struct {
        name    string
        config  NovaConfig
        wantErr bool
        errMsg  string
    }{
        {
            name: "configuração válida",
            config: NovaConfig{
                Endpoint:    "http://localhost:8080",
                Timeout:     30 * time.Second,
                MaxRetries:  3,
                EnableCache: true,
                CacheExpiry: 1 * time.Hour,
            },
            wantErr: false,
        },
        {
            name: "endpoint vazio",
            config: NovaConfig{
                Endpoint: "",
            },
            wantErr: true,
            errMsg:  "NOVA_ENDPOINT é obrigatório",
        },
        {
            name: "timeout negativo",
            config: NovaConfig{
                Endpoint: "http://localhost:8080",
                Timeout:  -1 * time.Second,
            },
            wantErr: true,
            errMsg:  "NOVA_TIMEOUT deve ser positivo",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.config.Validate()

            if tt.wantErr {
                assert.Error(t, err)
                if tt.errMsg != "" {
                    assert.Contains(t, err.Error(), tt.errMsg)
                }
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### Testando Múltiplos Ambientes

```go
func TestConfigEnvironments(t *testing.T) {
    tests := []struct {
        name        string
        envFile     string
        expectedDB  string
        expectedLog string
    }{
        {
            name:        "desenvolvimento",
            envFile:     ".env",
            expectedDB:  "mr_robot_dev",
            expectedLog: "debug",
        },
        {
            name:        "teste",
            envFile:     ".env.test",
            expectedDB:  "mr_robot_test",
            expectedLog: "error",
        },
        {
            name:        "produção",
            envFile:     ".env.prod",
            expectedDB:  "mr_robot_prod",
            expectedLog: "info",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Carregar arquivo específico
            err := godotenv.Load(tt.envFile)
            assert.NoError(t, err)

            // Testar configuração
            config, err := LoadAppConfig()
            assert.NoError(t, err)
            assert.Equal(t, tt.expectedDB, config.Database.Database)
            assert.Equal(t, tt.expectedLog, os.Getenv("LOG_LEVEL"))
        })
    }
}
```

## 🔧 Troubleshooting

### Problemas Comuns

| Problema | Causa Provável | Solução |
|----------|----------------|---------|
| **Config não carrega** | Arquivo .env não encontrado | Verificar se `.env` existe e está no local correto |
| **Valor sempre padrão** | Variável de ambiente mal formatada | Verificar nome da variável (case-sensitive) |
| **Parsing error** | Formato inválido (duration, int, bool) | Verificar formato: `"30s"`, `"123"`, `"true"` |
| **Configuração não aplicada** | Container DI não atualizado | Restart da aplicação após mudança de config |
| **Secret não carregado** | Arquivo de secret não existe | Verificar mounts e paths dos secrets |

### Debug de Configuração

```go
// Adicionar logs de debug no LoadAppConfig()
func LoadAppConfig() (*AppConfig, error) {
    log.Println("Loading application configuration...")

    if err := LoadEnv(); err != nil {
        log.Printf("Failed to load .env file: %v", err)
        return nil, err
    }

    // Log variáveis carregadas (cuidado com senhas!)
    log.Printf("POSTGRES_HOST: %s", getEnvOrDefault("POSTGRES_HOST", "localhost"))
    log.Printf("QUEUE_WORKERS: %s", getEnvOrDefault("QUEUE_WORKERS", "10"))

    config := &AppConfig{
        // ... configurações ...
    }

    log.Printf("Configuration loaded successfully: %+v", config)
    return config, nil
}
```

### Verificações de Configuração

```bash
# Verificar se variáveis estão definidas
env | grep POSTGRES
env | grep QUEUE
env | grep PAYMENT

# Testar parsing de duração
echo $NOVA_TIMEOUT
# Deve retornar formato válido como "30s", "1m", "1h"

# Verificar arquivo .env
cat .env | grep -v "PASSWORD\|SECRET\|TOKEN"

# Teste de conectividade com configuração
curl -f $DEFAULT_PROCESSOR_URL/health
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

## 📊 Monitoramento de Configuração

### Health Check de Configuração

```go
func (c *AppConfig) HealthCheck() map[string]string {
    status := make(map[string]string)

    // Verificar conexão com banco
    if c.Database.Host != "" && c.Database.Port != "" {
        status["database"] = "configured"
    } else {
        status["database"] = "missing_config"
    }

    // Verificar processadores
    if c.Payment.DefaultProcessorURL != "" {
        status["default_processor"] = "configured"
    } else {
        status["default_processor"] = "missing_config"
    }

    if c.Payment.FallbackProcessorURL != "" {
        status["fallback_processor"] = "configured"
    } else {
        status["fallback_processor"] = "missing_config"
    }

    return status
}
```

### Endpoint de Configuração (para debug)

```go
// Apenas em ambiente de desenvolvimento
func (hc *HealthController) ConfigStatus(c *gin.Context) {
    if os.Getenv("GIN_MODE") == "release" {
        c.JSON(http.StatusForbidden, gin.H{"error": "Not available in production"})
        return
    }

    config := hc.appConfig
    status := config.HealthCheck()

    c.JSON(http.StatusOK, gin.H{
        "config_status": status,
        "environment":   os.Getenv("GIN_MODE"),
        "version":       os.Getenv("APP_VERSION"),
    })
}
```

## 📞 Contato

Para dúvidas sobre configurações ou sugestões de melhorias, abra uma issue no repositório ou entre em contato com a equipe de desenvolvimento.

---

**📝 Nota**: Este documento deve ser atualizado sempre que novas configurações ou padrões forem adicionados à aplicação.
