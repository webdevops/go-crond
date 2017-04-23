FROM golang:alpine

COPY . /go/src/go-crond

WORKDIR /go/src/go-crond

RUN apk --no-cache add --virtual .gocrond-deps git \
    && go get \
    && go build \
    && mv go-crond /usr/local/bin \
    && rm -rf /go/src/ \
    && apk del .gocrond-deps

CMD ["go-crond"]
