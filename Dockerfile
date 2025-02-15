FROM ubuntu:22.04 AS dev-build

ARG DNS_PORT=53
ARG WEBSITE_PORT=8080

ENV DNS_PORT=${DNS_PORT}
ENV WEBSITE_PORT=${WEBSITE_PORT}

RUN apt update && \
    apt install -y build-essential curl git

RUN curl -sSL https://golang.org/dl/go1.23.2.linux-amd64.tar.gz | tar -xz -C /usr/local

ENV PATH=$PATH:/usr/local/go/bin
ENV GOPATH=/home/appuser/go
ENV PATH=$PATH:$GOPATH/bin

RUN adduser --disabled-password --gecos "" appuser

USER appuser
WORKDIR /home/appuser

RUN go install github.com/air-verse/air@latest

WORKDIR /app

USER root
RUN chown -R appuser:appuser /app
USER appuser

COPY . .

EXPOSE ${DNS_PORT}/tcp ${DNS_PORT}/udp ${WEBSITE_PORT}/tcp

ENTRYPOINT ["air"]


FROM ubuntu:22.04 AS prod-build

ARG DNS_PORT=53
ARG WEBSITE_PORT=8080

ENV DNS_PORT=${DNS_PORT}
ENV WEBSITE_PORT=${WEBSITE_PORT}

RUN apt-get update && \
    apt-get install -y curl passwd

RUN adduser --disabled-password --gecos "" appuser

WORKDIR /app

RUN curl https://raw.githubusercontent.com/pommee/goaway/main/installer.sh | sh /dev/stdin && \
    mv /root/.local/bin/goaway /app/goaway && \
    chmod +x /app/goaway && \
    chown appuser:appuser /app/goaway

EXPOSE ${DNS_PORT}/tcp ${DNS_PORT}/udp ${WEBSITE_PORT}/tcp

USER appuser

CMD ["sh", "-c", "/app/goaway", "--dnsport=${DNS_PORT}", "--webserverport=${WEBSITE_PORT}"]
