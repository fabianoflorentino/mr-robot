# Guia de Implementação de Novas Configurações

> **Versão**: v2.0 - Nova Arquitetura (Agosto 2025)  
> **Objetivo**: Guia passo-a-passo para adicionar novas configurações de forma segura

## 📋 Pré-requisitos

- Familiaridade com a nova arquitetura de configurações
- Conhecimento dos princípios de segurança implementados
- Acesso ao código fonte em `internal/app/`

## 🛠️ Implementação Passo-a-Passo

### Passo 1: Criar o Diretório e Estrutura

```bash
# Criar diretório para o novo domain
mkdir -p internal/app/meuservico

# Criar arquivos base
touch internal/app/meuservico/config.go
touch internal/app/meuservico/config_test.go
```

### Passo 2: Implementar o Config Manager

Crie o arquivo `internal/app/meuservico/config.go`:

```go
package meuservico

import (
    "fmt"
    "os"
    "strconv"
    "time"
)

// Config holds meuservico-specific configuration
type Config struct {
    Endpoint    string
    Timeout     time.Duration
    MaxRetries  int
    APIKey      string // Configuração sensível
    EnableDebug bool
}

// ConfigManager manages meuservico configuration
type ConfigManager struct {
    config *Config
}

// NewConfigManager creates a new meuservico configuration manager
func NewConfigManager() *ConfigManager {
    return &ConfigManager{}
}

// LoadConfig loads configuration from environment variables
func (cm *ConfigManager) LoadConfig() error {
    // Carregar configurações obrigatórias
    endpoint := os.Getenv("MEUSERVICO_ENDPOINT")
    if endpoint == "" {
        return fmt.Errorf("MEUSERVICO_ENDPOINT environment variable is required")
    }

    apiKey := os.Getenv("MEUSERVICO_API_KEY")
    if apiKey == "" {
        return fmt.Errorf("MEUSERVICO_API_KEY environment variable is required")
    }

    // Carregar configurações opcionais com defaults
    timeout, err := time.ParseDuration(getEnvOrDefault("MEUSERVICO_TIMEOUT", "30s"))
    if err != nil {
        return fmt.Errorf("invalid MEUSERVICO_TIMEOUT value: %w", err)
    }

    maxRetries, err := strconv.Atoi(getEnvOrDefault("MEUSERVICO_MAX_RETRIES", "3"))
    if err != nil {
        return fmt.Errorf("invalid MEUSERVICO_MAX_RETRIES value: %w", err)
    }

    enableDebug, err := strconv.ParseBool(getEnvOrDefault("MEUSERVICO_ENABLE_DEBUG", "false"))
    if err != nil {
        return fmt.Errorf("invalid MEUSERVICO_ENABLE_DEBUG value: %w", err)
    }

    cm.config = &Config{
        Endpoint:    endpoint,
        Timeout:     timeout,
        MaxRetries:  maxRetries,
        APIKey:      apiKey,
        EnableDebug: enableDebug,
    }

    return nil
}

// GetConfig returns the loaded configuration
func (cm *ConfigManager) GetConfig() *Config {
    return cm.config
}

// SetConfig sets the configuration (useful for testing)
func (cm *ConfigManager) SetConfig(config *Config) {
    cm.config = config
}

// Validate validates the configuration
func (cm *ConfigManager) Validate() error {
    if cm.config == nil {
        return fmt.Errorf("meuservico configuration not loaded")
    }

    if cm.config.Endpoint == "" {
        return fmt.Errorf("endpoint cannot be empty")
    }

    if cm.config.Timeout <= 0 {
        return fmt.Errorf("timeout must be greater than 0")
    }

    if cm.config.MaxRetries < 0 {
        return fmt.Errorf("max retries cannot be negative")
    }

    if cm.config.APIKey == "" {
        return fmt.Errorf("API key cannot be empty")
    }

    // Validação de formato de URL se necessário
    if !strings.HasPrefix(cm.config.Endpoint, "http://") && 
       !strings.HasPrefix(cm.config.Endpoint, "https://") {
        return fmt.Errorf("endpoint must be a valid HTTP/HTTPS URL")
    }

    return nil
}

// getEnvOrDefault retrieves the value of an environment variable or returns a default value
func getEnvOrDefault(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
```

### Passo 3: Implementar Testes

Crie o arquivo `internal/app/meuservico/config_test.go`:

