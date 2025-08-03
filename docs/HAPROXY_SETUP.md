# HAProxy Load Balancer Configuration

Este projeto utiliza HAProxy como load balancer - uma solução extremamente eficiente e leve.

## Por que HAProxy?

- **Baixo consumo de recursos**: ~5-15MB RAM, muito eficiente em CPU
- **Alta performance**: Capaz de lidar com milhares de conexões simultâneas
- **Configuração simples**: Arquivo de configuração direto e legível
- **Confiabilidade**: Amplamente usado em produção por grandes empresas

## Configuração

O HAProxy está configurado para:

- Escutar na porta 9999 (load balancer)
- Interface de estatísticas na porta 8404
- Balanceamento round-robin entre 3 instâncias da aplicação
- Health checks automáticos nos serviços backend
- Logs detalhados para monitoramento

## Arquivos importantes

- `docker-compose.prod.yml`: Configuração principal do HAProxy
- `config/haproxy.cfg`: Configuração do HAProxy com backends e health checks

## Portas expostas

- **9999**: Load balancer principal (entrada da aplicação)
- **8404**: Interface web de estatísticas do HAProxy

## Verificando o status

Para verificar se o HAProxy está funcionando:

```bash
# Verificar estatísticas do HAProxy
curl http://localhost:8404/stats

# Testar a aplicação através do HAProxy
curl http://localhost:9999/

# Verificar logs do container
docker logs mr_robot_lb
```

## Load Balancing

O HAProxy está configurado com:

- **Algoritmo**: Round-robin (distribuição circular)
- **Health checks**: GET / a cada 30 segundos
- **Failover**: Automático quando um backend falha
- **Recuperação**: Automática quando backend volta online

## Recursos utilizados

- **RAM**: ~5-15MB (muito eficiente)
- **CPU**: Máximo 0.5 cores
- **Imagem**: haproxy:2.9-alpine (imagem otimizada)

## Monitoramento

Acesse `http://localhost:8404/stats` para ver:

- Status de cada backend
- Número de conexões ativas
- Estatísticas de tráfego
- Health check status
