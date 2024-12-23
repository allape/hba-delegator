FROM golang:1.23.4-alpine3.20 AS builder

RUN apk add --no-cache build-base

WORKDIR /build

COPY go.mod go.mod
COPY go.sum go.sum
RUN /usr/local/go/bin/go mod download

COPY . .
RUN /usr/local/go/bin/go build -o app

FROM alpine:3.20

WORKDIR /data

WORKDIR /

COPY --from=builder /build/app /bin/hbad

CMD [ "/bin/hbad" ]

### build ###
# export docker_http_proxy=http://host.docker.internal:1080
# docker build --build-arg http_proxy=$docker_http_proxy --build-arg https_proxy=$docker_http_proxy -f Dockerfile -t allape/hbad:latest .
# docker build --platform linux/amd64 --build-arg http_proxy=$docker_http_proxy --build-arg https_proxy=$docker_http_proxy -f Dockerfile -t allape/hbad:latest .
# docker push allape/hbad:latest

###  run  ###
# docker compose -f docker.compose.yaml up -d