```go
package meuservico

import (
    "os"
    "testing"
    "time"
)

func TestConfigManager_LoadConfig(t *testing.T) {
    // Save original env vars
    originalVars := map[string]string{
        "MEUSERVICO_ENDPOINT":     os.Getenv("MEUSERVICO_ENDPOINT"),
        "MEUSERVICO_API_KEY":      os.Getenv("MEUSERVICO_API_KEY"),
        "MEUSERVICO_TIMEOUT":      os.Getenv("MEUSERVICO_TIMEOUT"),
        "MEUSERVICO_MAX_RETRIES":  os.Getenv("MEUSERVICO_MAX_RETRIES"),
        "MEUSERVICO_ENABLE_DEBUG": os.Getenv("MEUSERVICO_ENABLE_DEBUG"),
    }

    // Cleanup function
    defer func() {
        for key, value := range originalVars {
            if value == "" {
                os.Unsetenv(key)
            } else {
                os.Setenv(key, value)
            }
        }
    }()

    t.Run("Valid configuration", func(t *testing.T) {
        os.Setenv("MEUSERVICO_ENDPOINT", "https://api.example.com")
        os.Setenv("MEUSERVICO_API_KEY", "test-api-key")
        os.Setenv("MEUSERVICO_TIMEOUT", "45s")
        os.Setenv("MEUSERVICO_MAX_RETRIES", "5")
        os.Setenv("MEUSERVICO_ENABLE_DEBUG", "true")

        cm := NewConfigManager()
        err := cm.LoadConfig()
        if err != nil {
            t.Fatalf("Expected no error, got: %v", err)
        }

        config := cm.GetConfig()
        if config.Endpoint != "https://api.example.com" {
            t.Errorf("Expected endpoint to be 'https://api.example.com', got: %s", config.Endpoint)
        }
        if config.Timeout != 45*time.Second {
            t.Errorf("Expected timeout to be 45s, got: %s", config.Timeout)
        }
        if config.MaxRetries != 5 {
            t.Errorf("Expected max retries to be 5, got: %d", config.MaxRetries)
        }
        if !config.EnableDebug {
            t.Error("Expected debug to be enabled")
        }
    })

    t.Run("Missing required endpoint", func(t *testing.T) {
        os.Unsetenv("MEUSERVICO_ENDPOINT")
        os.Setenv("MEUSERVICO_API_KEY", "test-api-key")

        cm := NewConfigManager()
        err := cm.LoadConfig()
        if err == nil {
            t.Fatal("Expected error for missing endpoint")
        }
    })

    t.Run("Missing required API key", func(t *testing.T) {
        os.Setenv("MEUSERVICO_ENDPOINT", "https://api.example.com")
        os.Unsetenv("MEUSERVICO_API_KEY")

        cm := NewConfigManager()
        err := cm.LoadConfig()
        if err == nil {
            t.Fatal("Expected error for missing API key")
        }
    })

    t.Run("Default values", func(t *testing.T) {
        os.Setenv("MEUSERVICO_ENDPOINT", "https://api.example.com")
        os.Setenv("MEUSERVICO_API_KEY", "test-api-key")
        // Clear optional vars to test defaults
        os.Unsetenv("MEUSERVICO_TIMEOUT")
        os.Unsetenv("MEUSERVICO_MAX_RETRIES")
        os.Unsetenv("MEUSERVICO_ENABLE_DEBUG")

        cm := NewConfigManager()
        err := cm.LoadConfig()
        if err != nil {
            t.Fatalf("Expected no error, got: %v", err)
        }

        config := cm.GetConfig()
        if config.Timeout != 30*time.Second {
            t.Errorf("Expected default timeout to be 30s, got: %s", config.Timeout)
        }
        if config.MaxRetries != 3 {
            t.Errorf("Expected default max retries to be 3, got: %d", config.MaxRetries)
        }
        if config.EnableDebug {
            t.Error("Expected debug to be disabled by default")
        }
    })
}

func TestConfigManager_Validate(t *testing.T) {
    t.Run("Valid config", func(t *testing.T) {
        cm := NewConfigManager()
        cm.SetConfig(&Config{
            Endpoint:    "https://api.example.com",
            Timeout:     30 * time.Second,
            MaxRetries:  3,
            APIKey:      "valid-api-key",
            EnableDebug: false,
        })

        err := cm.Validate()
        if err != nil {
            t.Fatalf("Expected no error, got: %v", err)
        }
    })

    t.Run("Invalid endpoint", func(t *testing.T) {
        cm := NewConfigManager()
        cm.SetConfig(&Config{
            Endpoint:    "invalid-url",
            Timeout:     30 * time.Second,
            MaxRetries:  3,
            APIKey:      "valid-api-key",
            EnableDebug: false,
        })

        err := cm.Validate()
        if err == nil {
            t.Fatal("Expected error for invalid endpoint")
        }
    })

    t.Run("Nil config", func(t *testing.T) {
        cm := NewConfigManager()

        err := cm.Validate()
        if err == nil {
            t.Fatal("Expected error for nil config")
        }
    })
}
```

### Passo 4: Integrar ao Manager Principal

Edite `internal/app/config/manager.go`:

