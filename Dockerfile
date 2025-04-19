FROM golang:1.24.2-alpine AS builder

RUN set -xe && \
    apk upgrade --update-cache --available && \
    apk add  alpine-sdk openssh && \
    rm -rf /var/cache/apk/*

WORKDIR /builder
COPY go.mod .
COPY go.sum .

ENV GO111MODULE=on
RUN go mod download

COPY . .
RUN make build

FROM alpine:latest
RUN set -xe && \
    apk upgrade --update-cache --available && \
    rm -rf /var/cache/apk/*

RUN adduser -g ramonmedeiros -u 1890 -D ramonmedeiros
COPY --from=builder  /builder/bin /home/ramonmedeiros/bin
RUN chown -R ramonmedeiros:ramonmedeiros /home/ramonmedeiros
USER 1890
WORKDIR /home/ramonmedeiros

ARG PORT 
EXPOSE $PORT 
ENV GIN_MODE=release

ENV PATH="/home/ramonmedeiros/bin:${PATH}"
CMD ["bin/rsvp"]
