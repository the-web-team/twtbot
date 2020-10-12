FROM golang:latest as builder

WORKDIR /go/src/twtbot

COPY ./ /go/src/twtbot/

RUN go mod vendor
RUN go build -o twtbot

FROM alpine:latest

WORKDIR /var/app

COPY --from=builder /go/src/twtbot/twtbot /var/app/twtbot

CMD sh /var/app/twtbot
