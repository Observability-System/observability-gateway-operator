GREEN  := \033[32m
CYAN   := \033[36m
YELLOW := \033[33m
RESET  := \033[0m

IMAGE_NAME := observability-gateway-operator
REGISTRY   := ghcr.io/observability-system/observability-gateway-operator
TAG        := latest
PLATFORMS  := linux/amd64,linux/arm64

.PHONY: help
help:
	@echo "$(CYAN)Targets:$(RESET)"
	@echo "  $(GREEN)docker-build$(RESET)   Build image locally ($(REGISTRY):$(TAG))"
	@echo "  $(GREEN)docker-push$(RESET)    Push single-arch image (requires docker-build first)"
	@echo "  $(GREEN)docker-pushx$(RESET)   Build+push multi-arch image ($(PLATFORMS))"
	@echo ""

.PHONY: docker-build
docker-build:
	docker build -t $(REGISTRY):$(TAG) -f Dockerfile .

.PHONY: docker-push
docker-push: docker-build
	docker push $(REGISTRY):$(TAG)

.PHONY: docker-pushx
docker-pushx:
	docker buildx build --platform $(PLATFORMS) \
		-t $(REGISTRY):$(TAG) -f Dockerfile . --push