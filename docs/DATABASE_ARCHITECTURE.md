# Arquitetura do Diretório Database - Guia de Manutenção

> **Consulte também**: [📖 ARCHITECTURE_GUIDE.md](ARCHITECTURE_GUIDE.md) para padrões gerais e convenções consolidadas.

Este documento foca especificamente no **diretório `database`** e sua infraestrutura de dados.

## 🎯 Responsabilidades Específicas do Database

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

### Otimizações PostgreSQL Específicas

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

### Configurações postgresql.conf

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

---

**� Nota**: Para padrões gerais, convenções de nomenclatura e troubleshooting consolidado, consulte o [📖 ARCHITECTURE_GUIDE.md](ARCHITECTURE_GUIDE.md).
