FROM golang:1.23.1-alpine AS build

WORKDIR /app

RUN apk add --no-cache git gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o backup-tool ./cmd

FROM alpine:3.21.3

RUN apk add --no-cache postgresql17-client mysql-client ca-certificates

COPY --from=build /app/backup-tool /backup-tool

CMD ["/backup-tool"]