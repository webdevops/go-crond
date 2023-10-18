#############################################
# Build
#############################################
FROM --platform=$BUILDPLATFORM golang:1.20-alpine as build

RUN apk upgrade --no-cache --force
RUN apk add --update build-base make git

WORKDIR /go/src/github.com/webdevops/go-crond

# Dependencies
COPY go.mod go.sum .
RUN go mod download

# Compile
COPY . .
RUN make test
ARG TARGETOS TARGETARCH
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} make build

#############################################
# Test
#############################################
FROM gcr.io/distroless/static as test
USER 0:0
WORKDIR /app
COPY --from=build /go/src/github.com/webdevops/go-crond/go-crond .
RUN ["./go-crond", "--help"]

#############################################
# Final alpine
#############################################
FROM alpine as final-alpine
ENV SERVER_BIND=":8080" \
    SERVER_METRICS="1" \
    LOG_JSON="1"
WORKDIR /
COPY --from=test /app /usr/local/bin
EXPOSE 8080
ENTRYPOINT ["go-crond"]

#############################################
# FINAL debian
#############################################
FROM debian:stable-slim as final-debian
ENV SERVER_BIND=":8080" \
    SERVER_METRICS="1" \
    LOG_JSON="1"
COPY --from=test /app /usr/local/bin
EXPOSE 8080
ENTRYPOINT ["go-crond"]

#############################################
# FINAL ubuntu
#############################################
FROM ubuntu:latest as final-ubuntu
ENV SERVER_BIND=":8080" \
    SERVER_METRICS="1" \
    LOG_JSON="1"
COPY --from=test /app /usr/local/bin
EXPOSE 8080
ENTRYPOINT ["go-crond"]
