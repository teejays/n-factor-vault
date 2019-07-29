FROM golang:alpine

RUN apk update && apk add make git

ENV GO111MODULE=on 


# Pass the env variable: DEV, STAGE, PROD, TEST
ARG app_env
ENV ENV $app_env 


RUN mkdir -p /app
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . /app
WORKDIR /app/cmd/monoservice
RUN go mod download

# BUILD
RUN go build -o monoservice.bin main.go 

# Before we start the monoservice, we should wait for DB to be up and running. This script helps with that, if we define WAIT_HOSTS in the environments
# Source: https://dev.to/hugodias/wait-for-mongodb-to-start-on-docker-3h8b
ADD https://github.com/ufoscout/docker-compose-wait/releases/download/2.2.1/wait /wait
RUN chmod +x /wait

# RUN
# Declare ARGS
ARG pg_host
ARG pg_port
ARG pg_user
ARG pg_pwd
ARG pg_dbname
ENV POSTGRES_HOST=$pg_host POSTGRES_PORT=$pg_port POSTGRES_DBNAME=$pg_dbname POSTGRES_USER=$pg_user POSTGRES_PWD=$pg_pwd

CMD /wait && ./monoservice.bin