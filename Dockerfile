FROM golang:alpine

RUN apk add --no-cache git

ADD . /go/src/github.com/fankserver/docker-factorio-watchdog
RUN go get github.com/fankserver/docker-factorio-watchdog/... \
    && go install github.com/fankserver/docker-factorio-watchdog
ENTRYPOINT /go/bin/docker-factorio-watchdog