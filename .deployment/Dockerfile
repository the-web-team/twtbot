FROM golang:latest as builder

WORKDIR /go/src/twtbot
COPY . /go/src/twtbot

RUN go mod vendor

ENV GOOS=linux
ENV GOARCH=amd64
RUN go build -o twtbot .

FROM alpine:latest

WORKDIR /var/app

RUN apk add --update ca-certificates
RUN apk add --no-cache libc6-compat

COPY --from=builder /go/src/twtbot/twtbot /var/app/twtbot

CMD ./twtbot
