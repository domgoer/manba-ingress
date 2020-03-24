FROM golang:1.13.1-alpine AS builder
WORKDIR /manba-ingress
COPY go.mod .
COPY go.sum .
ENV GOPROXY=https://goproxy.cn GO111MODULE=on
RUN go mod download
COPY . .
RUN go build -o manba-ingress ./cmd

FROM alpine:3.11.3

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
    && apk add --no-cache --update ca-certificates

COPY --from=builder /manba-ingress/manba-ingress /manba-ingress
ENTRYPOINT ["/manba-ingress"]
