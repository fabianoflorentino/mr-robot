# Troubleshooting Unix Sockets - Mr. Robot

Este documento fornece soluções para problemas comuns relacionados à implementação de Unix sockets no projeto Mr. Robot.

## 🚨 Problemas Comuns

### 1. HAProxy não consegue conectar aos sockets

#### Sintomas
- HAProxy retorna erro 503 Service Unavailable
- Logs do HAProxy mostram "connect() failed"
- Backends aparecem como "DOWN" no HAProxy stats

#### Soluções

```bash
# 1. Verificar se os containers estão rodando
make dev-status

# 2. Verificar se os arquivos de socket existem
docker exec mr_robot1 ls -la /var/run/mr_robot/

# 3. Verificar permissões dos sockets
docker exec mr_robot1 stat -c "%a %n" /var/run/mr_robot/*.sock

# 4. Reiniciar o ambiente
make dev-restart
```

### 2. Arquivos de socket não são criados

#### Sintomas
- Aplicação inicia mas não cria arquivos .sock
- Erro "Failed to create Unix socket listener"
- Diretório /var/run/mr_robot/ está vazio

#### Soluções

```bash
# 1. Verificar configuração das variáveis de ambiente
make socket-mode-status

# 2. Verificar logs da aplicação
make dev-logs | grep -i socket

# 3. Verificar se o volume está montado corretamente
docker inspect mr_robot1 | grep -A 10 "Mounts"

# 4. Recriar volumes
make clean-volumes && make dev-up
```

### 3. Permissões incorretas nos sockets

#### Sintomas
- HAProxy não consegue ler os arquivos de socket
- Erro "Permission denied" nos logs
- Sockets existem mas HAProxy não conecta

#### Soluções

```bash
# 1. Verificar permissões atuais
docker exec mr_robot1 ls -la /var/run/mr_robot/

# 2. Corrigir permissões manualmente (temporário)
docker exec mr_robot1 chmod 666 /var/run/mr_robot/*.sock

# 3. Reiniciar para aplicar configuração automática
make dev-restart
```

### 4. Aplicação usa TCP em vez de Unix sockets

#### Sintomas
- Logs mostram "Starting HTTP server on port 8888"
- Não há arquivos .sock criados
- HAProxy falha ao conectar

#### Soluções

```bash
# 1. Verificar configuração
make socket-mode-status

# 2. Habilitar Unix sockets
make enable-unix-socket-mode

# 3. Verificar arquivo de configuração
cat config/.env | grep UNIX_SOCKET

# 4. Reiniciar com nova configuração
make dev-restart
```

## 🔧 Comandos de Diagnóstico

### Verificação Completa

```bash
# Executar diagnóstico completo
make debug-unix-sockets
```

### Verificações Específicas

```bash
# Status dos containers
docker ps --format "table {{.Names}}\t{{.Status}}" | grep mr_robot

# Verificar sockets nos containers
docker exec mr_robot1 ls -la /var/run/mr_robot/
docker exec mr_robot2 ls -la /var/run/mr_robot/

# Testar conectividade HAProxy
curl -v http://localhost:9999/health

# Verificar logs específicos de Unix socket
docker logs mr_robot1 2>&1 | grep -i "unix\|socket"
docker logs mr_robot2 2>&1 | grep -i "unix\|socket"

# Verificar configuração do HAProxy
docker exec mr_robot_lb cat /usr/local/etc/haproxy/haproxy.cfg | grep -A 5 backend
```

### HAProxy Stats e Monitoring

```bash
# Verificar status dos backends (se stats estiver habilitado)
curl -s http://localhost:8404/stats

# Verificar saúde dos endpoints diretamente
curl http://localhost:9999/health
curl http://localhost:9999/payment-summary
```

## 🔄 Soluções por Etapas

### Solução 1: Reset Completo

```bash
# Para quando tudo falhou
make dev-down
make clean-volumes
make dev-up
make test-unix-sockets
```

### Solução 2: Alternar para TCP Temporariamente

```bash
# Fallback rápido para TCP
make enable-tcp-mode
make dev-restart

# Verificar se funcionou
curl http://localhost:9999/health

# Voltar para Unix sockets quando corrigido
make enable-unix-socket-mode
make dev-restart
```

### Solução 3: Verificação de Configuração

```bash
# 1. Verificar se config/.env existe
ls -la config/.env

# 2. Se não existir, criar a partir do exemplo
cp .env.example config/.env

# 3. Verificar configuração
make socket-mode-status

# 4. Reiniciar
make dev-restart
```

## 📋 Checklist de Troubleshooting

### ✅ Verificações Básicas

- [ ] Containers mr_robot1, mr_robot2 e mr_robot_lb estão rodando
- [ ] Volume socket_volume está criado e montado
- [ ] Arquivo config/.env existe e está configurado
- [ ] USE_UNIX_SOCKET=true está definido
- [ ] SOCKET_PATH está configurado para cada instância

### ✅ Verificações de Sistema

- [ ] Arquivos .sock existem em /var/run/mr_robot/
- [ ] Permissões dos sockets são 666
- [ ] HAProxy consegue acessar o volume de sockets
- [ ] Logs da aplicação mostram "Unix socket" na inicialização

### ✅ Verificações de Conectividade

- [ ] HAProxy responde na porta 9999
- [ ] curl http://localhost:9999/health retorna 200
- [ ] HAProxy stats (se habilitado) mostra backends UP
- [ ] Load balancing funciona entre as instâncias

## 🐛 Problemas Conhecidos

### Docker Desktop no Windows/Mac

**Problema**: Unix sockets podem ter limitações no Docker Desktop.

**Solução**: Use TCP mode:
```bash
make enable-tcp-mode
make dev-restart
```

### SELinux no Linux

**Problema**: SELinux pode bloquear acesso aos sockets.

**Solução**: 
```bash
# Verificar se SELinux está ativo
sestatus

# Configurar contexto adequado (se necessário)
sudo setsebool -P container_manage_cgroup true
```

### Volumes persistentes corrompidos

**Problema**: Volumes Docker corrompidos podem causar problemas.

**Solução**:
```bash
# Reset completo de volumes
docker volume prune -f
make clean-volumes
make dev-up
```

## 🚀 Melhoramento de Performance

### Monitoramento de Performance

```bash
# Teste básico de performance
time curl -s http://localhost:9999/health

# Teste de carga simples
for i in {1..100}; do curl -s http://localhost:9999/health > /dev/null; done

# Monitor de recursos
docker stats mr_robot1 mr_robot2 mr_robot_lb
```

### Otimizações

1. **HAProxy**: Ajustar timeouts no config/haproxy.cfg
2. **Go Application**: Verificar configurações de timeout
3. **Docker**: Ajustar limites de recursos nos compose files

## 📞 Quando Procurar Ajuda

Se após seguir todos os passos o problema persistir:

1. **Colete informações**:
   ```bash
   make debug-unix-sockets > debug-output.txt
   make dev-logs >> debug-output.txt
   ```

2. **Documente o problema**:
   - Versão do Docker
   - Sistema operacional
   - Comandos executados
   - Mensagens de erro específicas

3. **Tente o fallback TCP**:
   ```bash
   make enable-tcp-mode
   make dev-restart
   ```

## 📚 Referências

- [📖 UNIX_SOCKETS.md](UNIX_SOCKETS.md) - Documentação técnica completa
- [📖 ARCHITECTURE_GUIDE.md](ARCHITECTURE_GUIDE.md) - Guia geral de arquitetura
- [📖 README.md](../README.md) - Documentação principal do projeto
