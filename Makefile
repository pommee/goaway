.PHONY: publish lint example-queries dev format test

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

start:
	docker compose up -d

lint:
	golangci-lint run

format:
	npx prettier --write "website/*.{html,css,js}"

example-queries:
	@./testing/dig-domains.sh

dev:
	@air

test:
	go test -count=1 -bench=. -benchmem ./test/...
