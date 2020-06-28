FROM golang:1.14 as build

WORKDIR /go/src/github.com/webdevops/go-crond

# Get deps (cached)
COPY ./go.mod /go/src/github.com/webdevops/go-crond
COPY ./go.sum /go/src/github.com/webdevops/go-crond
RUN go mod download

# Compile
COPY ./ /go/src/github.com/webdevops/go-crond
RUN make lint
RUN make build
RUN ./go-crond --help

#############################################
# FINAL IMAGE
#############################################
FROM alpine
COPY --from=buildenv /go/src/go-crond/go-crond /usr/local/bin
CMD ["go-crond"]
