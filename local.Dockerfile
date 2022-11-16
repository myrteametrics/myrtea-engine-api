FROM alpine:3.14
LABEL maintainer="Myrtea Metrics <contact@myrteametrics.com>"

RUN apk update && apk add -y ca-certificates tzdata && rm -rf /var/cache/apk/*
RUN addgroup -S myrtea -g "1001" &&  \
    adduser -S myrtea -G myrtea -u "1001"

USER myrtea

WORKDIR /app

COPY bin/myrtea-engine-api myrtea-engine-api
COPY config config

ENTRYPOINT ["./myrtea-engine-api"]
