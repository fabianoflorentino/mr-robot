# Sistema de Fallback - Mr Robot

## 📋 Visão Geral

O Mr Robot implementa um sistema de fallback robusto e automático para processamento de pagamentos. Este documento detalha como o sistema funciona, como configurá-lo e como monitorá-lo.

## 🏗️ Arquitetura do Fallback

### Implementação

O sistema de fallback é implementado na camada `PaymentServiceWithFallback` que:

1. **Gerencia dois processadores**: Default (principal) e Fallback (secundário)
2. **Aplica proteções**: Circuit Breaker e Rate Limiter para ambos
3. **Tenta sequencialmente**: Default primeiro, Fallback se o primeiro falhar
4. **Registra o resultado**: Persiste no banco qual processador foi usado

### Classes Principais

- `PaymentServiceWithFallback`: Service principal com lógica de fallback
- `ProcessGateway`: Gateway genérico que funciona para ambos os processadores
- `PaymentServiceInterface`: Interface comum para flexibilidade

## 🔧 Configuração

### Variáveis de Ambiente

```bash
# Processador principal (tentativa 1)
DEFAULT_PROCESSOR_URL=http://primary-gateway.com/api/payments

# Processador de fallback (tentativa 2, se a primeira falhar)
FALLBACK_PROCESSOR_URL=http://backup-gateway.com/api/payments
```

### Parâmetros do Circuit Breaker

```go
// Configuração atual no código
circuitBreaker: NewCircuitBreaker(5, 30*time.Second)
// 5 falhas consecutivas em 30 segundos para abrir o circuit
```

### Parâmetros do Rate Limiter

```go
// Configuração atual no código
rateLimiter: NewRateLimiter(3)
// Máximo 3 processamentos simultâneos
```

## 🔄 Fluxo de Funcionamento

### Cenário de Sucesso (Default)

```text
1. Pagamento recebido
2. Tenta Default Processor
3. ✅ Sucesso
4. Salva no DB com processor="default"
5. Fim
```

### Cenário de Fallback

```text
1. Pagamento recebido
2. Tenta Default Processor
3. ❌ Falha (timeout, erro HTTP, etc.)
4. Tenta Fallback Processor
5. ✅ Sucesso
6. Salva no DB com processor="fallback"
7. Fim
```

### Cenário de Falha Total

```text
1. Pagamento recebido
2. Tenta Default Processor
3. ❌ Falha
4. Tenta Fallback Processor
5. ❌ Falha também
6. Retorna erro
7. Job volta para a fila (retry)
```

## 📊 Monitoramento

### Endpoint de Resumo

```bash
GET /payment-summary
```

**Resposta de exemplo:**

```json
{
  "default": {
    "totalRequests": 950,
    "totalAmount": 125750.50
  },
  "fallback": {
    "totalRequests": 50,
    "totalAmount": 6250.00
  }
}
```

### Interpretação dos Dados

- **`default.totalRequests > 0`**: Processador principal funcionando
- **`fallback.totalRequests > 0`**: Houve falhas no processador principal
- **Proporção alta de fallback**: Possível problema no processador principal

### Logs de Monitoramento

O sistema registra logs quando usa o fallback:

```text
Default processor failed: connection timeout, trying fallback...
```

## 🚨 Alertas e Troubleshooting

### Quando se Preocupar

1. **Alta taxa de fallback** (>10%): Investigar processador principal
2. **Fallback total = 0**: Verificar se URL do fallback está correta
3. **Ambos falhando**: Problemas de conectividade ou configuração

### Verificações Comuns

```bash
# Testar conectividade com processadores
curl -X POST $DEFAULT_PROCESSOR_URL -H "Content-Type: application/json" -d '{}'
curl -X POST $FALLBACK_PROCESSOR_URL -H "Content-Type: application/json" -d '{}'

# Verificar configuração
echo $DEFAULT_PROCESSOR_URL
echo $FALLBACK_PROCESSOR_URL

# Verificar logs da aplicação
docker-compose logs mr_robot_app
```

## 🧪 Testes

### Teste Manual do Fallback

1. **Configure URLs de teste:**

   ```bash
   DEFAULT_PROCESSOR_URL=http://httpbin.org/status/500  # Sempre falha
   FALLBACK_PROCESSOR_URL=http://httpbin.org/status/200 # Sempre sucesso
   ```

2. **Envie um pagamento:**

   ```bash
   curl -X POST http://localhost:8888/payments \
     -H "Content-Type: application/json" \
     -d '{"correlationId": "123e4567-e89b-12d3-a456-426614174000", "amount": 100.50}'
   ```

3. **Verifique o resultado:**

   ```bash
   curl http://localhost:8888/payment-summary
   ```

   Deve mostrar o pagamento em `fallback.totalRequests`.

### Teste de Ambos Funcionando

1. **Configure URLs que funcionam:**

   ```bash
   DEFAULT_PROCESSOR_URL=http://httpbin.org/status/200
   FALLBACK_PROCESSOR_URL=http://httpbin.org/status/200
   ```

2. **Envie pagamentos** e verifique que todos vão para `default`.

## 🔒 Considerações de Segurança

- **URLs HTTPS**: Use sempre HTTPS em produção
- **Autenticação**: Configure autenticação adequada nos processadores
- **Timeouts**: Configure timeouts apropriados para evitar travamentos
- **Rate Limiting**: O Rate Limiter protege contra sobrecarga

## 📈 Melhorias Futuras

- [ ] Métricas detalhadas (Prometheus/Grafana)
- [ ] Configuração de timeouts por processador
- [ ] Health checks dos processadores
- [ ] Balanceamento de carga entre múltiplos fallbacks
- [ ] Retry com backoff exponencial no nível do processador
