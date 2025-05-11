#!/bin/bash

run_goaway() {
    while true; do
        echo "Starting goaway..."
        /home/appuser/goaway --dns-port=${DNS_PORT} --webserver-port=${WEBSITE_PORT}
        echo "goaway process exited with code $?. Restarting..."
    done
}

trap 'echo "Received SIGTERM, shutting down..."; exit 0' SIGTERM
run_goaway &

wait