# Fluxogramas Atualizados mr-robot - Agosto 2025

> **Data de Atualização**: 16 de Agosto de 2025  
> **Versão**: v0.0.4  
> **Baseado em**: Análise do código atual  

## 🔄 Fluxograma de Arquitetura Principal (Atualizado)

```mermaid
flowchart TD
    %% Estilos modernizados
    classDef entrypoint fill:#e1f5fe,stroke:#01579b,stroke-width:3px,color:#000
    classDef internal fill:#f1f8e9,stroke:#33691e,stroke-width:2px,color:#000
    classDef inbound fill:#f3e5f5,stroke:#4a148c,stroke-width:2px,color:#000
    classDef core fill:#e8f5e8,stroke:#1b5e20,stroke-width:2px,color:#000
    classDef outbound fill:#fff3e0,stroke:#e65100,stroke-width:2px,color:#000
    classDef infra fill:#fce4ec,stroke:#880e4f,stroke-width:2px,color:#000
    classDef async fill:#e3f2fd,stroke:#0277bd,stroke-width:2px,color:#000
    classDef config fill:#f9fbe7,stroke:#827717,stroke-width:2px,color:#000

    %% Componentes principais
    A["🚀 main.go<br/>Entry Point"] --> B["📦 Container DI<br/>AppContainer"]

    B --> C["🌐 HTTP Server<br/>Native HTTP"]
    B --> Q["⚡ Payment Queue<br/>Async Workers"]
    B --> K["⚙️ Config Managers<br/>Modular System"]

    %% Fluxo HTTP
    C --> D["🎯 Payment Controller<br/>HTTP Endpoints"]
    D --> Q

    %% Sistema de configuração
    K --> CM1["🗄️ Database Config<br/>Manager"]
    K --> CM2["💳 Payment Config<br/>Manager"]
    K --> CM3["⚡ Circuit Breaker Config<br/>Manager"]
    K --> CM4["📬 Queue Config<br/>Manager"]

    %% Processamento assíncrono
    Q --> E["💼 Payment Service<br/>With Fallback"]

    %% Proteções no Service
    E --> CB1["🛡️ Default Circuit Breaker<br/>Independent Protection"]
    E --> CB2["🛡️ Fallback Circuit Breaker<br/>Independent Protection"]
    E --> RL["⏱️ Rate Limiter<br/>Concurrency Control"]

    %% Core Domain
    CB1 --> F["📋 Payment Repository<br/>Interface"]
    CB2 --> F
    RL --> F

    %% Persistência
    F --> G["💾 Payment Repository Impl<br/>SQL + Transactions + Retry"]
    G --> H["🐘 PostgreSQL<br/>Database"]

    %% Processadores com Fallback
    CB1 --> I1["🏦 Default Processor<br/>Primary Gateway"]
    CB2 --> I2["🔄 Fallback Processor<br/>Secondary Gateway"]
    I1 -.->|"Auto Fallback on Failure"| I2

    %% Unix Sockets (opcional)
    US["📁 Unix Sockets<br/>/var/run/mr_robot/"] -.-> C
    HAP["⚖️ HAProxy<br/>Load Balancer"] -.-> US

    %% Agrupamentos de camadas
    subgraph "🚀 Entry Point"
        A
    end

    subgraph "🔧 Internal Layer"
        B
        C
        K
    end

    subgraph "⚙️ Configuration System"
        CM1
        CM2
        CM3
        CM4
    end

    subgraph "📥 Inbound Adapters"
        D
    end

    subgraph "⚡ Queue System"
        Q
    end

    subgraph "💚 Core Domain"
        E
        F
        CB1
        CB2
        RL
    end

    subgraph "📤 Outbound Adapters"
        G
        I1
        I2
    end

    subgraph "🏗️ Infrastructure"
        H
        US
        HAP
    end

    %% Aplicando estilos
    class A entrypoint
    class B,C,K internal
    class CM1,CM2,CM3,CM4 config
    class D inbound
    class Q async
    class E,F,CB1,CB2,RL core
    class G,I1,I2 outbound
    class H,US,HAP infra

    %% Setas com rótulos
    C -.->|"HTTP Request"| D
    D -.->|"Enqueue Job"| Q
    Q -.->|"Async Processing"| E
    E -.->|"Try Default First"| CB1
    E -.->|"Fallback if Default Fails"| CB2
    E -.->|"Rate Control"| RL
    F -.->|"Persist with Processor Name"| G
    G -.->|"SQL Transactions + Retry"| H
    CB1 -.->|"Process Payment"| I1
    CB2 -.->|"Process Payment"| I2
```

