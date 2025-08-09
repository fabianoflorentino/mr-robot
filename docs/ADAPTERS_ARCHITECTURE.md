# Arquitetura do Diretório Adapters - Guia de Manutenção

> **Consulte também**: [� ARCHITECTURE_GUIDE.md](ARCHITECTURE_GUIDE.md) para padrões gerais e convenções consolidadas.

Este documento foca especificamente no **diretório `adapters`** que implementa o padrão **Ports and Adapters** (Arquitetura Hexagonal).

## 🎯 Responsabilidades Específicas dos Adapters

- 📥 **Adaptadores de Entrada**: Controllers HTTP, mensageria, CLI
- 📤 **Adaptadores de Saída**: Repositórios, gateways externos, APIs
- 🔌 **Isolamento do Core**: Protege o domínio de detalhes técnicos
- 🔄 **Conversão de Dados**: Transforma dados entre formatos externos e internos
- 🛡️ **Validação de Entrada**: Sanitização e validação de dados externos

## 📁 Estrutura do Diretorio Adapters

```text
adapters/
├── inbound/                              # 📥 Adaptadores de entrada
│   └── http/                             # Protocolo HTTP
│       └── controllers/                  #  Controllers REST
│           ├── healthcheck_controller.go # Health check endpoint
│           └── payment_controller.go     # Endpoints de pagamento
└── outbound/                             # 📤 Adaptadores de saída
    ├── gateway/                          # Gateways externos
    │   ├── errors.go                     # Erros específicos de gateway
    │   ├── processor_factory.go          # Factory para processadores
    │   └── processor.go                  # Implementação de processador
    └── persistence/                      # Camada de persistência
        └── data/                         # Implementações de dados
            ├── payment_model.go          # Modelo de dados para DB
            └── payment_repository.go     # Implementação do repositório
```

### 🧩 Componentes Principais

| Componente | Responsabilidade | Arquivo Principal | Tipo |
|------------|------------------|-------------------|------|
| **Payment Controller** | Endpoints REST de pagamento | `inbound/http/controllers/payment_controller.go` | Inbound |
| **Healthcheck Controller** | Endpoint de health check | `inbound/http/controllers/healthcheck_controller.go` | Inbound |
| **Payment Repository** | Persistência de pagamentos | `outbound/persistence/data/payment_repository.go` | Outbound |
| **Process Gateway** | Gateway para processadores | `outbound/gateway/processor.go` | Outbound |
| **Processor Factory** | Factory para processadores | `outbound/gateway/processor_factory.go` | Outbound |

## 📥 Adaptadores Inbound

Os adaptadores inbound recebem requisições externas e as convertem para o domínio:

### HTTP Controllers

#### Payment Controller

```go
// Estrutura do controller
type PaymentController struct {
    paymentService interfaces.PaymentServiceInterface
    paymentQueue   *queue.PaymentQueue
}

// Endpoint principal
func (pc *PaymentController) ProcessPayment(c *gin.Context) {
    // 1. Bind e validação da requisição
    var payment domain.Payment
    if err := c.ShouldBindJSON(&payment); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    // 2. Enfileirar para processamento assíncrono
    if err := pc.paymentQueue.Enqueue(&payment); err != nil {
        c.JSON(500, gin.H{"error": "Failed to enqueue payment"})
        return
    }

    // 3. Resposta de aceite
    c.JSON(202, gin.H{"status": "accepted"})
}
```

**Responsabilidades dos Controllers:**

- ✅ **Validação HTTP**: Bind de JSON, query params, headers
- ✅ **Conversão**: Transform dados HTTP → domínio
- ✅ **Orquestração**: Chamar serviços do domínio
- ✅ **Resposta**: Formatar resposta HTTP adequada
- ✅ **Tratamento de Erros**: Converter erros de domínio → HTTP status

## 📤 Adaptadores Outbound

Os adaptadores outbound implementam interfaces do core para acessar recursos externos:

### Gateway (Processadores Externos)

```go
type ProcessGateway struct {
    URL  string
    Name string
}

func (pg *ProcessGateway) Process(payment *domain.Payment) (bool, error) {
    // 1. Preparar requisição HTTP
    reqData := map[string]interface{}{
        "correlationId": payment.CorrelationID,
        "amount":        payment.Amount,
    }

    // 2. Fazer chamada HTTP
    resp, err := http.Post(pg.URL, "application/json", body)
    if err != nil {
        return false, fmt.Errorf("failed to call processor: %w", err)
    }

    // 3. Processar resposta
    return resp.StatusCode == 200, nil
}
```

### Persistence (Repositórios)

```go
type DataPaymentRepository struct {
    db *gorm.DB
}

func (r *DataPaymentRepository) Process(ctx context.Context, payment *domain.Payment, processorName string) error {
    // 1. Converter domínio → modelo de dados
    model := &PaymentModel{
        CorrelationID: payment.CorrelationID,
        Amount:        payment.Amount,
        ProcessorName: processorName,
        ProcessedAt:   time.Now(),
    }

    // 2. Persistir no banco
    if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
        return fmt.Errorf("failed to save payment: %w", err)
    }

    return nil
}
```

