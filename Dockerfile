FROM alpine:3.22

ARG GOAWAY_VERSION=""
ARG DNS_PORT=53
ARG WEBSITE_PORT=8080

ENV DNS_PORT=${DNS_PORT} WEBSITE_PORT=${WEBSITE_PORT}

COPY installer.sh ./

RUN apk add --no-cache curl jq bash ca-certificates && \
    adduser -D -s /bin/bash appuser && \
    ./installer.sh $GOAWAY_VERSION && \
    mv /root/.local/bin/goaway /home/appuser/goaway && \
    chown -R appuser:appuser /home/appuser && \
    rm -rf /var/cache/apk/* /tmp/* /var/tmp/* /root/.cache /root/.local installer.sh

WORKDIR /home/appuser

COPY updater.sh start.sh ./

EXPOSE ${DNS_PORT}/tcp ${DNS_PORT}/udp ${WEBSITE_PORT}/tcp

USER appuser

CMD ["./start.sh"]
