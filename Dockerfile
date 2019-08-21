FROM golang:1.12.1-alpine3.9 AS builder

COPY . /go/src/github.com/huawei-cloudnative/ci-bot

RUN apk --no-cache update && \
apk --no-cache upgrade && \
CGO_ENABLED=1 go build -v -o /usr/local/bin/ci-bot -ldflags="-w -s -extldflags -static" \
github.com/huawei-cloudnative/ci-bot

ENTRYPOINT ["ci-bot"]

