FROM golang:1.20

WORKDIR /usr/src/app

COPY . .

RUN go mod tidy
