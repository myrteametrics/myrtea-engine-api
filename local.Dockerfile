FROM alpine:3.14
LABEL maintainer="Mind7 Consulting <contact@mind7.com>"

RUN apk update && apk add --no-cache ca-certificates tzdata && rm -rf /var/cache/apk/*
RUN addgroup -S myrtea -g "1001" &&  \
    adduser -S myrtea -G myrtea -u "1001"

USER myrtea

WORKDIR /app

COPY bin/myrtea-engine-api myrtea-engine-api
COPY config config

ENTRYPOINT ["./myrtea-engine-api"]