## ➕ Como Adicionar Novo Controller

### Passo 1: Definir Estrutura do Controller

Crie `adapters/inbound/http/controllers/novo_controller.go`:

```go
package controllers

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/fabianoflorentino/mr-robot/core/domain"
    "github.com/fabianoflorentino/mr-robot/internal/app/interfaces"
)

type NovoController struct {
    novoService interfaces.NovoServiceInterface
}

func NewNovoController(service interfaces.NovoServiceInterface) *NovoController {
    return &NovoController{
        novoService: service,
    }
}

// Registrar rotas do controller
func (nc *NovoController) RegisterRoutes(router *gin.Engine) {
    v1 := router.Group("/api/v1")
    {
        v1.POST("/nova-entidade", nc.CriarEntidade)
        v1.GET("/nova-entidade/:id", nc.BuscarEntidade)
        v1.PUT("/nova-entidade/:id", nc.AtualizarEntidade)
        v1.DELETE("/nova-entidade/:id", nc.DeletarEntidade)
    }
}
```

### Passo 2: Implementar Endpoints

```go
func (nc *NovoController) CriarEntidade(c *gin.Context) {
    // 1. Bind da requisição
    var req CriarEntidadeRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "Invalid request format",
            "details": err.Error(),
        })
        return
    }

    // 2. Validação adicional se necessário
    if err := nc.validarRequisicao(&req); err != nil {
        c.JSON(http.StatusUnprocessableEntity, gin.H{
            "error": "Validation failed",
            "details": err.Error(),
        })
        return
    }

    // 3. Converter para domínio
    entidade := &domain.NovaEntidade{
        ID:     uuid.New(),
        Nome:   req.Nome,
        Status: req.Status,
        Valor:  req.Valor,
    }

    // 4. Chamar serviço de domínio
    if err := nc.novoService.CriarEntidade(c.Request.Context(), entidade); err != nil {
        // Converter erro de domínio para HTTP
        status, message := nc.mapearErro(err)
        c.JSON(status, gin.H{"error": message})
        return
    }

    // 5. Resposta de sucesso
    c.JSON(http.StatusCreated, gin.H{
        "id":      entidade.ID,
        "status":  "created",
        "message": "Entidade criada com sucesso",
    })
}

func (nc *NovoController) mapearErro(err error) (int, string) {
    switch {
    case errors.Is(err, domain.ErrEntidadeInvalida):
        return http.StatusBadRequest, "Dados da entidade são inválidos"
    case errors.Is(err, domain.ErrEntidadeJaExiste):
        return http.StatusConflict, "Entidade já existe"
    case errors.Is(err, domain.ErrServicoIndisponivel):
        return http.StatusServiceUnavailable, "Serviço temporariamente indisponível"
    default:
        return http.StatusInternalServerError, "Erro interno do servidor"
    }
}
```

### Passo 3: Definir DTOs

```go
// Estruturas para requisição/resposta
type CriarEntidadeRequest struct {
    Nome   string  `json:"nome" binding:"required,min=3,max=100"`
    Status string  `json:"status" binding:"required,oneof=ativo inativo"`
    Valor  float64 `json:"valor" binding:"min=0"`
}

type EntidadeResponse struct {
    ID     uuid.UUID `json:"id"`
    Nome   string    `json:"nome"`
    Status string    `json:"status"`
    Valor  float64   `json:"valor"`
    CriadoEm time.Time `json:"criadoEm"`
}
```

## 🔌 Como Adicionar Novo Gateway

### Passo 1: Implementar Interface do Domínio

Crie `adapters/outbound/gateway/novo_gateway.go`:

```go
package gateway

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "time"

    "github.com/fabianoflorentino/mr-robot/core/domain"
)

type NovoGateway struct {
    URL     string
    Timeout time.Duration
    client  *http.Client
}

func NewNovoGateway(url string) *NovoGateway {
    return &NovoGateway{
        URL:     url,
        Timeout: 30 * time.Second,
        client: &http.Client{
            Timeout: 30 * time.Second,
        },
    }
}

// Implementar interface do domínio
func (g *NovoGateway) ProcessarEntidade(entidade *domain.NovaEntidade) (bool, error) {
    // 1. Preparar payload
    payload := map[string]interface{}{
        "id":     entidade.ID,
        "nome":   entidade.Nome,
        "status": entidade.Status,
        "valor":  entidade.Valor,
    }

    jsonData, err := json.Marshal(payload)
    if err != nil {
        return false, fmt.Errorf("failed to marshal payload: %w", err)
    }

    // 2. Fazer requisição HTTP
    req, err := http.NewRequest("POST", g.URL, bytes.NewBuffer(jsonData))
    if err != nil {
        return false, fmt.Errorf("failed to create request: %w", err)
    }

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("User-Agent", "mr-robot/1.0")

    // 3. Executar requisição
    resp, err := g.client.Do(req)
    if err != nil {
        return false, fmt.Errorf("failed to execute request: %w", err)
    }
    defer resp.Body.Close()

    // 4. Processar resposta
    switch resp.StatusCode {
    case http.StatusOK, http.StatusCreated:
        return true, nil
    case http.StatusBadRequest:
        return false, fmt.Errorf("invalid request data")
    case http.StatusServiceUnavailable:
        return false, fmt.Errorf("service unavailable")
    default:
        return false, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
    }
}

func (g *NovoGateway) GatewayName() string {
    return "novo-gateway"
}
```

