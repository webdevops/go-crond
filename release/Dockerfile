FROM alpine

RUN apk --no-cache add --virtual .gocrond-deps \
        ca-certificates  \
        wget \
    && GOCROND_RELEASE=0.2.0 \
    && wget -O /usr/local/bin/go-crond https://github.com/webdevops/go-crond/releases/download/$GOCROND_RELEASE/go-crond-64-linux \
    && chmod +x /usr/local/bin/go-crond \
    && apk del .gocrond-deps

CMD ["go-crond"]
