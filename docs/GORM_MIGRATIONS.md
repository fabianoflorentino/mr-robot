# Sistema de Migrações com GORM

## Visão Geral

O GORM oferece funcionalidades integradas para lidar com migrações de banco de dados, eliminando a necessidade de implementações customizadas complexas. Este documento apresenta as opções disponíveis.

## Opções de Implementação

### 1. Abordagem Simples (Implementada)

Usa `HasTable()` para verificar se a tabela existe antes de criar:

```go
if !db.Migrator().HasTable(&data.Payment{}) {
    db.AutoMigrate(&data.Payment{})
}
```

**Vantagens:**

- Código simples e direto
- Proteção contra execução duplicada
- Não requer tabelas de controle adicionais

### 2. Apenas AutoMigrate (Mais Simples)

Deixa o GORM decidir automaticamente:

```go
db.AutoMigrate(&data.Payment{})
```

**Vantagens:**

- Código mínimo
- GORM cuida de tudo automaticamente
- Idempotente por padrão

### 3. Abordagem Avançada (Opcional)

Para casos específicos que precisem de mais controle, você pode implementar verificações granulares usando métodos do Migrator:

```go
// Verificar colunas específicas
if !db.Migrator().HasColumn(&data.Payment{}, "processor") {
    db.AutoMigrate(&data.Payment{})
}

// Verificar índices
if !db.Migrator().HasIndex(&data.Payment{}, "correlation_id") {
    db.Migrator().CreateIndex(&data.Payment{}, "correlation_id")
}
```

**Vantagens:**

- Controle granular sobre mudanças
- Pode verificar colunas e índices individualmente
- Status detalhado de migrações

## Funcionalidades do GORM Migrator

### Verificações

```go
db.Migrator().HasTable(&Model{})           // Tabela existe?
db.Migrator().HasColumn(&Model{}, "name")  // Coluna existe?
db.Migrator().HasIndex(&Model{}, "idx")    // Índice existe?
db.Migrator().HasConstraint(&Model{}, "fk") // Constraint existe?
```

### Operações

```go
db.Migrator().CreateTable(&Model{})        // Criar tabela
db.Migrator().DropTable(&Model{})          // Remover tabela
db.Migrator().AddColumn(&Model{}, "name")  // Adicionar coluna
db.Migrator().DropColumn(&Model{}, "name") // Remover coluna
db.Migrator().CreateIndex(&Model{}, "idx") // Criar índice
db.Migrator().DropIndex(&Model{}, "idx")   // Remover índice
```

## Comparação das Abordagens

| Abordagem | Complexidade | Controle | Overhead | Recomendado Para |
|-----------|--------------|----------|----------|------------------|
| AutoMigrate | Baixa | Baixo | Mínimo | Projetos simples |
| HasTable + AutoMigrate | Média | Médio | Baixo | Projetos médios (implementada) |
| Migrator Avançado | Alta | Alto | Médio | Casos específicos |
| Sistema Custom | Muito Alta | Muito Alto | Alto | Casos muito específicos |

## Proteção Contra Concorrência

Todas as implementações incluem mutex para prevenir execução simultânea:

```go
m.mutex.Lock()
defer m.mutex.Unlock()
```

## Recomendação

Para a maioria dos casos, a **Abordagem Simples** (implementada no `Manager`) é suficiente:

- Protege contra duplicatas
- Código limpo e mantível
- Aproveita funcionalidades nativas do GORM
- Logs informativos

Use a **Abordagem Avançada** apenas se precisar implementar:

- Controle granular sobre mudanças específicas
- Status detalhado de migrações
- Operações de rollback customizadas

Use apenas **AutoMigrate** se:

- O projeto é muito simples
- Não há múltiplas instâncias
- Logs detalhados não são necessários

## Exemplo de Uso

```go
// Abordagem simples (implementada)
migrationManager := migration.NewManager(db)
if err := migrationManager.RunMigrations(); err != nil {
    log.Fatal("Failed to run migrations:", err)
}
```
