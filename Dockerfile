FROM golang:alpine

RUN apk add --no-cache curl gcc musl-dev gzip openjdk17-jre bash protoc protobuf-dev make file

COPY . /go/src/github.com/pact-foundation/pact-go

WORKDIR /go/src/github.com/pact-foundation/pact-go

CMD ["make", "test"]