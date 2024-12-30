#!/bin/bash

domains=("pagead-googlehosted.l.google.com." "pixel.rubiconproject.com." "cdn.adsafeprotected.com." "google.com")
dns_server_ip=${GOAWAY_IP}
dns_server_port=${GOAWAY_PORT:-53}

if [ -z "${dns_server_ip}" ]; then
    echo "GOAWAY_IP not set, quitting."
    exit 1
fi

echo "Sending ${#domains[@]} requests..."
for domain in "${domains[@]}"; do
    dig +short @$dns_server_ip -p $dns_server_port $domain
done
