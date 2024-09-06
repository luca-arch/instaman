# Golang app builder
FROM golang:1.23.0-alpine3.20 AS builder

WORKDIR /mnt/src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd cmd/
COPY database database/
COPY instaproxy instaproxy/
COPY internal internal/
COPY service service/
COPY webserver webserver/
RUN go build -o api-server ./cmd/api-server/main.go


# Golang app runner
FROM alpine:3.20.2

ENV GOMAXPROCS="1"
ENV ISDOCKER="1"
COPY --from=builder /mnt/src/api-server /srv/api-server

EXPOSE 10000

ENTRYPOINT [ "/srv/api-server" ]