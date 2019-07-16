FROM docker.io/golang:1.12.4-stretch as builder

COPY . /specs-bot
WORKDIR /specs-bot

RUN go build -v

# Use a separate image for building and running specs-bot, as it makes the size of the image drop from
# 820M to 25M
FROM docker.io/alpine:latest

MAINTAINER Informo core team <core@informo.network>

# Add SQLite support, see https://github.com/olivere/sqlite-docker
RUN apk --update upgrade && \
    apk add sqlite && \
    rm -rf /var/cache/apk/*
# See https://stackoverflow.com/questions/34729748/installed-go-binary-not-found-in-path-on-alpine-linux-docker
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

COPY --from=builder /specs-bot/specs-bot /usr/local/bin/specs-bot
RUN mkdir -p /etc/specs-bot
COPY ./docker/data /etc/specs-bot

EXPOSE 8080/tcp
VOLUME ["/etc/specs-bot"]

WORKDIR /etc/specs-bot

ENTRYPOINT ["/usr/local/bin/specs-bot", "--config", "/etc/specs-bot/config.yaml", "--debug"]
