FROM ubuntu:22.04

WORKDIR /app
ADD minlog /app/minlog
ADD startup.sh /app/startup.sh


CMD ["sh", "-c", "/app/startup.sh"]
