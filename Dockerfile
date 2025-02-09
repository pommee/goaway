FROM ubuntu:22.04

ARG DNS_PORT=53
ARG WEBSITE_PORT=8080

ENV DNS_PORT=${DNS_PORT}
ENV WEBSITE_PORT=${WEBSITE_PORT}

RUN apt-get update && \
    apt-get install -y curl passwd && \
    adduser --disabled-password --gecos "" appuser

WORKDIR /app

RUN chown appuser:appuser /app

RUN curl https://raw.githubusercontent.com/pommee/goaway/main/installer.sh | sh /dev/stdin && \
    mv /root/.local/bin/goaway /app/goaway && \
    chmod +x /app/goaway && \
    chown appuser:appuser /app/goaway

EXPOSE ${DNS_PORT}/tcp ${DNS_PORT}/udp ${WEBSITE_PORT}/tcp

USER appuser

CMD ["sh", "-c", "/app/goaway", "--dnsport=${DNS_PORT}", "--webserverport=${WEBSITE_PORT}"]
