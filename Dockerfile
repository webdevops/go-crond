FROM golang:1.14 as build

WORKDIR /go/src/github.com/webdevops/go-crond

# Get deps (cached)
COPY ./go.mod /go/src/github.com/webdevops/go-crond
COPY ./go.sum /go/src/github.com/webdevops/go-crond
RUN go mod download

# Compile
COPY ./ /go/src/github.com/webdevops/go-crond
RUN make lint
RUN make build-local

#############################################
# FINAL IMAGE
#############################################
FROM alpine
COPY --from=build /go/src/github.com/webdevops/go-crond/go-crond /usr/local/bin
CMD ["go-crond"]
