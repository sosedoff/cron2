# build stage
FROM golang:1.11 AS build
ADD . /go/src/github.com/sosedoff/cron2
WORKDIR /go/src/github.com/sosedoff/cron2
RUN go build -o cron2

# final stage
FROM alpine:3.6
RUN apk update && apk add --no-cache ca-certificates openssl 
WORKDIR /app
COPY --from=build /go/src/github.com/sosedoff/cron2/cron2 /bin/
ENTRYPOINT /bin/cron2