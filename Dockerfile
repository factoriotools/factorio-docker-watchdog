FROM golang:1.13-alpine AS build
WORKDIR /go/src/github.com/factoriotools/factorio-docker-watchdog
RUN apk add --no-cache --no-progress g++ git
COPY . .
RUN go get ./... && \
  go build -a -installsuffix cgo -o app .

FROM alpine:3.10
RUN adduser -D -u 678 watchdog && \
  apk add --no-cache --no-progress git
USER watchdog
COPY --from=build /go/src/github.com/factoriotools/factorio-docker-watchdog/app /app
CMD ["/app"]
