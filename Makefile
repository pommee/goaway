DNS_PORT = $(or $(GOAWAY_PORT),53)
WEBSITE_PORT = $(or $(GOAWAY_WEBSITE_PORT),8080)

.PHONY: build start example-queries

build:
	docker build -t goaway \
		--build-arg DNS_PORT=${DNS_PORT} \
		--build-arg WEBSITE_PORT=${WEBSITE_PORT} .

start: build
	DNS_PORT=${DNS_PORT} WEBSITE_PORT=${WEBSITE_PORT} docker compose up goaway -d

example-queries:
	@./dig-domains.sh
