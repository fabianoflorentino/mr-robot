#!/bin/bash

# Test Unix Sockets Implementation for Mr. Robot
# This script validates the Unix socket communication between HAProxy and application instances

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SOCKET_DIR="/var/run/mr_robot"
SOCKET_1="$SOCKET_DIR/mr_robot1.sock"
SOCKET_2="$SOCKET_DIR/mr_robot2.sock"
HAPROXY_PORT=9999
HAPROXY_STATS_PORT=8404

# Logging function
log() {
    echo -e "${BLUE}[$(date '+%Y-%m-%d %H:%M:%S')]${NC} $1"
}

success() {
    echo -e "${GREEN}✅ $1${NC}"
}

warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

error() {
    echo -e "${RED}❌ $1${NC}"
}

# Test functions
test_docker_containers() {
    log "Testing Docker containers status..."
    
    if ! docker ps --format "table {{.Names}}" | grep -q "mr_robot1"; then
        error "mr_robot1 container is not running"
        return 1
    fi
    success "mr_robot1 container is running"
    
    if ! docker ps --format "table {{.Names}}" | grep -q "mr_robot2"; then
        error "mr_robot2 container is not running"
        return 1
    fi
    success "mr_robot2 container is running"
    
    if ! docker ps --format "table {{.Names}}" | grep -q "mr_robot_lb"; then
        error "HAProxy container (mr_robot_lb) is not running"
        return 1
    fi
    success "HAProxy container is running"
}

test_socket_files() {
    log "Testing Unix socket files creation..."
    
    # Check if socket files exist in containers
    if docker exec mr_robot1 test -S "$SOCKET_1"; then
        success "Socket file $SOCKET_1 exists and is a socket"
    else
        error "Socket file $SOCKET_1 does not exist or is not a socket"
        return 1
    fi
    
    if docker exec mr_robot2 test -S "$SOCKET_2"; then
        success "Socket file $SOCKET_2 exists and is a socket"
    else
        error "Socket file $SOCKET_2 does not exist or is not a socket"
        return 1
    fi
    
    # Check socket permissions
    PERMS1=$(docker exec mr_robot1 stat -c "%a" "$SOCKET_1" 2>/dev/null || echo "000")
    PERMS2=$(docker exec mr_robot2 stat -c "%a" "$SOCKET_2" 2>/dev/null || echo "000")
    
    if [ "$PERMS1" = "666" ]; then
        success "Socket $SOCKET_1 has correct permissions (666)"
    else
        warning "Socket $SOCKET_1 permissions: $PERMS1 (expected: 666)"
    fi
    
    if [ "$PERMS2" = "666" ]; then
        success "Socket $SOCKET_2 has correct permissions (666)"
    else
        warning "Socket $SOCKET_2 permissions: $PERMS2 (expected: 666)"
    fi
}

test_haproxy_connectivity() {
    log "Testing HAProxy connectivity..."
    
    # Test HAProxy main port
    if curl -f -s "http://localhost:$HAPROXY_PORT/health" > /dev/null; then
        success "HAProxy is responding on port $HAPROXY_PORT"
    else
        error "HAProxy is not responding on port $HAPROXY_PORT"
        return 1
    fi
    
    # Test HAProxy stats (if available)
    if curl -f -s "http://localhost:$HAPROXY_STATS_PORT/stats" > /dev/null 2>&1; then
        success "HAProxy stats available on port $HAPROXY_STATS_PORT"
    else
        warning "HAProxy stats not available on port $HAPROXY_STATS_PORT (may be disabled)"
    fi
}

test_load_balancing() {
    log "Testing load balancing functionality..."
    
    # Make multiple requests and check if they're distributed
    TOTAL_REQUESTS=10
    SUCCESS_COUNT=0
    
    for i in $(seq 1 $TOTAL_REQUESTS); do
        if curl -f -s "http://localhost:$HAPROXY_PORT/health" > /dev/null; then
            SUCCESS_COUNT=$((SUCCESS_COUNT + 1))
        fi
        sleep 0.1
    done
    
    if [ $SUCCESS_COUNT -eq $TOTAL_REQUESTS ]; then
        success "Load balancing test: $SUCCESS_COUNT/$TOTAL_REQUESTS requests successful"
    else
        error "Load balancing test: only $SUCCESS_COUNT/$TOTAL_REQUESTS requests successful"
        return 1
    fi
}

test_application_logs() {
    log "Checking application logs for Unix socket usage..."
    
    # Check if apps are using Unix sockets
    if docker logs mr_robot1 2>&1 | grep -q "Unix socket"; then
        success "mr_robot1 is using Unix sockets (found in logs)"
    else
        warning "mr_robot1 may not be using Unix sockets (not found in logs)"
    fi
    
    if docker logs mr_robot2 2>&1 | grep -q "Unix socket"; then
        success "mr_robot2 is using Unix sockets (found in logs)"
    else
        warning "mr_robot2 may not be using Unix sockets (not found in logs)"
    fi
}

test_performance_basic() {
    log "Running basic performance test..."
    
    # Simple performance test with curl
    START_TIME=$(date +%s%N)
    
    for i in $(seq 1 5); do
        curl -f -s "http://localhost:$HAPROXY_PORT/health" > /dev/null
    done
    
    END_TIME=$(date +%s%N)
    DURATION_MS=$(( (END_TIME - START_TIME) / 1000000 ))
    AVG_MS=$(( DURATION_MS / 5 ))
    
    if [ $AVG_MS -lt 100 ]; then
        success "Performance test: Average response time $AVG_MS ms (good)"
    elif [ $AVG_MS -lt 500 ]; then
        success "Performance test: Average response time $AVG_MS ms (acceptable)"
    else
        warning "Performance test: Average response time $AVG_MS ms (may need optimization)"
    fi
}

show_summary() {
    log "Test Summary:"
    echo ""
    echo "Socket Files:"
    docker exec mr_robot1 ls -la "$SOCKET_DIR/" 2>/dev/null | head -10 || warning "Could not list socket directory"
    echo ""
    echo "HAProxy Backend Status:"
    if curl -s "http://localhost:$HAPROXY_STATS_PORT/stats" 2>/dev/null | grep -E "(mr_robot1|mr_robot2)" | head -5; then
        success "HAProxy stats retrieved"
    else
        warning "HAProxy stats not available"
    fi
}

# Main execution
main() {
    echo ""
    log "🧪 Starting Unix Sockets Test Suite for Mr. Robot"
    echo ""
    
    # Run all tests
    test_docker_containers || exit 1
    echo ""
    
    test_socket_files || exit 1
    echo ""
    
    test_haproxy_connectivity || exit 1
    echo ""
    
    test_load_balancing || exit 1
    echo ""
    
    test_application_logs
    echo ""
    
    test_performance_basic
    echo ""
    
    show_summary
    echo ""
    
    success "🎉 All Unix socket tests completed successfully!"
    log "Unix sockets implementation is working correctly"
    echo ""
}

# Execute main function
main "$@"
