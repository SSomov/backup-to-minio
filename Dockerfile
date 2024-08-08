FROM alpine:latest

RUN apk add --no-cache curl tar gzip yq \
    && curl -O https://dl.min.io/client/mc/release/linux-amd64/mc \
    && chmod +x mc \
    && mv mc /usr/local/bin/

COPY backup.sh /usr/local/bin/backup.sh
COPY backup.yml /backup.yml

RUN chmod +x /usr/local/bin/backup.sh

ENTRYPOINT ["/usr/local/bin/backup.sh"]
