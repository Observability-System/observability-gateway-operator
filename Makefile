docker-build:
	docker buildx build --platform linux/amd64,linux/arm64 --tag $(IMG) --push --file Dockerfile .
