FROM golang:1.12-alpine3.9

RUN apk add --update \
    ca-certificates \
    git

ENV GO111MODULE=on

COPY . $GOPATH/src/github.com/azhuox/blogs/golang/go_modules/example/micro-service
RUN go build -o /usr/bin/micro-service github.com/azhuox/blogs/golang/go_modules/example/micro-service/server && rm -rf $GOPATH/*

ENTRYPOINT ["/usr/bin/micro-service"]
