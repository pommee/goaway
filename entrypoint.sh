#!/bin/sh

if [ -z "$DNS_PORT" ] || [ -z "$WEBSITE_PORT" ]; then
  echo "Error: DNS_PORT and WEBSITE_PORT environment variables must be set."
  exit 1
fi

exec /app/goaway --dnsport "$DNS_PORT" --webserverport "$WEBSITE_PORT"
