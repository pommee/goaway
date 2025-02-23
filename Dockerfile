FROM ubuntu:22.04

ARG DNS_PORT=53
ARG WEBSITE_PORT=8080

ENV DNS_PORT=${DNS_PORT}
ENV WEBSITE_PORT=${WEBSITE_PORT}

RUN apt-get update && \
    apt-get install -y curl passwd jq

WORKDIR /root

COPY updater.sh /root/updater.sh
RUN chmod +x /root/updater.sh

RUN curl https://raw.githubusercontent.com/pommee/goaway/main/installer.sh | sh /dev/stdin && \
    mv /root/.local/bin/goaway /root/goaway && \
    chmod +x /root/goaway

EXPOSE ${DNS_PORT}/tcp ${DNS_PORT}/udp ${WEBSITE_PORT}/tcp

CMD ["sh", "-c", "/root/goaway", "--dnsport=${DNS_PORT}", "--webserverport=${WEBSITE_PORT}"]
