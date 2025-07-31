# Makefile for Mr. Robot Application
# Author: Fabiano Florentino
# Description: Docker commands for managing the mr-robot application

.PHONY: help clean clean-all clean-volumes up down logs restart stats ps

# Default target
help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

#
# Production Environment Commands
prod-down: ## Stop and remove production containers with volumes
	docker compose -f docker-compose.prod.yml down --volumes --remove-orphans

prod-up: ## Start production containers and follow logs
	docker compose -f docker-compose.prod.yml up -d && docker compose -f docker-compose.prod.yml logs -f

prod-restart: ## Restart production environment (down + up)
	$(MAKE) prod-down
	$(MAKE) prod-up

prod-logs: ## Follow production logs
	docker compose -f docker-compose.prod.yml logs -f

#
# Development Environment Commands
dev-down: ## Stop and remove development containers with volumes
	docker compose -f docker-compose.dev.yml down --volumes --remove-orphans

dev-up: ## Start development containers and follow logs
	docker compose -f docker-compose.dev.yml up -d && docker compose -f docker-compose.dev.yml logs -f

dev-restart: ## Restart development environment (down + up)
	$(MAKE) dev-down
	$(MAKE) dev-up

dev-logs: ## Follow development logs
	docker compose -f docker-compose.dev.yml logs -f

#
# Payment Processor Environment
processor-up: ## Start payment processor service
	cd ./infra/payment-processor && docker compose up -d && docker compose logs -f

processor-down: ## Stop payment processor service
	cd ./infra/payment-processor && docker compose down --volumes --remove-orphans

#
# Monitoring Commands
stats: ## Show Docker container statistics
	docker stats

ps: ## Show running Docker containers
	docker ps

# Utility Commands
clean: ## Clean up Docker system (remove unused containers, networks, images)
	docker system prune -f
	docker volume prune -f

clean-all: ## Clean up everything including unused images and build cache
	docker system prune -a -f
	docker volume prune -f
	docker builder prune -f

build-dev: ## Build development images
	docker build --no-cache -t fabianoflorentino/mr_robot:v0.0.1 -f ./build/Dockerfile.dev .

build-prod: ## Build production images
	docker build --no-cache -t fabianoflorentino/mr_robot:v0.0.1 -f ./build/Dockerfile.prod .

# Database Commands
db-reset: ## Reset database (remove volumes and restart)
	docker compose -f docker-compose.dev.yml down -v
	docker volume rm mr_robot_db 2>/dev/null || true
	docker compose -f docker-compose.dev.yml up -d db

db-clean: ## Clean database tables to fix migration issues
	docker exec -it mr_robot_db psql -U mr_robot -d mr_robot -c "DROP TABLE IF EXISTS payments CASCADE;"

db-reset-full: ## Full database reset - clean everything and restart
	docker compose -f docker-compose.dev.yml down --volumes --remove-orphans
	docker volume rm mr_robot_db 2>/dev/null || true
	docker compose -f docker-compose.dev.yml up -d

db-logs: ## Show database logs
	docker compose -f docker-compose.dev.yml logs -f db

db-shell: ## Connect to database shell
	docker exec -it mr_robot_db psql -U mr_robot -d mr_robot

db-registers: ## List all registers in the database
	docker exec -it mr_robot_db psql -U mr_robot -d mr_robot -c "SELECT * FROM payments ORDER BY created_at DESC LIMIT 15;"

db-count: ## Count all registers in the database
	docker exec -it mr_robot_db psql -U mr_robot -d mr_robot -c "SELECT COUNT(*) FROM payments;"

db-status: ## Check database status and tables
	docker exec -it mr_robot_db psql -U mr_robot -d mr_robot -c "\dt"

# Application Commands
app-shell: ## Connect to application container shell
	docker exec -it mr_robot /bin/bash

app-logs: ## Show application logs only
	docker compose -f docker-compose.dev.yml logs -f app1 app2

# Quick shortcuts for most used commands
up: dev-up ## Alias for dev-up
down: dev-down ## Alias for dev-down
logs: dev-logs ## Alias for dev-logs
restart: dev-restart ## Alias for dev-restart

# Problem-solving commands
fix-volumes: ## Fix volume issues by complete reset
	@echo "Performing complete volume reset..."
	docker compose -f docker-compose.dev.yml down --volumes --remove-orphans
	docker volume rm mr_robot_db 2>/dev/null || true

clean-volumes: ## Clean up volumes
	@echo "Cleaning up volumes..."
	docker rmi $(docker images -f "dangling=true" -q) 2>/dev/null || true
	docker volume rm $(docker volume ls -qf "dangling=true") 2>/dev/null || true
	docker system prune -f --volumes
