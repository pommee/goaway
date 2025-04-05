FROM ubuntu:22.04

ARG DNS_PORT=53
ARG WEBSITE_PORT=8080

ENV DNS_PORT=${DNS_PORT}
ENV WEBSITE_PORT=${WEBSITE_PORT}

RUN apt-get update && \
    apt-get install -y curl passwd jq sudo net-tools && \
    rm -rf /var/lib/apt/lists/*

RUN useradd -m -s /bin/bash -G sudo appuser && \
    echo "appuser ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers

WORKDIR /home/appuser

COPY updater.sh /home/appuser/updater.sh
RUN chmod +x /home/appuser/updater.sh

RUN curl https://raw.githubusercontent.com/pommee/goaway/main/installer.sh | sh /dev/stdin && \
    mv /root/.local/bin/goaway /home/appuser/goaway && \
    chmod +x /home/appuser/goaway

COPY start.sh /home/appuser/start.sh
RUN chmod +x /home/appuser/start.sh
RUN chown -R appuser:appuser /home/appuser

EXPOSE ${DNS_PORT}/tcp ${DNS_PORT}/udp ${WEBSITE_PORT}/tcp

USER appuser

CMD ["/home/appuser/start.sh"]