## 🔧 Fluxograma de Configuração (Detalhado)

```mermaid
flowchart TD
    %% Estilos
    classDef env fill:#fff3e0,stroke:#e65100,stroke-width:2px,color:#000
    classDef manager fill:#e8f5e8,stroke:#1b5e20,stroke-width:2px,color:#000
    classDef config fill:#f3e5f5,stroke:#4a148c,stroke-width:2px,color:#000
    classDef validation fill:#ffebee,stroke:#c62828,stroke-width:2px,color:#000

    %% Variáveis de ambiente
    ENV1["🌍 POSTGRES_*<br/>Database Variables"]
    ENV2["🌍 DEFAULT_PROCESSOR_URL<br/>FALLBACK_PROCESSOR_URL"]
    ENV3["🌍 CIRCUIT_BREAKER_*<br/>Protection Variables"]
    ENV4["🌍 QUEUE_*<br/>Queue Variables"]

    %% Managers
    ENV1 --> DBM["🗄️ Database ConfigManager<br/>LoadConfig()"]
    ENV2 --> PM["💳 Payment ConfigManager<br/>LoadConfig()"]
    ENV3 --> CBM["⚡ CircuitBreaker ConfigManager<br/>LoadConfig()"]
    ENV4 --> QM["📬 Queue ConfigManager<br/>LoadConfig()"]

    %% Validação
    DBM --> DBV{"🔍 Validate DB Config"}
    PM --> PMV{"🔍 Validate Payment Config"}
    CBM --> CBMV{"🔍 Validate CB Config"}
    QM --> QMV{"🔍 Validate Queue Config"}

    %% Configurações finais
    DBV -->|"✅ Valid"| DBC["📊 DatabaseConfig"]
    PMV -->|"✅ Valid"| PMC["📊 PaymentConfig"]
    CBMV -->|"✅ Valid"| CBC["📊 CircuitBreakerConfig"]
    QMV -->|"✅ Valid"| QC["📊 QueueConfig"]

    %% Erros
    DBV -->|"❌ Invalid"| ERR["💥 Configuration Error"]
    PMV -->|"❌ Invalid"| ERR
    CBMV -->|"❌ Invalid"| ERR
    QMV -->|"❌ Invalid"| ERR

    %% Container Integration
    DBC --> CONT["📦 AppContainer<br/>Dependency Injection"]
    PMC --> CONT
    CBC --> CONT
    QC --> CONT

    %% Aplicando estilos
    class ENV1,ENV2,ENV3,ENV4 env
    class DBM,PM,CBM,QM manager
    class DBC,PMC,CBC,QC config
    class DBV,PMV,CBMV,QMV,ERR validation
    class CONT manager
```

## 🚀 Fluxograma de Processamento de Pagamento (Detalhado)

