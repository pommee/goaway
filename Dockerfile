FROM golang:1.23-alpine AS builder

ARG DNS_PORT
ARG WEBSITE_PORT

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /dns-server

FROM alpine:3.18

ARG DNS_PORT
ARG WEBSITE_PORT

RUN adduser -D appuser && \
    apk add --no-cache libcap

WORKDIR /app
COPY --from=builder /dns-server .
COPY blacklist.json .

RUN chown -R appuser:appuser /app && \
    setcap 'cap_net_bind_service=+ep' /app/dns-server

ENV DNS_PORT=${DNS_PORT} \
    WEBSITE_PORT=${WEBSITE_PORT}

EXPOSE $DNS_PORT/tcp $DNS_PORT/udp $WEBSITE_PORT/tcp

USER appuser
ENTRYPOINT ["/app/dns-server"]
