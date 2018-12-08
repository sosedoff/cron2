FROM golang:1.11 AS build
ADD . /go/src/github.com/sosedoff/cron2
WORKDIR /go/src/github.com/sosedoff/cron2
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./cron2-linux

FROM alpine:3.6
RUN apk update && apk add --no-cache ca-certificates openssl bash nano
COPY --from=build /go/src/github.com/sosedoff/cron2/cron2-linux /bin/cron2
CMD ["/bin/cron2"]