```mermaid
flowchart TD
    %% Estilos
    classDef request fill:#e3f2fd,stroke:#0277bd,stroke-width:2px,color:#000
    classDef queue fill:#f1f8e9,stroke:#33691e,stroke-width:2px,color:#000
    classDef protection fill:#fff3e0,stroke:#e65100,stroke-width:2px,color:#000
    classDef processor fill:#f3e5f5,stroke:#4a148c,stroke-width:2px,color:#000
    classDef storage fill:#e8f5e8,stroke:#1b5e20,stroke-width:2px,color:#000
    classDef success fill:#e8f5e8,stroke:#2e7d32,stroke-width:2px,color:#000
    classDef failure fill:#ffebee,stroke:#c62828,stroke-width:2px,color:#000

    %% Fluxo principal
    REQ["📥 HTTP POST /payments<br/>correlationId + amount"] --> VAL{"🔍 Validate Request"}
    VAL -->|"✅ Valid"| ENQ["📬 Enqueue Payment Job"]
    VAL -->|"❌ Invalid"| ERR1["❌ 400 Bad Request"]

    ENQ --> TIMEOUT{"⏱️ Enqueue Timeout<br/>5 seconds"}
    TIMEOUT -->|"✅ Enqueued"| OK1["✅ 200 OK Response"]
    TIMEOUT -->|"❌ Timeout"| ERR2["❌ 408 Request Timeout"]
    TIMEOUT -->|"❌ Queue Full"| ERR3["❌ 429 Too Many Requests"]

    %% Processamento assíncrono
    ENQ --> WORKER["👷 Queue Worker<br/>Background Processing"]
    WORKER --> ACQUIRE{"🎯 Rate Limiter<br/>Acquire Slot"}
    ACQUIRE -->|"✅ Acquired"| PROCESS
    ACQUIRE -->|"❌ Context Timeout"| RETRY1

    %% Processamento com proteções
    PROCESS["💼 Payment Service<br/>Process Payment"] --> CBDEF{"🛡️ Default Circuit Breaker<br/>State Check"}
    
    CBDEF -->|"🟢 Closed/Half-Open"| TRYDEF["🏦 Try Default Processor"]
    CBDEF -->|"🔴 Open"| TRYFALL["🔄 Try Fallback Processor"]

    TRYDEF --> DEFRESULT{"🎯 Default Result"}
    DEFRESULT -->|"✅ Success"| SAVEDEF["💾 Save to DB<br/>processor='default'"]
    DEFRESULT -->|"❌ Failure"| CBFALL{"🛡️ Fallback Circuit Breaker<br/>State Check"}

    CBFALL -->|"🟢 Closed/Half-Open"| TRYFALL
    CBFALL -->|"🔴 Open"| FAIL["❌ Both Processors Failed"]

    TRYFALL --> FALLRESULT{"🎯 Fallback Result"}
    FALLRESULT -->|"✅ Success"| SAVEFALL["💾 Save to DB<br/>processor='fallback'"]
    FALLRESULT -->|"❌ Failure"| FAIL

    %% Persistência com retry
    SAVEDEF --> DBTX{"🗄️ Database Transaction<br/>with Retry"}
    SAVEFALL --> DBTX
    DBTX -->|"✅ Success"| SUCCESS["✅ Payment Processed"]
    DBTX -->|"❌ Deadlock"| DBRETRY["🔄 Exponential Backoff<br/>100ms, 400ms, 900ms"]
    DBRETRY --> DBTX
    DBTX -->|"❌ Max Retries"| DBFAIL["❌ Database Error"]

    %% Retry do job
    FAIL --> RETRY1["🔄 Job Retry<br/>Exponential Backoff"]
    DBFAIL --> RETRY1
    RETRY1 --> RETRYCHECK{"🔍 Retry Count<br/>< Max Retries?"}
    RETRYCHECK -->|"✅ Retry"| DELAY["⏱️ Backoff Delay<br/>1s, 2s, 4s, 8s"]
    RETRYCHECK -->|"❌ Max Reached"| ABANDON["💀 Abandon Job"]
    DELAY --> WORKER

    %% Aplicando estilos
    class REQ,OK1 request
    class ENQ,WORKER,DELAY queue
    class ACQUIRE,CBDEF,CBFALL,DBTX protection
    class TRYDEF,TRYFALL,DEFRESULT,FALLRESULT processor
    class SAVEDEF,SAVEFALL,SUCCESS storage
    class ERR1,ERR2,ERR3,FAIL,DBFAIL,ABANDON failure
    class VAL,TIMEOUT,PROCESS,RETRY1,RETRYCHECK success
```

## 🔌 Fluxograma de Unix Sockets (Detalhado)

