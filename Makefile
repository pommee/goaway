.PHONY: build start example-queries logs

DNS_PORT = $(or $(GOAWAY_PORT),53)
WEBSITE_PORT = $(or $(GOAWAY_WEBSITE_PORT),8080)
VERSION = $(or $(GOAWAY_VERSION),latest)

build:
	docker build -t pommee/goaway:${VERSION} \
		--build-arg DNS_PORT=${DNS_PORT} \
		--build-arg WEBSITE_PORT=${WEBSITE_PORT} \
		.

publish: build
	docker tag pommee/goaway:${VERSION} pommee/goaway:latest
	docker push pommee/goaway:${VERSION}
	docker push pommee/goaway:latest

start: build
	DNS_PORT=${DNS_PORT} \
	WEBSITE_PORT=${WEBSITE_PORT} \
	docker compose up goaway -d

lint:
	golangci-lint run

example-queries:
	@./testing/dig-domains.sh
