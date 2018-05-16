FROM golang:1.10-alpine AS build
WORKDIR /go/src/github.com/fankserver/docker-factorio-watchdog
COPY . .
RUN apk add --no-cache alpine-sdk \
    && go get ./... \
    && go build -a -installsuffix cgo -o app .

FROM alpine:latest
RUN adduser -D -u 678 watchdog && apk add --no-cache alpine-sdk
USER watchdog

# Add app
COPY --from=build /go/src/github.com/fankserver/docker-factorio-watchdog/app /app

# This container will be executable
ENTRYPOINT ["/app"]