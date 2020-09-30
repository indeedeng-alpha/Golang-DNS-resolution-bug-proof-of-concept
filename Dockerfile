FROM golang:1.15.2

WORKDIR /src
COPY . .
RUN go build

