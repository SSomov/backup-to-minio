FROM alpine:latest

RUN apk add --no-cache bash curl tar gzip yq \
    && curl -O https://dl.min.io/client/mc/release/linux-amd64/mc \
    && chmod +x mc \
    && mv mc /usr/local/bin/

WORKDIR /app

COPY backup.sh /app/backup.sh
# COPY backup.yml /backup.yml

RUN chmod +x /app/backup.sh

ENTRYPOINT ["sh","-c","/app/backup.sh"]
