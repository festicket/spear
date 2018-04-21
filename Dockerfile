FROM alpine:3.5
# FROM phusion/baseimage:0.9.22

RUN apk add --no-cache ca-certificates

ADD ./bin/spear /main/spear
WORKDIR /main

CMD ["./spear"]
