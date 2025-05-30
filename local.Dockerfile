FROM alpine:3.14
LABEL maintainer="Myrtea Metrics <contact@myrteametrics.com>"

RUN apk update
RUN apk add --no-cache ca-certificates tzdata
RUN rm -rf /var/cache/apk/*
RUN addgroup -S myrtea -g "1001"
RUN adduser -S myrtea -G myrtea -u "1001"

USER myrtea

WORKDIR /app

COPY bin/myrtea-engine-api myrtea-engine-api
COPY config config
COPY plugin plugin
COPY pkg/plugins/config plugins/config

ENTRYPOINT ["./myrtea-engine-api"]
