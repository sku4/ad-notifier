FROM golang:1.20.5-alpine3.18 AS builder

RUN go version

COPY . /ad-notifier/
WORKDIR /ad-notifier/

ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0

RUN go mod download
RUN go build -o ./.bin/ad-notifier -tags=go_tarantool_ssl_disable ./cmd/notifier/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /ad-notifier/.bin/ad-notifier .
COPY --from=builder /ad-notifier/configs/config.yml configs/config.yml
COPY --from=builder /ad-notifier/internal/sender/templates internal/sender/templates/
COPY --from=builder /ad-notifier/static static/
RUN touch .env

CMD /app/ad-notifier
