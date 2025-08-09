# Makefile for Mr. Robot Application
# Author: Fabiano Florentino
# Description: Docker commands for managing the mr-robot application

# Include version configuration
include VERSION.mk

# Variables for better maintainability
APP_NAME 					:= mr-robot
DEV_COMPOSE_FILE 	:= docker-compose.dev.yml
PROD_COMPOSE_FILE	:= docker-compose.prod.yml
DOCKERFILE 				:= ./build/Dockerfile
PROCESSOR_DIR 		:= ./infra/payment-processor
DB_CONTAINER 			:= mr_robot_db
DB_USER 					:= mr_robot
DB_NAME 					:= mr_robot
VOLUME_NAME 			:= mr_robot_db

# Colors for output (using printf for better compatibility)
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[1;33m
BLUE := \033[0;34m
PURPLE := \033[0;35m
CYAN := \033[0;36m
NC := \033[0m # No Color

# Shell used for echo commands that support colors
SHELL := /bin/bash

.PHONY: help clean clean-all clean-volumes up down logs restart stats ps validate-docker check-compose help-simple

# Simple help without colors (fallback)
help-simple: ## Show help message without colors
	@echo "Mr. Robot Application - Available Commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' Makefile | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'
	@echo ""
	@echo "Variables:"
	@echo "  IMAGE_NAME: $(FULL_IMAGE_NAME)"
	@echo "  VERSION:    $(VERSION)"
	@echo "  APP_NAME:   $(APP_NAME)"

# Default target
help: ## Show this help message
	@printf "\033[0;34m%s\033[0m\n" "Mr. Robot Application - Available Commands:"
	@printf "\n"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' Makefile | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[0;32m%-20s\033[0m %s\n", $$1, $$2}'
	@printf "\n"
	@printf "\033[1;33m%s\033[0m\n" "Variables:"
	@printf "  IMAGE_NAME: %s\n" "$(FULL_IMAGE_NAME)"
	@printf "  VERSION:    %s\n" "$(VERSION)"
	@printf "  BUILD_DATE: %s\n" "$(BUILD_DATE)"
	@printf "  GIT_COMMIT: %s\n" "$(GIT_COMMIT)"
	@printf "  APP_NAME:   %s\n" "$(APP_NAME)"
	@printf "\n"
	@printf "\033[0;36m%s\033[0m\n" "Quick Start:"
	@printf "  make up      # Start development environment\n"
	@printf "  make down    # Stop development environment\n"
	@printf "  make logs    # View logs\n"
	@printf "  make status  # Check services status\n"
	@printf "\n"

# Validation targets
validate-docker: ## Check if Docker is running
	@printf "\033[0;34m%s\033[0m\n" "Checking Docker status..."
	@docker info > /dev/null 2>&1 || (printf "\033[0;31m%s\033[0m\n" "Docker is not running!" && exit 1)
	@printf "\033[0;32m%s\033[0m\n" "Docker is running"

check-compose: validate-docker ## Validate docker-compose files
	@printf "\033[0;34m%s\033[0m\n" "Validating docker-compose files..."
	@docker compose -f $(DEV_COMPOSE_FILE) config > /dev/null || (printf "\033[0;31m%s\033[0m\n" "Dev compose file has errors!" && exit 1)
	@docker compose -f $(PROD_COMPOSE_FILE) config > /dev/null || (printf "\033[0;31m%s\033[0m\n" "Prod compose file has errors!" && exit 1)
	@printf "\033[0;32m%s\033[0m\n" "Compose files are valid"

#
# Production Environment Commands
prod-down: validate-docker ## Stop and remove production containers with volumes
	@printf "\033[1;33m%s\033[0m\n" "Stopping production environment..."
	docker compose -f $(PROD_COMPOSE_FILE) down --volumes --remove-orphans

prod-up: check-compose ## Start production containers and follow logs
	@printf "\033[0;34m%s\033[0m\n" "Starting production environment..."
	docker compose -f $(PROD_COMPOSE_FILE) up -d && docker compose -f $(PROD_COMPOSE_FILE) logs -f

prod-restart: ## Restart production environment (down + up)
	@printf "\033[1;33m%s\033[0m\n" "Restarting production environment..."
	$(MAKE) prod-down
	$(MAKE) prod-up

prod-logs: validate-docker ## Follow production logs
	docker compose -f $(PROD_COMPOSE_FILE) logs -f

prod-status: validate-docker ## Show production services status
	@printf "\033[0;34m%s\033[0m\n" "Production services status:"
	@docker compose -f $(PROD_COMPOSE_FILE) ps

#
# Development Environment Commands
dev-down: validate-docker ## Stop and remove development containers with volumes
	@printf "\033[1;33m%s\033[0m\n" "Stopping development environment..."
	docker compose -f $(DEV_COMPOSE_FILE) down --volumes --remove-orphans

dev-up: check-compose ## Start development containers and follow logs
	@printf "\033[0;34m%s\033[0m\n" "Starting development environment..."
	docker compose -f $(DEV_COMPOSE_FILE) up -d && docker compose -f $(DEV_COMPOSE_FILE) logs -f

dev-restart: ## Restart development environment (down + up)
	@printf "\033[1;33m%s\033[0m\n" "Restarting development environment..."
	$(MAKE) dev-down
	$(MAKE) dev-up

dev-logs: validate-docker ## Follow development logs
	docker compose -f $(DEV_COMPOSE_FILE) logs -f

dev-status: validate-docker ## Show development services status
	@printf "\033[0;34m%s\033[0m\n" "Development services status:"
	@docker compose -f $(DEV_COMPOSE_FILE) ps

#
# Payment Processor Environment
processor-up: validate-docker ## Start payment processor service
	@printf "\033[0;34m%s\033[0m\n" "Starting payment processor..."
	cd $(PROCESSOR_DIR) && docker compose up -d && docker compose logs -f

processor-down: validate-docker ## Stop payment processor service
	@printf "\033[1;33m%s\033[0m\n" "Stopping payment processor..."
	cd $(PROCESSOR_DIR) && docker compose down --volumes --remove-orphans

processor-status: validate-docker ## Show payment processor status
	@printf "\033[0;34m%s\033[0m\n" "Payment processor status:"
	@cd $(PROCESSOR_DIR) && docker compose ps

#
# Monitoring Commands
stats: validate-docker ## Show Docker container statistics
	@printf "\033[0;34m%s\033[0m\n" "Container statistics:"
	docker stats

ps: validate-docker ## Show running Docker containers
	@printf "\033[0;34m%s\033[0m\n" "Running containers:"
	docker ps --format "table {{.Names}}\t{{.Image}}\t{{.Status}}\t{{.Ports}}"

logs-all: validate-docker ## Show logs from all services
	docker compose -f $(DEV_COMPOSE_FILE) logs -f

# Utility Commands
clean: validate-docker ## Clean up Docker system (remove unused containers, networks, images)
	@printf "\033[1;33m%s\033[0m\n" "Cleaning Docker system..."
	docker system prune -f
	docker volume prune -f
	@printf "\033[0;32m%s\033[0m\n" "Cleanup completed"

clean-all: validate-docker ## Clean up everything including unused images and build cache
	@printf "\033[1;33m%s\033[0m\n" "Performing deep cleanup..."
	docker system prune -a -f
	docker volume prune -f
	docker builder prune -f
	@printf "\033[0;32m%s\033[0m\n" "Deep cleanup completed"

#
# Build Commands
build-dev: validate-docker ## Build development images
	@printf "\033[0;34m%s\033[0m\n" "Building development image..."
	docker build --no-cache $(DOCKER_LABELS) --target development -t $(FULL_IMAGE_NAME)-dev -f $(DOCKERFILE) .
	@printf "\033[0;32m%s\033[0m\n" "Development image built successfully: $(FULL_IMAGE_NAME)-dev"

build-prod: validate-docker ## Build production images
	@printf "\033[0;34m%s\033[0m\n" "Building production image..."
	docker build --no-cache $(DOCKER_LABELS) --target production -t $(FULL_IMAGE_NAME) -f $(DOCKERFILE) .
	@printf "\033[0;32m%s\033[0m\n" "Production image built successfully: $(FULL_IMAGE_NAME)"

build-all: build-dev build-prod ## Build both development and production images

# Dockerfile Commands
dockerfile-stages: ## Show available Dockerfile stages
	@printf "\033[0;34m%s\033[0m\n" "Available Dockerfile stages:"
	@printf "  \033[0;32m%-15s\033[0m %s\n" "base" "Base stage with Go dependencies"
	@printf "  \033[0;32m%-15s\033[0m %s\n" "development" "Development stage with Air"
	@printf "  \033[0;32m%-15s\033[0m %s\n" "prod-build" "Production build stage"
	@printf "  \033[0;32m%-15s\033[0m %s\n" "production" "Production runtime stage"

dockerfile-info: ## Show Dockerfile information
	@printf "\033[0;34m%s\033[0m\n" "Dockerfile Information:"
	@printf "Path: %s\n" "$(DOCKERFILE)"
	@printf "Stages: base -> development | prod-build -> production\n"
	@printf "Dev target: development (includes Air for hot reload)\n"
	@printf "Prod target: production (minimal Alpine image)\n"

# Quick build commands for different scenarios
quick-dev: validate-docker ## Quick development build and run
	@printf "\033[0;34m%s\033[0m\n" "Quick development setup..."
	docker build --target development -t $(APP_NAME):dev -f $(DOCKERFILE) .
	docker run --rm -p 8888:8888 -v "$(shell pwd)":/mr_robot $(APP_NAME):dev

quick-prod: validate-docker ## Quick production build and run
	@printf "\033[0;34m%s\033[0m\n" "Quick production setup..."
	docker build --target production -t $(APP_NAME):prod -f $(DOCKERFILE) .
	docker run --rm -p 8888:8888 $(APP_NAME):prod

# Image Management
image-ls: ## List mr-robot images
	@printf "\033[0;34m%s\033[0m\n" "Mr. Robot images:"
	@docker images $(IMAGE_REGISTRY)/$(IMAGE_NAME) || printf "\033[1;33m%s\033[0m\n" "No images found"

image-clean: ## Remove mr-robot images
	@printf "\033[1;33m%s\033[0m\n" "Removing mr-robot images..."
	@docker rmi $$(docker images $(IMAGE_REGISTRY)/$(IMAGE_NAME) -q) 2>/dev/null || printf "\033[1;33m%s\033[0m\n" "No images to remove"

# Database Commands
db-reset: validate-docker ## Reset database (remove volumes and restart)
	@printf "\033[1;33m%s\033[0m\n" "Resetting database..."
	docker compose -f $(DEV_COMPOSE_FILE) down -v
	docker volume rm $(VOLUME_NAME) 2>/dev/null || true
	docker compose -f $(DEV_COMPOSE_FILE) up -d db
	@printf "\033[0;32m%s\033[0m\n" "Database reset completed"

db-clean: validate-docker ## Clean database tables to fix migration issues
	@printf "\033[1;33m%s\033[0m\n" "Cleaning database tables..."
	docker exec -it $(DB_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME) -c "DROP TABLE IF EXISTS payments CASCADE;" || printf "\033[1;33m%s\033[0m\n" "Table might not exist"

db-reset-full: validate-docker ## Full database reset - clean everything and restart
	@printf "\033[1;33m%s\033[0m\n" "Performing full database reset..."
	docker compose -f $(DEV_COMPOSE_FILE) down --volumes --remove-orphans
	docker volume rm $(VOLUME_NAME) 2>/dev/null || true
	docker compose -f $(DEV_COMPOSE_FILE) up -d
	@printf "\033[0;32m%s\033[0m\n" "Full database reset completed"

db-logs: validate-docker ## Show database logs
	docker compose -f $(DEV_COMPOSE_FILE) logs -f db

db-shell: validate-docker ## Connect to database shell
	@printf "\033[0;34m%s\033[0m\n" "Connecting to database shell..."
	docker exec -it $(DB_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME)

db-registers: validate-docker ## List all registers in the database
	@printf "\033[0;34m%s\033[0m\n" "Latest 15 payment registers:"
	docker exec -it $(DB_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME) -c "SELECT * FROM payments ORDER BY created_at DESC LIMIT 15;"

db-count: validate-docker ## Count all registers in the database
	@printf "\033[0;34m%s\033[0m\n" "Payment count:"
	docker exec -it $(DB_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME) -c "SELECT COUNT(*) as total_payments FROM payments;"

db-status: validate-docker ## Check database status and tables
	@printf "\033[0;34m%s\033[0m\n" "Database tables:"
	docker exec -it $(DB_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME) -c "\dt"

db-backup: validate-docker ## Backup database
	@printf "\033[0;34m%s\033[0m\n" "Creating database backup..."
	@mkdir -p ./backups
	docker exec $(DB_CONTAINER) pg_dump -U $(DB_USER) $(DB_NAME) > ./backups/backup_$(shell date +%Y%m%d_%H%M%S).sql
	@printf "\033[0;32m%s\033[0m\n" "Backup created in ./backups/"

db-restore: validate-docker ## Restore database from backup (usage: make db-restore BACKUP_FILE=backup.sql)
	@if [ -z "$(BACKUP_FILE)" ]; then printf "\033[0;31m%s\033[0m\n" "Please specify BACKUP_FILE=filename"; exit 1; fi
	@printf "\033[0;34m%s\033[0m\n" "Restoring database from $(BACKUP_FILE)..."
	docker exec -i $(DB_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME) < ./backups/$(BACKUP_FILE)
	@printf "\033[0;32m%s\033[0m\n" "Database restored"

# Application Commands
app-shell: validate-docker ## Connect to application container shell
	@printf "\033[0;34m%s\033[0m\n" "Connecting to application container..."
	@docker exec -it mr_robot /bin/sh || printf "\033[0;31m%s\033[0m\n" "Container not running or not found"

app-logs: validate-docker ## Show application logs only
	docker compose -f $(DEV_COMPOSE_FILE) logs -f app1 app2

app-health: validate-docker ## Check application health
	@printf "\033[0;34m%s\033[0m\n" "Checking application health..."
	@curl -s http://localhost:8888/health || printf "\033[0;31m%s\033[0m\n" "Health check failed"

# Quick shortcuts for most used commands
up: dev-up ## Alias for dev-up
down: dev-down ## Alias for dev-down
logs: dev-logs ## Alias for dev-logs
restart: dev-restart ## Alias for dev-restart
status: dev-status ## Alias for dev-status

# Problem-solving commands
fix-volumes: validate-docker ## Fix volume issues by complete reset
	@printf "\033[1;33m%s\033[0m\n" "Performing complete volume reset..."
	docker compose -f $(DEV_COMPOSE_FILE) down --volumes --remove-orphans
	docker volume rm $(VOLUME_NAME) 2>/dev/null || true
	@printf "\033[0;32m%s\033[0m\n" "Volume issues fixed"

clean-volumes: validate-docker ## Clean up volumes
	@printf "\033[1;33m%s\033[0m\n" "Cleaning up volumes..."
	@docker rmi $$(docker images -f "dangling=true" -q) 2>/dev/null || true
	@docker volume rm $$(docker volume ls -qf "dangling=true") 2>/dev/null || true
	docker system prune -f --volumes
	@printf "\033[0;32m%s\033[0m\n" "Volumes cleaned"

# Development workflow commands
dev-full: validate-docker ## Full development setup (build + up)
	@printf "\033[0;34m%s\033[0m\n" "Setting up full development environment..."
	$(MAKE) build-dev
	$(MAKE) dev-up

prod-deploy: validate-docker ## Build and deploy production
	@printf "\033[0;34m%s\033[0m\n" "Deploying to production..."
	$(MAKE) build-prod
	$(MAKE) prod-up

# Environment info
env-info: ## Show environment information
	@printf "\033[0;34m%s\033[0m\n" "Environment Information:"
	@printf "Docker version: %s\n" "$$(docker --version)"
	@printf "Docker Compose version: %s\n" "$$(docker compose version)"
	@printf "Current user: %s\n" "$$(whoami)"
	@printf "Working directory: %s\n" "$$(pwd)"
	@printf "\n"
	@printf "\033[0;34m%s\033[0m\n" "Project Configuration:"
	@printf "App Name: %s\n" "$(APP_NAME)"
	@printf "Full Image: %s\n" "$(FULL_IMAGE_NAME)"
	@printf "Version: %s\n" "$(VERSION)"
	@printf "Build Date: %s\n" "$(BUILD_DATE)"
	@printf "Git Commit: %s\n" "$(GIT_COMMIT)"
	@printf "DB Container: %s\n" "$(DB_CONTAINER)"
	@printf "Volume: %s\n" "$(VOLUME_NAME)"

# Testing commands
test: validate-docker ## Run tests in development container
	@printf "\033[0;34m%s\033[0m\n" "Running tests in development container..."
	@docker exec mr_robot1 go test ./... || printf "\033[0;31m%s\033[0m\n" "Tests failed or container not running"

test-coverage: validate-docker ## Run tests with coverage
	@printf "\033[0;34m%s\033[0m\n" "Running tests with coverage..."
	@docker exec mr_robot1 go test -cover -coverprofile=coverage.out ./... || printf "\033[0;31m%s\033[0m\n" "Coverage test failed or container not running"

test-db-connection: validate-docker ## Test database connection
	@printf "\033[0;34m%s\033[0m\n" "Testing database connection..."
	@docker exec $(DB_CONTAINER) pg_isready -U $(DB_USER) -d $(DB_NAME) && printf "\033[0;32m%s\033[0m\n" "Database connection OK" || printf "\033[0;31m%s\033[0m\n" "Database connection failed"

# Security commands
security-scan: validate-docker ## Run security scan on images
	@printf "\033[0;34m%s\033[0m\n" "Running security scan..."
	@docker run --rm -v /var/run/docker.sock:/var/run/docker.sock \
		aquasec/trivy image $(FULL_IMAGE_NAME) || printf "\033[1;33m%s\033[0m\n" "Error to scanning, trivy not found or image not built"
