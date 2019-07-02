FROM golang:1.12-alpine AS build
WORKDIR /go/src/github.com/factoriotools/factorio-docker-watchdog
RUN apk add --no-cache g++ git
COPY . .
RUN go get ./... && \
  go build -a -installsuffix cgo -o app .

FROM alpine
RUN adduser -D -u 678 watchdog
USER watchdog
COPY --from=build /go/src/github.com/factoriotools/factorio-docker-watchdog/app /app
ENTRYPOINT ["/app"]