```mermaid
flowchart TD
    %% Estilos
    classDef external fill:#e3f2fd,stroke:#0277bd,stroke-width:2px,color:#000
    classDef haproxy fill:#fff3e0,stroke:#e65100,stroke-width:2px,color:#000
    classDef socket fill:#f3e5f5,stroke:#4a148c,stroke-width:2px,color:#000
    classDef app fill:#e8f5e8,stroke:#1b5e20,stroke-width:2px,color:#000
    classDef config fill:#f1f8e9,stroke:#33691e,stroke-width:2px,color:#000

    %% Cliente externo
    CLIENT["🌍 External Client<br/>HTTP Request"]

    %% HAProxy
    HAP["⚖️ HAProxy<br/>Port 9999<br/>Round Robin LB"]

    %% Configuração de socket
    CONFIG{"⚙️ USE_UNIX_SOCKET<br/>Environment Variable"}
    CONFIG -->|"true"| UNIX_MODE["🔌 Unix Socket Mode"]
    CONFIG -->|"false"| TCP_MODE["🌐 TCP Mode<br/>Port 8888"]

    %% Unix Socket Mode
    UNIX_MODE --> SOCKET_CHECK{"📁 Socket Directory<br/>/var/run/mr_robot/"}
    SOCKET_CHECK -->|"✅ Exists"| CREATE_SOCKETS
    SOCKET_CHECK -->|"❌ Missing"| CREATE_DIR["📁 Create Directory<br/>mkdir -p /var/run/mr_robot"]
    CREATE_DIR --> CREATE_SOCKETS

    CREATE_SOCKETS["📄 Create Socket Files<br/>mr_robot1.sock, mr_robot2.sock"]

    %% Instâncias da aplicação
    CREATE_SOCKETS --> APP1["📱 App Instance 1<br/>Bind to mr_robot1.sock"]
    CREATE_SOCKETS --> APP2["📱 App Instance 2<br/>Bind to mr_robot2.sock"]

    %% HAProxy backend config
    HAP --> BACKEND{"⚙️ HAProxy Backend<br/>Configuration"}
    BACKEND -->|"Unix Socket Mode"| UNIX_BACKEND["🔌 Backend Unix<br/>server mr_robot1 /var/run/mr_robot/mr_robot1.sock<br/>server mr_robot2 /var/run/mr_robot/mr_robot2.sock"]
    BACKEND -->|"TCP Mode"| TCP_BACKEND["🌐 Backend TCP<br/>server mr_robot1 localhost:8888<br/>server mr_robot2 localhost:8889"]

    %% Fluxo de requisição
    CLIENT --> HAP
    HAP --> UNIX_BACKEND
    HAP --> TCP_BACKEND
    UNIX_BACKEND -.->|"Load Balance"| APP1
    UNIX_BACKEND -.->|"Load Balance"| APP2
    TCP_BACKEND -.->|"Load Balance"| TCP_APP1["📱 App Instance 1<br/>TCP Port 8888"]
    TCP_BACKEND -.->|"Load Balance"| TCP_APP2["📱 App Instance 2<br/>TCP Port 8889"]

    %% Health checks
    APP1 --> HEALTH1["❤️ Health Check<br/>GET /health"]
    APP2 --> HEALTH2["❤️ Health Check<br/>GET /health"]
    TCP_APP1 --> HEALTH3["❤️ Health Check<br/>GET /health"]
    TCP_APP2 --> HEALTH4["❤️ Health Check<br/>GET /health"]

    %% Performance comparison
    UNIX_BACKEND -.->|"⚡ 20% faster<br/>🔒 More secure"| PERF_UNIX["📊 Unix Socket Performance"]
    TCP_BACKEND -.->|"🌐 Standard<br/>🔌 Port-based"| PERF_TCP["📊 TCP Performance"]

    %% Aplicando estilos
    class CLIENT external
    class HAP,BACKEND,UNIX_BACKEND,TCP_BACKEND haproxy
    class CREATE_SOCKETS,UNIX_MODE,SOCKET_CHECK,CREATE_DIR socket
    class APP1,APP2,TCP_APP1,TCP_APP2 app
    class CONFIG,TCP_MODE,HEALTH1,HEALTH2,HEALTH3,HEALTH4,PERF_UNIX,PERF_TCP config
```

## 📊 Fluxograma de Monitoramento e Métricas

