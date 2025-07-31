# Configuração de Versão para Makefiles
# Este arquivo usa sintaxe específica do Make
# Para usar: include VERSION.mk no seu Makefile

# Versão base
VERSION := 0.0.1
IMAGE_TAG := v$(VERSION)

# Data de build (sintaxe Makefile)
BUILD_DATE := $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')

# Commit do Git (sintaxe Makefile)
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Configurações da Imagem Docker
IMAGE_REGISTRY := fabianoflorentino
IMAGE_NAME := mr_robot
FULL_IMAGE_NAME := $(IMAGE_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)

# Labels padrão para imagens Docker (sintaxe Makefile)
DOCKER_LABELS := --label "version=$(VERSION)" \
                --label "build-date=$(BUILD_DATE)" \
                --label "git-commit=$(GIT_COMMIT)" \
                --label "maintainer=fabiano.florentino"

# Exportar variáveis para sub-makes
export VERSION
export IMAGE_TAG
export BUILD_DATE
export GIT_COMMIT
export IMAGE_REGISTRY
export IMAGE_NAME
export FULL_IMAGE_NAME
export DOCKER_LABELS
