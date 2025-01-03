FROM golang:1.23-alpine AS builder

ARG DNS_PORT
ARG WEBSITE_PORT

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

RUN apk add --no-cache gcc musl-dev

COPY . .

RUN CGO_ENABLED=1 go build -trimpath -ldflags="-w -s" -o /goaway

FROM alpine:3.18

ARG DNS_PORT
ARG WEBSITE_PORT

RUN adduser -D appuser && \
    apk add --no-cache libcap

WORKDIR /app

COPY --from=builder /goaway .

RUN [ -f database.db ] && cp database.db /app/ || echo "No database found." \
    && [ -f settings.json ] && cp settings.json /app/ || echo "No settings found."

RUN chown -R appuser:appuser /app && \
    setcap 'cap_net_bind_service=+ep' /app/goaway

EXPOSE ${DNS_PORT}/tcp ${DNS_PORT}/udp ${WEBSITE_PORT}/tcp

USER appuser

CMD ["sh", "-c", "/app/goaway --dnsport=${DNS_PORT} --webserverport=${WEBSITE_PORT}"]
