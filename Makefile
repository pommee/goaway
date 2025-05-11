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
	pnpm -C client install && pnpm -C client build

start:
	docker compose up -d

lint:
	pnpm -C client lint && \
	golangci-lint run ./backend/...

format:
	npx prettier --write "client/**/*.{html,css,js,tsx}"

example-queries:
	@./test/dig-domains.sh

dev-website:
	pnpm -C client install && pnpm -C client dev

dev-server:
	mkdir client/dist ; touch client/dist/.fake
	air .

test:
	go test -count=1 -bench=. -benchmem ./test/...

bench:
	go run test/benchmark.go -test.bench=.

bench-profile:
	go run test/benchmark.go -test.bench=. & \
	go tool pprof http://localhost:6060/debug/pprof/profile\?seconds\=5
