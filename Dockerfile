FROM alpine:3.18

ARG DNS_PORT=53
ARG WEBSITE_PORT=8080

WORKDIR /app

RUN apk add curl && \
    curl https://raw.githubusercontent.com/pommee/goaway/main/installer.sh | sh /dev/stdin && \
    mv /root/.local/bin/ /app/goaway

EXPOSE ${DNS_PORT}/tcp ${DNS_PORT}/udp ${WEBSITE_PORT}/tcp

CMD ["sh", "-c", "/app/goaway --dnsport=${DNS_PORT} --webserverport=${WEBSITE_PORT}"]