```mermaid
flowchart TD
    %% Estilos
    classDef endpoint fill:#e3f2fd,stroke:#0277bd,stroke-width:2px,color:#000
    classDef metric fill:#e8f5e8,stroke:#1b5e20,stroke-width:2px,color:#000
    classDef aggregation fill:#fff3e0,stroke:#e65100,stroke-width:2px,color:#000
    classDef response fill:#f3e5f5,stroke:#4a148c,stroke-width:2px,color:#000

    %% Endpoints de monitoramento
    HEALTH["🏥 GET /health<br/>Application Health"]
    SUMMARY["📊 GET /payment-summary<br/>Payment Statistics"]
    PURGE["🗑️ DELETE /payments<br/>Purge All Data"]

    %% Health check components
    HEALTH --> HC_APP{"🔍 App Status"}
    HEALTH --> HC_DB{"🔍 Database Status"}
    HEALTH --> HC_QUEUE{"🔍 Queue Status"}
    HEALTH --> HC_CB{"🔍 Circuit Breakers"}

    HC_APP -->|"✅ OK"| H_OK["✅ 200 OK"]
    HC_DB -->|"✅ Connected"| H_OK
    HC_QUEUE -->|"✅ Running"| H_OK
    HC_CB -->|"🟢 Monitoring"| H_OK

    HC_APP -->|"❌ Error"| H_ERR["❌ 503 Service Unavailable"]
    HC_DB -->|"❌ Disconnected"| H_ERR
    HC_QUEUE -->|"❌ Stopped"| H_ERR
    HC_CB -->|"🔴 All Open"| H_ERR

    %% Payment summary components
    SUMMARY --> QUERY_DB["🗄️ Query Database<br/>Aggregate by Processor"]
    QUERY_DB --> AGG_DEFAULT["📈 Aggregate Default<br/>COUNT(*), SUM(amount)"]
    QUERY_DB --> AGG_FALLBACK["📈 Aggregate Fallback<br/>COUNT(*), SUM(amount)"]

    AGG_DEFAULT --> METRICS_DEF["📊 Default Metrics<br/>totalRequests, totalAmount"]
    AGG_FALLBACK --> METRICS_FALL["📊 Fallback Metrics<br/>totalRequests, totalAmount"]

    METRICS_DEF --> JSON_RESP["📋 JSON Response<br/>{default: {...}, fallback: {...}}"]
    METRICS_FALL --> JSON_RESP

    %% Time filtering
    SUMMARY --> TIME_FILTER{"🕐 Time Range Provided?"}
    TIME_FILTER -->|"✅ from & to"| FILTERED_QUERY["🗄️ Filtered Query<br/>WHERE created_at BETWEEN"]
    TIME_FILTER -->|"❌ No filter"| QUERY_DB
    FILTERED_QUERY --> AGG_DEFAULT
    FILTERED_QUERY --> AGG_FALLBACK

    %% Purge operation
    PURGE --> PURGE_AUTH{"🔐 Authorization Check"}
    PURGE_AUTH -->|"✅ Authorized"| PURGE_DB["🗑️ DELETE FROM payments"]
    PURGE_AUTH -->|"❌ Unauthorized"| PURGE_ERR["❌ 401 Unauthorized"]

    PURGE_DB --> PURGE_OK["✅ 204 No Content"]

    %% Circuit breaker metrics
    HC_CB --> CB_DEFAULT["🛡️ Default CB State<br/>Closed/Open/HalfOpen"]
    HC_CB --> CB_FALLBACK["🛡️ Fallback CB State<br/>Closed/Open/HalfOpen"]
    HC_CB --> CB_FAILURES["📊 Failure Counts<br/>Per Processor"]

    %% Queue metrics
    HC_QUEUE --> Q_WORKERS["👷 Active Workers<br/>Count"]
    HC_QUEUE --> Q_JOBS["📬 Queue Size<br/>Pending Jobs"]
    HC_QUEUE --> Q_PROCESSED["✅ Jobs Processed<br/>Total Count"]
    HC_QUEUE --> Q_FAILED["❌ Jobs Failed<br/>Total Count"]

    %% Aplicando estilos
    class HEALTH,SUMMARY,PURGE endpoint
    class METRICS_DEF,METRICS_FALL,CB_DEFAULT,CB_FALLBACK,CB_FAILURES,Q_WORKERS,Q_JOBS,Q_PROCESSED,Q_FAILED metric
    class QUERY_DB,AGG_DEFAULT,AGG_FALLBACK,FILTERED_QUERY,PURGE_DB aggregation
    class JSON_RESP,H_OK,H_ERR,PURGE_OK,PURGE_ERR response
```

## 🎯 Legenda dos Fluxogramas

### 🎨 Códigos de Cores

| Cor | Componente | Descrição |
|-----|------------|-----------|
| 🔵 **Azul** | Entry Point/External | Pontos de entrada e clientes externos |
| 🟢 **Verde** | Core/Internal | Lógica de negócio e componentes internos |
| 🟡 **Amarelo** | Configuration | Sistema de configuração e managers |
| 🟠 **Laranja** | Outbound/Processing | Adaptadores de saída e processamento |
| 🟣 **Roxo** | Inbound/Controllers | Adaptadores de entrada e controladores |
| 🔴 **Vermelho** | Infrastructure | Infraestrutura externa (DB, HAProxy) |
| ⚪ **Cinza** | Queue/Async | Sistema de filas e processamento assíncrono |

### 📊 Símbolos

| Símbolo | Significado |
|---------|-------------|
| `-->` | Fluxo síncrono |
| `-.->` | Fluxo assíncrono ou opcional |
| `{}` | Decisão/Condição |
| `[]` | Processo/Ação |
| `()` | Dados/Estado |

### 🚀 Estados do Circuit Breaker

| Estado | Cor | Descrição |
|--------|-----|-----------|
| 🟢 **Closed** | Verde | Funcionando normalmente |
| 🔴 **Open** | Vermelho | Bloqueando requisições |
| 🟡 **Half-Open** | Amarelo | Testando recuperação |

---

**📝 Nota**: Estes fluxogramas foram atualizados baseados na análise do código atual (v0.0.4) e refletem com precisão a implementação real do sistema.
