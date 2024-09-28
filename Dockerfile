FROM golang:1.22.6-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o backup-tool .

FROM alpine:3.18

RUN apk --no-cache add ca-certificates

COPY --from=build /app/backup-tool /backup-tool

CMD ["/backup-tool"]