```go
import (
    // ... outros imports ...
    "github.com/fabianoflorentino/mr-robot/internal/app/meuservico"
)

type Manager struct {
    // ... outros managers ...
    meuservicoManager *meuservico.ConfigManager
}

func NewManager() *Manager {
    return &Manager{
        // ... outros managers ...
        meuservicoManager: meuservico.NewConfigManager(),
    }
}

func (m *Manager) LoadConfiguration() error {
    // ... outras configurações ...

    // Load meuservico configuration
    if err := m.meuservicoManager.LoadConfig(); err != nil {
        return fmt.Errorf("failed to load meuservico configuration: %w", err)
    }

    return nil
}

func (m *Manager) ValidateConfiguration() error {
    // ... outras validações ...

    if err := m.meuservicoManager.Validate(); err != nil {
        return fmt.Errorf("invalid meuservico configuration: %w", err)
    }

    return nil
}

// GetMeuservicoConfig returns the meuservico configuration
func (m *Manager) GetMeuservicoConfig() *meuservico.Config {
    return m.meuservicoManager.GetConfig()
}

// GetMeuservicoManager returns the meuservico config manager
func (m *Manager) GetMeuservicoManager() *meuservico.ConfigManager {
    return m.meuservicoManager
}
```

### Passo 5: Adicionar Variáveis de Ambiente

Documente no `.env` ou no README:

```bash
# Meu Serviço Configuration
MEUSERVICO_ENDPOINT=https://api.meuservico.com/v1    # OBRIGATÓRIO
MEUSERVICO_API_KEY=your-secret-api-key               # OBRIGATÓRIO
MEUSERVICO_TIMEOUT=30s                               # Opcional (default: 30s)
MEUSERVICO_MAX_RETRIES=3                             # Opcional (default: 3)
MEUSERVICO_ENABLE_DEBUG=false                        # Opcional (default: false)
```

### Passo 6: Testes de Integração

Adicione ao `internal/app/config/manager_test.go`:

```go
t.Run("Load all configurations including meuservico", func(t *testing.T) {
    // Set required environment variables
    os.Setenv("DEFAULT_PROCESSOR_URL", "http://default.example.com")
    os.Setenv("FALLBACK_PROCESSOR_URL", "http://fallback.example.com")
    os.Setenv("MEUSERVICO_ENDPOINT", "https://api.meuservico.com")
    os.Setenv("MEUSERVICO_API_KEY", "test-key")

    manager := NewManager()
    
    err := manager.LoadConfiguration()
    if err != nil {
        t.Fatalf("Failed to load configuration: %v", err)
    }

    err = manager.ValidateConfiguration()
    if err != nil {
        t.Fatalf("Failed to validate configuration: %v", err)
    }

    // Test meuservico config
    meuservicoConfig := manager.GetMeuservicoConfig()
    if meuservicoConfig == nil {
        t.Fatal("Meuservico config is nil")
    }
    if meuservicoConfig.Endpoint != "https://api.meuservico.com" {
        t.Errorf("Expected endpoint to be 'https://api.meuservico.com', got: %s", meuservicoConfig.Endpoint)
    }
})
```

## ✅ Checklist de Implementação

- [ ] Diretório criado em `internal/app/meuservico/`
- [ ] `config.go` implementado com Config, ConfigManager
- [ ] Função `LoadConfig()` implementada
- [ ] Função `Validate()` implementada
- [ ] `config_test.go` com testes completos
- [ ] Integração no `internal/app/config/manager.go`
- [ ] Variáveis de ambiente documentadas
- [ ] Testes de integração adicionados

## 🔐 Boas Práticas de Segurança

1. **Configurações Sensíveis**: Sempre marque como obrigatórias
2. **Validação Rigorosa**: Implemente validações específicas do domínio
3. **Defaults Seguros**: Use valores padrão conservadores
4. **Testes Completos**: Teste cenários de erro e valores inválidos
5. **Documentação Clara**: Documente todas as variáveis e seus formatos

## 🚨 Problemas Comuns

### Erro: "configuration not loaded"
**Solução**: Sempre chame `LoadConfig()` antes de `GetConfig()`

### Erro: "environment variable is required"
**Solução**: Verifique se todas as variáveis obrigatórias estão definidas

### Erro: "invalid value"
**Solução**: Verifique o formato dos valores nas variáveis de ambiente

### Testes falhando
**Solução**: Use o padrão de cleanup de variáveis de ambiente nos testes

## 📚 Referências

- [CONFIG_ARCHITECTURE.md](CONFIG_ARCHITECTURE.md) - Arquitetura geral
- [CONFIG_REFACTORING.md](CONFIG_REFACTORING.md) - Processo de refatoração
- [SECURITY_REFACTORING_SUMMARY.md](SECURITY_REFACTORING_SUMMARY.md) - Benefícios de segurança
