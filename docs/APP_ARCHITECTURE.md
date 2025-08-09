# Arquitetura do Diretório App - Guia de Manutenção

Este documento serve como guia para desenvolvedores que irão realizar manutenção e adicionar novas funcionalidades na aplicação mr-robot.

## 📋 Índice

- [Visão Geral](#visao-geral)
- [Estrutura do Diretório App](#estrutura-do-diretorio-app)
- [Fluxo de Inicialização](#fluxo-de-inicializacao)
- [Como Adicionar Nova Configuração](#como-adicionar-nova-configuracao)
- [Padrões e Convenções](#padroes-e-convencoes)
- [Testes](#testes)
- [Troubleshooting](#troubleshooting)

## 🎯 Visao Geral

O diretório `internal/app` implementa o padrão **Dependency Injection Container** e é responsável por:

- ⚙️ Gerenciamento de configurações
- 🗄️ Inicialização do banco de dados
- 🔧 Configuração de serviços
- 📊 Execução de migrações
- 🚦 Controle de ciclo de vida da aplicação

## 📁 Estrutura do Diretorio App

```text
internal/app/
├── container.go              # 🏗️ Container principal de DI
├── container_builder.go      # 🔨 Builder pattern para construção
├── container_test.go         # 🧪 Testes do container
├── interfaces.go            # 📝 Interfaces dos componentes
├── config/                  # ⚙️ Gerenciamento de configuração
│   └── manager.go
├── database/               # 🗄️ Gerenciamento de banco de dados
│   └── manager.go
├── services/              # 🔧 Gerenciamento de serviços
│   └── manager.go
├── migration/             # 📊 Gerenciamento de migrações
│   └── manager.go
├── queue/                # 📬 Sistema de filas
│   └── payment_queue.go
└── interfaces/           # 📋 Interfaces específicas
    └── payment_service.go
```

### 🧩 Componentes Principais

| Componente | Responsabilidade | Arquivo Principal |
|------------|------------------|-------------------|
| **Container** | Orquestração geral e DI | `container.go` |
| **Config Manager** | Carregamento de configurações | `config/manager.go` |
| **Database Manager** | Conexão com PostgreSQL | `database/manager.go` |
| **Services Manager** | Inicialização de serviços | `services/manager.go` |
| **Migration Manager** | Execução de migrações | `migration/manager.go` |

## 🔄 Fluxo de Inicializacao

O `AppContainer` segue uma sequência específica de inicialização:

```mermaid
graph TD
    A[1. Config Manager] --> B[2. Database Manager]
    B --> C[3. Services Manager]
    C --> D[4. Migration Manager]
    D --> E[Container Pronto]
```

### Sequência Detalhada

1. **📋 Configuração**: Carrega variáveis de ambiente
2. **🗄️ Banco de Dados**: Estabelece conexão com PostgreSQL
3. **🔧 Serviços**: Inicializa serviços de negócio (Payment, Queue)
4. **📊 Migrações**: Executa migrações pendentes

## ➕ Como Adicionar Nova Configuracao

### Passo 1: Definir Estrutura da Configuração

Edite `config/app_config.go`:

```go
// Adicione sua nova estrutura
type NovaConfig struct {
    Campo1 string
    Campo2 int
    Campo3 bool
    // Adicione campos conforme necessário
}

// Integre na AppConfig
type AppConfig struct {
    Database DatabaseConfig
    Payment  PaymentConfig
    Queue    QueueConfig
    Nova     NovaConfig  // ⬅️ Nova configuração aqui
}
```

### Passo 2: Implementar Carregamento de Variáveis

Na função `LoadAppConfig()`:

```go
func LoadAppConfig() (*AppConfig, error) {
    // ... código existente ...

    // Conversões com tratamento de erro
    campo2, err := strconv.Atoi(getEnvOrDefault("NOVA_CAMPO2", "42"))
    if err != nil {
        campo2 = 42
    }

    campo3, err := strconv.ParseBool(getEnvOrDefault("NOVA_CAMPO3", "false"))
    if err != nil {
        campo3 = false
    }

    return &AppConfig{
        // ... configurações existentes ...
        Nova: NovaConfig{
            Campo1: getEnvOrDefault("NOVA_CAMPO1", "valor_default"),
            Campo2: campo2,
            Campo3: campo3,
        },
    }, nil
}
```

### Passo 3: Criar Manager (se necessário)

Para componentes complexos, crie `internal/app/nova/manager.go`:

```go
package nova

import (
    "fmt"
    "github.com/fabianoflorentino/mr-robot/config"
)

type Manager struct {
    config *config.AppConfig
    // outros campos necessários
}

func NewManager(cfg *config.AppConfig) *Manager {
    return &Manager{
        config: cfg,
    }
}

func (n *Manager) Initialize() error {
    // 🚀 Lógica de inicialização
    fmt.Printf("Inicializando Nova com configuração: %+v\n", n.config.Nova)
    return nil
}

func (n *Manager) Shutdown() {
    // 🛑 Lógica de shutdown
    fmt.Println("Finalizando Nova...")
}

// Adicione métodos específicos do componente
func (n *Manager) GetSomeService() SomeServiceInterface {
    // implementação
}
```

### Passo 4: Integrar no Container

Modifique `container.go`:

```go
import (
    // ... imports existentes ...
    "github.com/fabianoflorentino/mr-robot/internal/app/nova"
)

type AppContainer struct {
    configManager    *config.Manager
    databaseManager  *database.Manager
    serviceManager   *appServices.Manager
    migrationManager *migration.Manager
    novaManager      *nova.Manager  // ⬅️ Novo manager
}

func NewAppContainer() (Container, error) {
    container := &AppContainer{}

    // Steps 1-4: inicializações existentes...

    // Step 5: Initialize nova manager
    container.novaManager = nova.NewManager(container.configManager.GetConfig())
    if err := container.novaManager.Initialize(); err != nil {
        return nil, fmt.Errorf("failed to initialize nova: %w", err)
    }

    return container, nil
}
```

### Passo 5: Atualizar Interface (se necessário)

Se outros componentes precisam acessar, atualize a interface:

```go
type Container interface {
    GetDB() *gorm.DB
    GetPaymentService() interfaces.PaymentServiceInterface
    GetPaymentQueue() *queue.PaymentQueue
    GetNovaManager() *nova.Manager  // ⬅️ Novo método
    Shutdown() error
}

// Implementar o método no AppContainer
func (c *AppContainer) GetNovaManager() *nova.Manager {
    return c.novaManager
}
```

### Passo 6: Atualizar Shutdown

No método `Shutdown()`:

```go
func (c *AppContainer) Shutdown() error {
    log.Println("Shutting down application container...")

    // Shutdown em ordem reversa da inicialização
    if c.novaManager != nil {
        log.Println("Shutting down nova...")
        c.novaManager.Shutdown()
    }

    // ... outros shutdowns existentes ...

    return nil
}
```

## 📏 Padroes e Convencoes

### ✅ Boas Práticas

- **🏗️ Manager Pattern**: Cada área tem seu próprio manager
- **🔄 Ordem de Inicialização**: Sempre seguir a sequência definida
- **❌ Tratamento de Erros**: Wrapping de erros com contexto
- **🧪 Testabilidade**: Interfaces para facilitar mocking
- **📝 Logging**: Log detalhado de inicialização e shutdown

### 📋 Convenções de Nomenclatura

| Tipo | Padrão | Exemplo |
|------|---------|---------|
| **Manager** | `{Area}Manager` | `ConfigManager`, `DatabaseManager` |
| **Config Struct** | `{Area}Config` | `PaymentConfig`, `QueueConfig` |
| **Env Variables** | `{AREA}_{CAMPO}` | `PAYMENT_URL`, `QUEUE_WORKERS` |
| **Interface** | `{Nome}Interface` | `PaymentServiceInterface` |

### 🏷️ Variaveis de Ambiente

```bash
# Exemplo de .env
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=secret
POSTGRES_DB=mr_robot

DEFAULT_PROCESSOR_URL=https://api.processor.com
FALLBACK_PROCESSOR_URL=https://fallback.processor.com

QUEUE_WORKERS=10
QUEUE_BUFFER_SIZE=10000
QUEUE_MAX_ENQUEUE_RETRIES=4

# Sua nova configuração
NOVA_CAMPO1=valor
NOVA_CAMPO2=42
NOVA_CAMPO3=true
```

## 🧪 Testes

### Testando Configurações

```go
func TestNovaConfig(t *testing.T) {
    // Setup
    os.Setenv("NOVA_CAMPO1", "test_value")
    os.Setenv("NOVA_CAMPO2", "100")
    defer os.Clearenv()

    // Act
    config, err := config.LoadAppConfig()

    // Assert
    assert.NoError(t, err)
    assert.Equal(t, "test_value", config.Nova.Campo1)
    assert.Equal(t, 100, config.Nova.Campo2)
}
```

### Testando Container

```go
func TestContainerWithNova(t *testing.T) {
    container, err := NewAppContainer()
    assert.NoError(t, err)
    assert.NotNil(t, container.GetNovaManager())

    defer container.Shutdown()
}
```

## 🔧 Troubleshooting

### Problemas Comuns

| Problema | Causa Provável | Solução |
|----------|----------------|---------|
| **Container falha na inicialização** | Ordem de dependências | Verificar sequência no `NewAppContainer()` |
| **Configuração não carrega** | Variável de ambiente inexistente | Verificar `.env` e valores default |
| **Panic no shutdown** | Manager nil | Adicionar verificação `if manager != nil` |
| **Testes falhando** | Configuração de teste | Usar `SetConfig()` no manager |

### Debug Útil

```go
// Adicionar logs para debug
log.Printf("Config loaded: %+v", config)
log.Printf("Manager initialized: %T", manager)
```

### Verificação de Saúde

```bash
# Verificar se todas as configurações estão carregadas
curl http://localhost:8080/health

# Verificar logs de inicialização
docker logs mr-robot-api
```

## 📞 Contato

Para dúvidas sobre a arquitetura ou sugestões de melhorias, abra uma issue no repositório ou entre em contato com a equipe de desenvolvimento.

---

**📝 Nota**: Este documento deve ser atualizado sempre que novos padrões ou componentes forem adicionados à arquitetura.
