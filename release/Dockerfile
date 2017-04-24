FROM alpine

RUN GOCROND_RELEASE=0.1.0 \
    && wget -O /usr/local/bin/go-crond https://github.com/webdevops/go-crond/releases/download/$GOCROND_RELEASE/go-crond-64-linux \
    && chmod +x /usr/local/bin/go-crond

CMD ["go-crond"]
