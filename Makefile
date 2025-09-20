# Makefile for development and production workflows
# Targets:
#   make dev        - start development compose (uses compose.override.yml)
#   make dev-down   - stop development compose
#   make prod-up    - start production compose (detached)
#   make prod-down  - stop production compose
#   make build-dev  - build images with compose override
#   make build-prod - build images for production compose
#   make build-images - docker build images locally for services
#   make push       - push built images to registry (requires login)
#   make fmt        - run go fmt in modules
#   make test       - run go test in modules
#   make shell      - open shell in running service (SERVICE=...)

COMPOSE_FILES_DEV := -f compose.yml -f compose.override.yml
COMPOSE_FILES_PROD := -f compose.yml
COMPOSE := docker compose
SERVICES := polling-service results-service
IMAGE_PREFIX ?= lekob/pub-sub-playground
TAG ?= latest
SERVICE ?= polling-service

.DEFAULT_GOAL := prod-up

.PHONY: help dev dev-down prod-up prod-down build-dev build-prod build-images push fmt test shell clean-images

help:
	@echo "Makefile targets:"
	@echo "  make dev                 # start dev environment (compose with override)"
	@echo "  make dev-down            # stop dev environment"
	@echo "  make prod-up             # start production environment (detached)"
	@echo "  make prod-down           # stop production environment"
	@echo "  make build-dev           # build dev images (compose with override)"
	@echo "  make build-prod          # build prod images (compose.yml)"
	@echo "  make build-images        # docker build images locally"
	@echo "  make push                # push built images to registry (requires login)"
	@echo "  make fmt                 # go fmt for all modules"
	@echo "  make test                # run go test in each module"
	@echo "  make shell SERVICE=<svc> # open a shell in a running service"

## Development (uses compose.override.yml)
dev:
	$(COMPOSE) $(COMPOSE_FILES_DEV) up --build

dev-down:
	$(COMPOSE) $(COMPOSE_FILES_DEV) down

## Production (compose.yml only)
prod-up:
	$(COMPOSE) $(COMPOSE_FILES_PROD) up -d --build

prod-down:
	$(COMPOSE) $(COMPOSE_FILES_PROD) down

build-dev:
	$(COMPOSE) $(COMPOSE_FILES_DEV) build

build-prod:
	$(COMPOSE) $(COMPOSE_FILES_PROD) build

## Build images directly with docker build. Useful to tag/push.
build-images:
	@echo "Building Docker images for services: $(SERVICES)"
	@for svc in $(SERVICES); do \
		echo "-> building $$svc"; \
		docker build -t $(IMAGE_PREFIX)/$$svc:$(TAG) ./$$svc; \
	done

push: build-images
	@echo "Pushing images with tag $(TAG)"
	@for svc in $(SERVICES); do \
		echo "-> pushing $$svc"; \
		docker push $(IMAGE_PREFIX)/$$svc:$(TAG); \
	done

fmt:
	@for d in common polling-service results-service; do \
		(cd $$d && go fmt ./...); \
	done

test:
	@for d in common polling-service results-service; do \
		(cd $$d && go test ./...); \
	done

shell:
	@$(COMPOSE) $(COMPOSE_FILES_DEV) exec $(SERVICE) sh

clean-images:
	@for svc in $(SERVICES); do \
		docker image rm -f $(IMAGE_PREFIX)/$$svc:$(TAG) || true; \
	done
