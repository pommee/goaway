FROM golang:1.23-alpine AS builder

ARG DNS_PORT
ARG WEBSITE_PORT

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /goaway

FROM alpine:3.18

ARG DNS_PORT
ARG WEBSITE_PORT

RUN adduser -D appuser && \
    apk add --no-cache libcap

WORKDIR /app
COPY --from=builder /goaway .
COPY blacklist.json .
COPY settings.json .

RUN chown -R appuser:appuser /app && \
    setcap 'cap_net_bind_service=+ep' /app/goaway

EXPOSE ${DNS_PORT}/tcp ${DNS_PORT}/udp ${WEBSITE_PORT}/tcp

USER appuser

COPY entrypoint.sh /app/entrypoint.sh

ENTRYPOINT ["/app/entrypoint.sh"]
