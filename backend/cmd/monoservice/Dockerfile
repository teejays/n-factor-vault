FROM golang:1.13.1-alpine

RUN apk update && apk add make git

ENV GO111MODULE=on 
ENV CGO_ENABLED=0

# Before we start the monoservice, we should wait for DB to be up and running. This script helps with that, if we define WAIT_HOSTS in the environments
# Source: https://dev.to/hugodias/wait-for-mongodb-to-start-on-docker-3h8b
ADD https://github.com/ufoscout/docker-compose-wait/releases/download/2.2.1/wait /wait
RUN chmod +x /wait
RUN /wait

# We don't need these since we're now mapping the code as a volume on the container through docker-compose  
# RUN mkdir -p /app
WORKDIR /app
# COPY go.mod .
# COPY go.sum .
# RUN go mod download
# COPY . /app