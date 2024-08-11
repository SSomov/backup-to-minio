FROM golang:1.22.6-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o backup-tool .

FROM alpine:3.18

RUN apk --no-cache add ca-certificates docker-cli

COPY --from=build /app/backup-tool /usr/local/bin/backup-tool

# ENTRYPOINT ["/usr/local/bin/backup-tool"]