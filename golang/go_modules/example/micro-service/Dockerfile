FROM golang:1.12-alpine3.9

COPY . $GOPATH/src/github.com/aaronzhuo1990/blogs/golang/go_modules/example/micro-service
RUN go build -o /usr/bin/micro-service github.com/aaronzhuo1990/blogs/golang/go_modules/example/micro-service/server && rm -rf $GOPATH/*

ENTRYPOINT ["/usr/bin/micro-service"]
