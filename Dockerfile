FROM ubuntu:22.04

ENV APPLOGSDIR=/var/log/containers
ENV LOKIEP=http://loki:3100/loki/api/v1/push
ENV ALLOYURL=http://127.0.0.1:12345
ENV ALLOYFILE=/etc/alloy/config.alloy

WORKDIR /app

ADD minlog /app/minlog
ADD startup.sh /app/startup.sh



ENTRYPOINT ["/usr/bin/bash", "./startup.sh"]
