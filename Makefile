.PHONY: publish lint example-queries dev

DNS_PORT = $(or $(GOAWAY_PORT),53)
WEBSITE_PORT = $(or $(GOAWAY_WEBSITE_PORT),8080)
VERSION = $(or $(GOAWAY_VERSION),latest)

publish:
	docker buildx create --name multiarch-builder --use

	docker buildx build \
	--platform linux/amd64 \
	--tag pommee/goaway:${VERSION} \
	--push \
	.

	docker buildx rm multiarch-builder

start-dev:
	docker build --target dev-build -t goaway-dev .
	docker run --rm --name goaway-dev -p 8080:8080 -p 6121:6121 -v ./:/app goaway-dev

start-prod:
	docker build --target prod -t goaway .

lint:
	golangci-lint run

example-queries:
	@./testing/dig-domains.sh

dev:
	@air
