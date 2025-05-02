#!/bin/bash

domains=("pagead-googlehosted.l.google.com." "pixel.rubiconproject.com." "cdn.adsafeprotected.com." "google.com")
dns_server_ip=${GOAWAY_IP}
dns_server_port=${GOAWAY_PORT:-53}

if [ -z "${dns_server_ip}" ]; then
    echo "GOAWAY_IP not set, quitting."
    exit 1
fi

success_count=0
fail_count=0

echo "Sending ${#domains[@]} requests..."
for domain in "${domains[@]}"; do
    result=$(dig +short @$dns_server_ip -p $dns_server_port $domain)

    if [[ "$result" == "NXDOMAIN" || -z "$result" ]]; then
        ((fail_count++))
    else
        ((success_count++))
    fi
done

echo "$success_count requests succeeded."
echo "$fail_count requests failed."
