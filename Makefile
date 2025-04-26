.PHONY: publish build lint example-queries dev format test bench

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

build:
	pnpm -C website install && pnpm -C website build

start:
	docker compose up -d

lint:
	pnpm -C website lint && \
	golangci-lint run

format:
	npx prettier --write "website/**/*.{html,css,js,tsx}"

example-queries:
	@./testing/dig-domains.sh

dev-website:
	pnpm -C website install && pnpm -C website dev

dev-server:
	mkdir website/dist ; touch website/dist/.fake
	air .

test:
	go test -count=1 -bench=. -benchmem ./test/...

bench:
	go run testing/benchmark.go -test.bench=.
