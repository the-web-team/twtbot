FROM golang:alpine

RUN apk add inotify-tools

WORKDIR /go/src/twtbot

CMD inotifywait -r /go/src/twtbot | go run .