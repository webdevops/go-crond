FROM golang:1.4.2
COPY . /go/src/go-cron
WORKDIR /go/src/go-cron

ENV GOPATH /go/src/go-cron/Godeps/_workspace:$GOPATH
RUN go install -v 

ENTRYPOINT ["go-cron"]
CMD ["-h"]
