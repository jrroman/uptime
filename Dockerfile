FROM golang:1.11.1-alpine3.8

LABEL Author="John Roman <jrrmod@gmail.com>"

WORKDIR /go/src/hlab

RUN apk add \
    make \
    git \
    curl \
    --no-cache \
    # install dep dependency tool
    && curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

COPY . /go/src/hlab

EXPOSE 8080

CMD ["./bin/run.sh"]
