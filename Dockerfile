FROM golang:1.12-alpine AS build
WORKDIR /go/src/github.com/fankserver/docker-factorio-watchdog
RUN apk add --no-cache alpine-sdk
COPY . .
RUN go get ./... && \
  go build -a -installsuffix cgo -o app .

FROM alpine
RUN adduser -D -u 678 watchdog && \
  apk add --no-cache alpine-sdk
USER watchdog
COPY --from=build /go/src/github.com/fankserver/docker-factorio-watchdog/app /app
ENTRYPOINT ["/app"]
