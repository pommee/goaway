#!/bin/bash

dns_server_ip=${GOAWAY_IP:-localhost}
dns_server_port=${GOAWAY_PORT:-53}
DOMAIN="example.com"
EXPECTED_RESPONSE="127.0.0.1"

RECORD_TYPES=(A AAAA CNAME TXT MX NS SOA PTR SRV)

check_dns() {
    local type=$1
    echo "Checking $type record for $DOMAIN"
    RESPONSE=$(dig @$dns_server_ip -p $dns_server_port $DOMAIN $type +short)

    if [[ -z "$RESPONSE" ]]; then
        echo "[ERROR] No response for $type record"
    else
        echo "Response for $type record: $RESPONSE"
    fi
    echo "---------------------------"
}

for TYPE in "${RECORD_TYPES[@]}"; do
    check_dns "$TYPE"
done
