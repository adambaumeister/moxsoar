## First stage: build the moxsoar binary ##
FROM golang:1.12.6 AS builder
ADD . /go/src/github.com/abaumeister/moxsoar/ 
WORKDIR /go/src/github.com/abaumeister/moxsoar/
# Get the dependenices
#RUN go get -v \
#  && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o moxsoar .
RUN go get -v
RUN go build -o moxsoar .

## Deploy everything into the final container ##
FROM ubuntu:latest 
RUN mkdir /etc/moxsoar
WORKDIR /etc/moxsoar
RUN apt update -y \
 && apt install -y git vim 
COPY --from=builder /go/src/github.com/abaumeister/moxsoar/ .