### Passo 2: Adicionar Factory (se necessário)

```go
// Em processor_factory.go ou criar novo arquivo
func CreateNovoGateway(config *config.AppConfig) domain.NovaEntidadeProcessor {
    return NewNovoGateway(config.NovoGateway.URL)
}
```

## 📏 Padroes e Convencoes

### ✅ Boas Práticas dos Adaptadores

- **🔄 Conversão Explícita**: Sempre converter entre formatos externos e domínio
- **❌ Tratamento de Erros**: Mapear erros específicos para respostas adequadas
- **⏱️ Timeouts**: Configurar timeouts apropriados para chamadas externas
- **📝 Logging**: Log detalhado de operações externas
- **🧪 Testabilidade**: Interfaces mockáveis para testes

### 📋 Convenções de Nomenclatura

| Tipo | Padrão | Exemplo |
|------|---------|---------|
| **Controller** | `{Entidade}Controller` | `PaymentController`, `UserController` |
| **Gateway** | `{Nome}Gateway` | `PaymentGateway`, `NotificationGateway` |
| **Repository** | `Data{Entidade}Repository` | `DataPaymentRepository`, `DataUserRepository` |
| **DTO Request** | `{Acao}{Entidade}Request` | `CreatePaymentRequest`, `UpdateUserRequest` |
| **DTO Response** | `{Entidade}Response` | `PaymentResponse`, `UserResponse` |

### 📊 Estrutura de Respostas HTTP

```go
// Sucesso
{
    "data": { ... },
    "status": "success",
    "message": "Operation completed successfully"
}

// Erro
{
    "error": "Brief error description",
    "details": "Detailed error information",
    "code": "ERROR_CODE",
    "timestamp": "2024-01-01T12:00:00Z"
}

// Lista com paginação
{
    "data": [ ... ],
    "pagination": {
        "page": 1,
        "limit": 10,
        "total": 100,
        "totalPages": 10
    }
}
```

## 🧪 Testes

### Testando Controllers

```go
func TestPaymentController_ProcessPayment(t *testing.T) {
    // Setup
    mockService := &MockPaymentService{}
    mockQueue := &MockPaymentQueue{}
    controller := NewPaymentController(mockService, mockQueue)

    router := gin.New()
    controller.RegisterRoutes(router)

    // Test data
    payment := domain.Payment{
        CorrelationID: uuid.New(),
        Amount:        100.50,
    }

    // Mock expectations
    mockQueue.On("Enqueue", &payment).Return(nil)

    // Prepare request
    jsonData, _ := json.Marshal(payment)
    req := httptest.NewRequest("POST", "/payments", bytes.NewBuffer(jsonData))
    req.Header.Set("Content-Type", "application/json")

    w := httptest.NewRecorder()

    // Act
    router.ServeHTTP(w, req)

    // Assert
    assert.Equal(t, http.StatusAccepted, w.Code)

    var response map[string]string
    json.Unmarshal(w.Body.Bytes(), &response)
    assert.Equal(t, "accepted", response["status"])

    mockQueue.AssertExpectations(t)
}
```

### Testando Gateways

```go
func TestProcessGateway_Process(t *testing.T) {
    // Setup mock server
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        assert.Equal(t, "POST", r.Method)
        assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"status": "success"}`))
    }))
    defer server.Close()

    // Create gateway
    gateway := NewProcessGateway(server.URL, "test-processor")

    // Test data
    payment := &domain.Payment{
        CorrelationID: uuid.New(),
        Amount:        100.50,
    }

    // Act
    success, err := gateway.Process(payment)

    // Assert
    assert.NoError(t, err)
    assert.True(t, success)
    assert.Equal(t, "test-processor", gateway.ProcessorName())
}
```

### Testando Repositórios

```go
func TestDataPaymentRepository_Process(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)

    repo := NewDataPaymentRepository(db)

    // Test data
    payment := &domain.Payment{
        CorrelationID: uuid.New(),
        Amount:        100.50,
    }

    // Act
    err := repo.Process(context.Background(), payment, "test-processor")

    // Assert
    assert.NoError(t, err)

    // Verify persistence
    var model PaymentModel
    err = db.Where("correlation_id = ?", payment.CorrelationID).First(&model).Error
    assert.NoError(t, err)
    assert.Equal(t, payment.Amount, model.Amount)
    assert.Equal(t, "test-processor", model.ProcessorName)
}
```

---

**📝 Nota**: Para padrões gerais, convenções de nomenclatura e troubleshooting consolidado, consulte o [📖 ARCHITECTURE_GUIDE.md](ARCHITECTURE_GUIDE.md).
