FROM golang:alpine AS buildenv

COPY . /go/src/go-crond
WORKDIR /go/src/go-crond

RUN apk --no-cache add git \
    && go get \
    && go build \
    && chmod +x go-crond \
    && ./go-crond --version

FROM alpine
COPY --from=buildenv /go/src/go-crond/go-crond /usr/local/bin
CMD ["go-crond"]
