FROM golang:alpine

COPY . /go/src/go-crond

WORKDIR /go/src/go-crond

RUN apk --no-cache add --virtual .gocrond-deps git \
    && go get \
    && go build \
    && rm -rf /go/src/github.com/ \
    && apk del .gocrond-deps

CMD ["./go-crond"]
