# Sistema de Migrações com SQL Nativo

> 🗄️ **Guia de Migrações de Banco de Dados usando SQL puro no Mr. Robot**

O sistema de migrações agora utiliza SQL nativo através do driver pgx, eliminando a dependência de ORMs e proporcionando controle total sobre as operações de banco de dados. Este documento apresenta as opções e implementações disponíveis.

## Abordagem Atual: SQL Nativo

### Implementação Automática

O sistema verifica automaticamente se as tabelas existem e cria conforme necessário:

```go
// internal/app/migration/manager.go
func (m *Manager) RunMigrations() error {
    if !m.isTableExists("payments") {
        if err := m.createPaymentsTable(); err != nil {
            return fmt.Errorf("failed to create payments table: %w", err)
        }
    }
    return nil
}
```

### Vantagens do SQL Nativo

- ✅ Controle total sobre DDL (Data Definition Language)
- ✅ Performance otimizada sem camadas de abstração
- ✅ Queries SQL específicas para PostgreSQL
- ✅ Menor dependência externa
- ✅ Controle fino sobre índices e constraints

## Script de Migração Atual

```sql
CREATE TABLE IF NOT EXISTS payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    correlation_id UUID NOT NULL,
    amount DECIMAL(15,2) NOT NULL,
    processor VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payments_correlation_id ON payments(correlation_id);
CREATE INDEX IF NOT EXISTS idx_payments_processor ON payments(processor);
CREATE INDEX IF NOT EXISTS idx_payments_created_at ON payments(created_at);
```

## Funcionalidades do Sistema de Migração

### Verificação Inteligente

```go
func (m *Manager) isTableExists(tableName string) bool {
    var exists bool
    query := `SELECT EXISTS (
        SELECT 1 FROM information_schema.tables 
        WHERE table_schema = 'public' AND table_name = $1
    )`
    err := m.db.QueryRow(query, tableName).Scan(&exists)
    return err == nil && exists
}
```

### Criação Segura de Tabelas

- ✅ Usa `CREATE TABLE IF NOT EXISTS` para evitar conflitos
- ✅ Cria índices automaticamente para otimização
- ✅ Suporte completo a UUIDs do PostgreSQL
- ✅ Timestamps automáticos com timezone

## Comandos Úteis

```bash
# Verificar estrutura da tabela
docker exec -it mr-robot-db psql -U mr_robot -d mr_robot -c "\d payments"

# Ver todos os índices
docker exec -it mr-robot-db psql -U mr_robot -d mr_robot -c "\di"

# Verificar dados
docker exec -it mr-robot-db psql -U mr_robot -d mr_robot -c "SELECT * FROM payments LIMIT 5;"
```
