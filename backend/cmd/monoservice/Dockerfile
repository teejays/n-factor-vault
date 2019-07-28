FROM golang:alpine

ENV GO111MODULE=on

RUN apk update && apk add make git

RUN mkdir -p /app

WORKDIR /app
RUN go mod download
COPY go.mod .
COPY go.sum .
COPY . /app

WORKDIR /app/cmd/monoservice

RUN make dev-bin-build

CMD make dev-bin-run