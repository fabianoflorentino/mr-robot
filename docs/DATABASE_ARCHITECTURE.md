# Arquitetura do Diretório Database - Guia de Manutenção

> **Consulte também**: [📖 ARCHITECTURE_GUIDE.md](ARCHITECTURE_GUIDE.md) para padrões gerais e convenções consolidadas.

Este documento foca especificamente no **diretório `database`** e sua infraestrutura de dados.

## 🎯 Responsabilidades Específicas do Database

- 🔌 **Gerenciamento de Conexões**: Pool de conexões com PostgreSQL
- 🏗️ **Abstração de Database**: Interface para diferentes tipos de banco
- 🔄 **Retry Logic**: Reconexão automática em caso de falhas
- ⚙️ **SQL Nativo**: Implementação usando pgx driver para PostgreSQL
- 🔒 **Transações**: Suporte a transações seguras com SQL nativo
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
    Connect() (*sql.DB, error)
    Close() error
    GetDB() *sql.DB
}
```

### Implementação PostgreSQL

```go
type PostgreSQLConnection struct {
    config *config.DatabaseConfig
    db     *sql.DB
}

func (p *PostgreSQLConnection) Connect() (*sql.DB, error) {
    // 1. Construir string de conexão
    dsn := p.buildConnectionString()

    // 2. Conectar usando pgx driver
    db, err := sql.Open("pgx", dsn)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to database: %w", err)
    }

    // 3. Testar conexão
    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }

    // 4. Configurar pool de conexões
    db.SetMaxIdleConns(5)
    db.SetMaxOpenConns(25)
    db.SetConnMaxLifetime(time.Hour)

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

## ⚙️ Configuração de Pool de Conexões

```go
// Configurações recomendadas para PostgreSQL
db.SetMaxIdleConns(5)        // Conexões inativas
db.SetMaxOpenConns(25)       // Máximo de conexões
db.SetConnMaxLifetime(time.Hour) // Tempo de vida das conexões
```

### Configurações por Ambiente

| Ambiente | MaxIdle | MaxOpen | MaxLifetime |
|----------|---------|---------|-------------|
| **Development** | 2 | 5 | 30min |
| **Testing** | 1 | 3 | 15min |

---

## 🔄 Migrations com SQL Nativo

O sistema utiliza migrações SQL nativas para controle de versão do banco de dados:

### Estrutura de Migrations

```sql
-- 001_create_payments_table.sql
CREATE TABLE IF NOT EXISTS payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    amount DECIMAL(10,2) NOT NULL,
    status VARCHAR(20) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_payments_status ON payments(status);
CREATE INDEX idx_payments_created_at ON payments(created_at);
```

### Passo 2: Atualizar a Factory

### Execução de Migrations

```go
// migration/manager.go
type MigrationManager struct {
    db *sql.DB
}

func (m *MigrationManager) RunMigrations() error {
    migrations := []string{
        "001_create_payments_table.sql",
        "002_add_payment_indexes.sql",
        "003_create_audit_table.sql",
    }
    
    for _, migration := range migrations {
        if err := m.executeMigration(migration); err != nil {
            return fmt.Errorf("failed to execute migration %s: %w", migration, err)
        }
    }
    
    return nil
}
```

---

## 🗄️ Repositórios com SQL Nativo

### Interface do Repositório

```go
// core/repository/payment_repository.go
type PaymentRepository interface {
    Save(ctx context.Context, payment *domain.Payment) error
    FindByID(ctx context.Context, id string) (*domain.Payment, error)
    FindAll(ctx context.Context) ([]*domain.Payment, error)
    Update(ctx context.Context, payment *domain.Payment) error
    Delete(ctx context.Context, id string) error
}
```

### Implementação com SQL Nativo

```go
// adapters/outbound/persistence/data/payment_repository.go
type paymentRepository struct {
    db *sql.DB
}

func NewPaymentRepository(db *sql.DB) core.PaymentRepository {
    return &paymentRepository{db: db}
}

func (r *paymentRepository) Save(ctx context.Context, payment *domain.Payment) error {
    query := `
        INSERT INTO payments (id, amount, status, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5)
    `
    
    _, err := r.db.ExecContext(ctx, query,
        payment.ID,
        payment.Amount,
        payment.Status,
        payment.CreatedAt,
        payment.UpdatedAt,
    )
    
    return err
}
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

}
```

### Controle de Versão das Migrations

```go
func (m *Manager) createMigrationTable() error {
    query := `
        CREATE TABLE IF NOT EXISTS schema_migrations (
            version INTEGER PRIMARY KEY,
            applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )
    `
    _, err := m.db.Exec(query)
    return err
}

func (m *Manager) getCurrentVersion() (int, error) {
    var version int
    query := "SELECT COALESCE(MAX(version), 0) FROM schema_migrations"
    err := m.db.QueryRow(query).Scan(&version)
    return version, err
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
