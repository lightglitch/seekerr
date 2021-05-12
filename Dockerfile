# GitHub:       https://github.com/lightglitch/seekerr

FROM golang:1.16-alpine3.13 AS build

ARG CGO=1
ENV CGO_ENABLED=${CGO}
ENV GOOS=linux
ENV GO111MODULE=on

WORKDIR /go/src/github.com/lightglitch/seekerr

COPY . /go/src/github.com/lightglitch/seekerr/

RUN apk update

RUN go build -o seekerr main.go

# ---

FROM  python:3.7-alpine3.13

RUN pip install guessit

ENV \
  SEEKERR_CONFIG_PATH=/config/ \
  SEEKERR_LOGGER_FILE="/config/log/seekerr.%Y%m%d.log"

COPY --from=build /go/src/github.com/lightglitch/seekerr/seekerr /usr/bin/seekerr

# ca-certificates are required to fetch outside resources
RUN apk update && \
    apk add --no-cache ca-certificates

RUN mkdir /config
COPY --from=build /go/src/github.com/lightglitch/seekerr/config/seekerr.sample.yaml /config/seekerr.sample.yaml

# Config volume
VOLUME /config

ENTRYPOINT ["seekerr"]
CMD ["cron"]
