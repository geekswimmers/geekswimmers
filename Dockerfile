FROM golang:1.23

WORKDIR /usr/src/app

COPY . .

RUN go mod tidy
