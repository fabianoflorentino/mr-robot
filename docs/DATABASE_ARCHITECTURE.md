# Arquitetura do Diretório Database - Guia de Manutenção

Este documento serve como guia para desenvolvedores que irão realizar manutenção e modificações na camada de banco de dados da aplicação mr-robot.

## 📋 Índice

- [Visao Geral](#visao-geral)
- [Estrutura do Diretorio Database](#estrutura-do-diretorio-database)
- [Sistema de Conexao](#sistema-de-conexao)
- [Como Adicionar Nova Conexao](#como-adicionar-nova-conexao)
- [Migracoes e Schema](#migracoes-e-schema)
- [Padroes e Convencoes](#padroes-e-convencoes)
- [Testes](#testes)
- [Troubleshooting](#troubleshooting)

## 🎯 Visao Geral

O diretório `database/` é responsável por toda a **infraestrutura de dados** da aplicação e implementa:

- 🔌 **Gerenciamento de Conexões**: Pool de conexões com PostgreSQL
- 🏗️ **Abstração de Database**: Interface para diferentes tipos de banco
- 🔄 **Retry Logic**: Reconexão automática em caso de falhas
- ⚙️ **Configuração GORM**: ORM configurado com otimizações
- 🔒 **Transações**: Suporte a transações seguras
- 📊 **Monitoramento**: Métricas de performance e saúde

## 📁 Estrutura do Diretorio Database

```text
database/
├── connection.go          # 🔌 Interface e factory de conexões
└── postgres.go           # 🐘 Implementação específica do PostgreSQL
```

### 🧩 Componentes Principais

| Componente | Responsabilidade | Arquivo Principal | Tipo |
|------------|------------------|-------------------|------|
| **DatabaseConnection** | Interface de conexão | `connection.go` | Interface |
| **PostgreSQLConnection** | Implementação PostgreSQL | `postgres.go` | Implementação |
| **NewDatabaseConnection()** | Factory de conexões | `connection.go` | Factory |

## 🔌 Sistema de Conexao

### Interface DatabaseConnection

```go
type DatabaseConnection interface {
    Connect() (*gorm.DB, error)
    Close() error
    HealthCheck() error
    GetConnectionString() string
}
```

### Implementação PostgreSQL

```go
type PostgreSQLConnection struct {
    config *config.DatabaseConfig
    db     *gorm.DB
}

func (p *PostgreSQLConnection) Connect() (*gorm.DB, error) {
    // 1. Construir string de conexão
    dsn := p.GetConnectionString()

    // 2. Configurar GORM
    gormConfig := &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
        NamingStrategy: schema.NamingStrategy{
            SingularTable: false,
        },
    }

    // 3. Conectar com retry
    db, err := gorm.Open(postgres.Open(dsn), gormConfig)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to database: %w", err)
    }

    // 4. Configurar pool de conexões
    sqlDB, err := db.DB()
    if err != nil {
        return nil, fmt.Errorf("failed to get sql.DB: %w", err)
    }

    sqlDB.SetMaxIdleConns(10)
    sqlDB.SetMaxOpenConns(100)
    sqlDB.SetConnMaxLifetime(time.Hour)

    p.db = db
    return db, nil
}
```

### Factory Pattern

```go
func NewDatabaseConnection(cfg *config.DatabaseConfig) (DatabaseConnection, error) {
    switch cfg.Type {
    case "postgres", "postgresql", "":
        return NewPostgreSQLConnection(cfg), nil
    case "mysql":
        return NewMySQLConnection(cfg), nil
    case "sqlite":
        return NewSQLiteConnection(cfg), nil
    default:
        return nil, fmt.Errorf("unsupported database type: %s", cfg.Type)
    }
}
```

## ➕ Como Adicionar Nova Conexao

### Passo 1: Implementar a Interface

Crie `database/mysql.go` (exemplo):

```go
package database

import (
    "fmt"
    "time"

    "gorm.io/driver/mysql"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
    "gorm.io/gorm/schema"

    "github.com/fabianoflorentino/mr-robot/config"
)

type MySQLConnection struct {
    config *config.DatabaseConfig
    db     *gorm.DB
}

func NewMySQLConnection(cfg *config.DatabaseConfig) DatabaseConnection {
    return &MySQLConnection{
        config: cfg,
    }
}

func (m *MySQLConnection) Connect() (*gorm.DB, error) {
    // 1. Construir DSN específico do MySQL
    dsn := m.GetConnectionString()

    // 2. Configurar GORM para MySQL
    gormConfig := &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
        NamingStrategy: schema.NamingStrategy{
            SingularTable: false,
        },
        DisableForeignKeyConstraintWhenMigrating: true,
    }

    // 3. Conectar usando driver MySQL
    db, err := gorm.Open(mysql.Open(dsn), gormConfig)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to MySQL: %w", err)
    }

    // 4. Configurar pool específico para MySQL
    sqlDB, err := db.DB()
    if err != nil {
        return nil, fmt.Errorf("failed to get sql.DB: %w", err)
    }

    // Configurações otimizadas para MySQL
    sqlDB.SetMaxIdleConns(25)
    sqlDB.SetMaxOpenConns(200)
    sqlDB.SetConnMaxLifetime(30 * time.Minute)

    m.db = db
    return db, nil
}

func (m *MySQLConnection) GetConnectionString() string {
    return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
        m.config.User,
        m.config.Password,
        m.config.Host,
        m.config.Port,
        m.config.Database,
    )
}

func (m *MySQLConnection) Close() error {
    if m.db != nil {
        sqlDB, err := m.db.DB()
        if err != nil {
            return fmt.Errorf("failed to get sql.DB: %w", err)
        }
        return sqlDB.Close()
    }
    return nil
}

func (m *MySQLConnection) HealthCheck() error {
    if m.db == nil {
        return fmt.Errorf("database connection not initialized")
    }

    sqlDB, err := m.db.DB()
    if err != nil {
        return fmt.Errorf("failed to get sql.DB: %w", err)
    }

    return sqlDB.Ping()
}
```

### Passo 2: Atualizar a Factory

Em `connection.go`:

```go
func NewDatabaseConnection(cfg *config.DatabaseConfig) (DatabaseConnection, error) {
    switch cfg.Type {
    case "postgres", "postgresql", "":
        return NewPostgreSQLConnection(cfg), nil
    case "mysql":
        return NewMySQLConnection(cfg), nil  // ⬅️ Nova conexão
    case "sqlite":
        return NewSQLiteConnection(cfg), nil
    default:
        return nil, fmt.Errorf("unsupported database type: %s", cfg.Type)
    }
}
```

### Passo 3: Adicionar Configuração

Em `config/app_config.go`:

```go
type DatabaseConfig struct {
    Type     string // ⬅️ Adicionar tipo de banco
    Host     string
    Port     string
    User     string
    Password string
    Database string
    SSLMode  string
    Timezone string
}

// Na função LoadAppConfig()
Database: DatabaseConfig{
    Type:     getEnvOrDefault("DB_TYPE", "postgres"),  // ⬅️ Nova config
    Host:     getEnvOrDefault("POSTGRES_HOST", "localhost"),
    // ... resto das configurações
},
```

### Passo 4: Adicionar Dependências

Em `go.mod`:

```go
require (
    // ... dependências existentes
    gorm.io/driver/mysql v1.5.2  // ⬅️ Nova dependência
)
```

## 🔄 Migracoes e Schema

### Estrutura de Migrações

```text
database/migrations/
├── 001_create_payments_table.sql
├── 002_add_processor_name_to_payments.sql
├── 003_create_users_table.sql
└── 004_add_indexes.sql
```

### Implementação de Migração

```go
// Em internal/app/migration/manager.go
type Migration struct {
    Version     int
    Description string
    SQL         string
}

var migrations = []Migration{
    {
        Version:     1,
        Description: "Create payments table",
        SQL: `
            CREATE TABLE IF NOT EXISTS payments (
                id SERIAL PRIMARY KEY,
                correlation_id UUID NOT NULL UNIQUE,
                amount DECIMAL(10,2) NOT NULL,
                processor_name VARCHAR(100) NOT NULL,
                processed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
            );
        `,
    },
    {
        Version:     2,
        Description: "Add indexes to payments",
        SQL: `
            CREATE INDEX IF NOT EXISTS idx_payments_correlation_id ON payments(correlation_id);
            CREATE INDEX IF NOT EXISTS idx_payments_processed_at ON payments(processed_at);
            CREATE INDEX IF NOT EXISTS idx_payments_processor_name ON payments(processor_name);
        `,
    },
}

func (m *Manager) RunMigrations() error {
    // 1. Criar tabela de controle de migrações
    if err := m.createMigrationTable(); err != nil {
        return err
    }

    // 2. Verificar versão atual
    currentVersion, err := m.getCurrentVersion()
    if err != nil {
        return err
    }

    // 3. Executar migrações pendentes
    for _, migration := range migrations {
        if migration.Version > currentVersion {
            if err := m.executeMigration(migration); err != nil {
                return fmt.Errorf("failed to execute migration %d: %w", migration.Version, err)
            }
        }
    }

    return nil
}
```

### GORM Auto-Migrate (Alternativa)

```go
func (m *Manager) RunAutoMigrate() error {
    // Auto-migrate usando structs GORM
    return m.db.AutoMigrate(
        &data.PaymentModel{},
        &data.UserModel{},
        // ... outros modelos
    )
}
```

## 📏 Padroes e Convencoes

### ✅ Boas Práticas de Database

- **🔌 Interface First**: Sempre definir interfaces antes de implementações
- **🏭 Factory Pattern**: Usar factory para criação de conexões
- **🔄 Connection Pooling**: Configurar pools otimizados por tipo de banco
- **❌ Error Handling**: Tratar erros específicos de cada banco
- **📊 Monitoring**: Implementar health checks e métricas

### 📋 Convenções de Nomenclatura

| Tipo | Padrão | Exemplo |
|------|---------|---------|
| **Connection Struct** | `{Banco}Connection` | `PostgreSQLConnection`, `MySQLConnection` |
| **Factory Function** | `New{Banco}Connection` | `NewPostgreSQLConnection()` |
| **Config Type** | `{Banco}Config` | `PostgreSQLConfig`, `MySQLConfig` |
| **Migration File** | `{numero}_{descricao}.sql` | `001_create_table.sql` |

### 🔧 Configurações de Pool

```go
// PostgreSQL - Otimizado para alta concorrência
sqlDB.SetMaxIdleConns(10)    // Conexões idle
sqlDB.SetMaxOpenConns(100)   // Conexões máximas
sqlDB.SetConnMaxLifetime(time.Hour)  // Tempo de vida

// MySQL - Otimizado para throughput
sqlDB.SetMaxIdleConns(25)
sqlDB.SetMaxOpenConns(200)
sqlDB.SetConnMaxLifetime(30 * time.Minute)

// SQLite - Configuração mínima
sqlDB.SetMaxIdleConns(1)
sqlDB.SetMaxOpenConns(1)
sqlDB.SetConnMaxLifetime(0)
```

### 🔒 Configurações de Segurança

```go
// PostgreSQL com SSL
dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
    cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode, cfg.Timezone)

// Validações de segurança
func (p *PostgreSQLConnection) validateConfig() error {
    if p.config.Password == "" && p.config.SSLMode != "disable" {
        return fmt.Errorf("password required when SSL is enabled")
    }

    if p.config.SSLMode == "" {
        p.config.SSLMode = "require" // Default seguro
    }

    return nil
}
```

## 🧪 Testes

### Testando Conexões

```go
func TestPostgreSQLConnection_Connect(t *testing.T) {
    // Setup
    config := &config.DatabaseConfig{
        Host:     "localhost",
        Port:     "5432",
        User:     "test_user",
        Password: "test_pass",
        Database: "test_db",
        SSLMode:  "disable",
        Timezone: "UTC",
    }

    conn := NewPostgreSQLConnection(config)

    // Act
    db, err := conn.Connect()

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, db)

    // Cleanup
    defer conn.Close()

    // Test health check
    err = conn.HealthCheck()
    assert.NoError(t, err)
}

func TestPostgreSQLConnection_InvalidConfig(t *testing.T) {
    config := &config.DatabaseConfig{
        Host: "invalid-host",
        Port: "9999",
    }

    conn := NewPostgreSQLConnection(config)

    _, err := conn.Connect()
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "failed to connect")
}
```

### Testando Factory

```go
func TestNewDatabaseConnection(t *testing.T) {
    tests := []struct {
        name     string
        dbType   string
        wantType interface{}
        wantErr  bool
    }{
        {
            name:     "PostgreSQL",
            dbType:   "postgres",
            wantType: &PostgreSQLConnection{},
            wantErr:  false,
        },
        {
            name:     "MySQL",
            dbType:   "mysql",
            wantType: &MySQLConnection{},
            wantErr:  false,
        },
        {
            name:    "Unsupported",
            dbType:  "oracle",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            config := &config.DatabaseConfig{Type: tt.dbType}

            conn, err := NewDatabaseConnection(config)

            if tt.wantErr {
                assert.Error(t, err)
                assert.Nil(t, conn)
            } else {
                assert.NoError(t, err)
                assert.IsType(t, tt.wantType, conn)
            }
        })
    }
}
```

### Testando Migrações

```go
func TestMigrationManager_RunMigrations(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)

    manager := NewMigrationManager(db)

    // Act
    err := manager.RunMigrations()

    // Assert
    assert.NoError(t, err)

    // Verify tables were created
    var tableExists bool
    err = db.Raw("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'payments')").Scan(&tableExists).Error
    assert.NoError(t, err)
    assert.True(t, tableExists)

    // Verify migration version
    version, err := manager.getCurrentVersion()
    assert.NoError(t, err)
    assert.Greater(t, version, 0)
}
```

## 🔧 Troubleshooting

### Problemas Comuns

| Problema | Causa Provável | Solução |
|----------|----------------|---------|
| **Connection refused** | Banco não está rodando | Verificar se container/serviço está ativo |
| **Authentication failed** | Credenciais incorretas | Verificar user/password no .env |
| **Too many connections** | Pool mal configurado | Ajustar MaxOpenConns/MaxIdleConns |
| **SSL error** | Configuração SSL incorreta | Verificar SSLMode (disable/require/verify-full) |
| **Migration failed** | Schema inconsistente | Verificar logs e estado da tabela de migrações |

### Debug de Conexão

```go
// Habilitar logs detalhados do GORM
gormConfig := &gorm.Config{
    Logger: logger.Default.LogMode(logger.Info),
}

// Log de connection string (sem senha)
func (p *PostgreSQLConnection) logConnectionInfo() {
    log.Printf("Connecting to PostgreSQL at %s:%s/%s",
        p.config.Host, p.config.Port, p.config.Database)
    log.Printf("SSL Mode: %s", p.config.SSLMode)
    log.Printf("Timezone: %s", p.config.Timezone)
}
```

### Verificações de Saúde

```bash
# Testar conexão direta com banco
psql -h localhost -p 5432 -U postgres -d mr_robot -c "SELECT 1;"

# Verificar pool de conexões
docker-compose exec mr_robot_app \
  sh -c "curl localhost:8888/health | jq '.database'"

# Monitorar conexões ativas
psql -h localhost -p 5432 -U postgres -c \
  "SELECT count(*) as active_connections FROM pg_stat_activity WHERE state = 'active';"
```

### Métricas de Performance

```go
type DatabaseMetrics struct {
    ActiveConnections   int64
    IdleConnections     int64
    ConnectionsInUse    int64
    ConnectionWaitCount int64
    ConnectionWaitTime  time.Duration
}

func (p *PostgreSQLConnection) GetMetrics() (*DatabaseMetrics, error) {
    sqlDB, err := p.db.DB()
    if err != nil {
        return nil, err
    }

    stats := sqlDB.Stats()

    return &DatabaseMetrics{
        ActiveConnections:   int64(stats.OpenConnections),
        IdleConnections:     int64(stats.Idle),
        ConnectionsInUse:    int64(stats.InUse),
        ConnectionWaitCount: stats.WaitCount,
        ConnectionWaitTime:  stats.WaitDuration,
    }, nil
}
```

### Monitoramento Avançado

```go
// Health check detalhado
func (p *PostgreSQLConnection) DetailedHealthCheck() map[string]interface{} {
    result := make(map[string]interface{})

    // 1. Ping básico
    if err := p.HealthCheck(); err != nil {
        result["ping"] = "failed"
        result["error"] = err.Error()
        return result
    }
    result["ping"] = "ok"

    // 2. Métricas de conexão
    if metrics, err := p.GetMetrics(); err == nil {
        result["metrics"] = metrics
    }

    // 3. Teste de query
    var version string
    if err := p.db.Raw("SELECT version()").Scan(&version).Error; err == nil {
        result["query_test"] = "ok"
        result["postgres_version"] = version
    } else {
        result["query_test"] = "failed"
    }

    // 4. Verificar transações
    tx := p.db.Begin()
    if tx.Error == nil {
        tx.Rollback()
        result["transaction_test"] = "ok"
    } else {
        result["transaction_test"] = "failed"
    }

    return result
}
```

## 📊 Otimizações de Performance

### Índices Recomendados

```sql
-- Índices para tabela de pagamentos
CREATE INDEX CONCURRENTLY idx_payments_correlation_id ON payments(correlation_id);
CREATE INDEX CONCURRENTLY idx_payments_processed_at ON payments(processed_at);
CREATE INDEX CONCURRENTLY idx_payments_processor_name ON payments(processor_name);
CREATE INDEX CONCURRENTLY idx_payments_amount ON payments(amount);

-- Índice composto para queries de resumo
CREATE INDEX CONCURRENTLY idx_payments_summary
ON payments(processor_name, processed_at)
INCLUDE (amount);
```

### Configurações PostgreSQL

```sql
-- Em postgresql.conf
shared_buffers = 256MB
effective_cache_size = 1GB
maintenance_work_mem = 64MB
checkpoint_completion_target = 0.9
wal_buffers = 16MB
default_statistics_target = 100
random_page_cost = 1.1
effective_io_concurrency = 200
```

## 📞 Contato

Para dúvidas sobre a arquitetura de banco de dados ou sugestões de melhorias, abra uma issue no repositório ou entre em contato com a equipe de desenvolvimento.

---

**📝 Nota**: Este documento deve ser atualizado sempre que novos tipos de banco, otimizações ou padrões forem adicionados à aplicação.
