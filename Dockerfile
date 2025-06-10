FROM alpine:3.22

ARG GOAWAY_VERSION=""
ARG DNS_PORT=53
ARG WEBSITE_PORT=8080

ENV DNS_PORT=${DNS_PORT} WEBSITE_PORT=${WEBSITE_PORT}

COPY installer.sh ./

RUN apk add --no-cache curl jq bash ca-certificates && \
    adduser -D -s /bin/bash appuser && \
    mkdir -p /app && \
    ./installer.sh $GOAWAY_VERSION && \
    mv /root/.local/bin/goaway /app/goaway && \
    chown -R appuser:appuser /app && \
    rm -rf /var/cache/apk/* /tmp/* /var/tmp/* /root/.cache /root/.local installer.sh

WORKDIR /app

COPY updater.sh ./
RUN chown appuser:appuser updater.sh

EXPOSE ${DNS_PORT}/tcp ${DNS_PORT}/udp ${WEBSITE_PORT}/tcp

USER appuser

ENTRYPOINT [ "./goaway" ]
