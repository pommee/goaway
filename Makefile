.PHONY: publish lint example-queries dev

DNS_PORT = $(or $(GOAWAY_PORT),53)
WEBSITE_PORT = $(or $(GOAWAY_WEBSITE_PORT),8080)
VERSION = $(or $(GOAWAY_VERSION),latest)

publish:
	docker buildx create --name multiarch-builder --use

	docker buildx build \
	--platform linux/amd64 \
	--tag pommee/goaway:${VERSION} \
	.

	docker buildx rm multiarch-builder

lint:
	golangci-lint run

example-queries:
	@./testing/dig-domains.sh

dev:
	@air
