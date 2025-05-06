FROM ubuntu:22.04

ENV APPLOGSDIR /var/log/containers
ENV LOKIEP http://loki:3100/loki/api/v1/push
ENV GRAFANAALLOYURL http://127.0.0.1:12345

WORKDIR /app

ADD minlog /app/minlog
ADD startup.sh /app/startup.sh


ENV GRAFANAALLOYFILE /app/confs/alloy.alloy

CMD ["sh", "-c", "/app/startup.sh"]
