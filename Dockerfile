FROM golang:alpine

COPY . /go/src/go-crond

WORKDIR /go/src/go-crond

RUN apk --no-cache add git
RUN go get
RUN go build

CMD ["./go-crond"]
