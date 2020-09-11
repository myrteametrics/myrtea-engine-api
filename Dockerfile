# Stage 1 - Build binary
FROM golang:1.14-alpine as builder
LABEL maintainer="Myrtea Metrics <contact@myrteametrics.com>"

RUN apk --no-cache add curl git make \
    && rm -rf /var/cache/apk/*

WORKDIR /build
COPY internals internals
COPY main.go ./

RUN make swag
RUN make build


# Stage 2 - Run binary
FROM alpine:3.7
LABEL maintainer="Myrtea Metrics <contact@myrteametrics.com>"

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

COPY --from=builder /build/bin/myrtea-engine-api myrtea-engine-api
COPY config config
COPY certs certs

ENTRYPOINT ["./myrtea-engine-api"]
