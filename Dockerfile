## Package moxsoar-ui ##
FROM node:latest as ui-builder
WORKDIR /tmp
RUN git clone https://github.com/adambaumeister/moxsoar-ui.git \
    && cd moxsoar-ui
WORKDIR /tmp/moxsoar-ui
RUN npm install \
    && npm run build

## First stage: build the moxsoar binary ##
FROM golang:1.14 AS builder
ADD . /go/src/github.com/abaumeister/moxsoar/ 
WORKDIR /go/src/github.com/abaumeister/moxsoar/
# Get the dependenices
#RUN go get -v \
#  && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o moxsoar .
RUN go build -o moxsoar .

## Deploy everything into the final container ##
FROM ubuntu:latest
# Add all the required directories  
RUN mkdir /etc/moxsoar \
    && mkdir /etc/moxsoar/data \
    && mkdir /etc/moxsoar/content \
    && mkdir /certs \
    && mkdir /etc/moxsoar/static
WORKDIR /etc/moxsoar
RUN apt update -y \
 && apt install -y git vim 
# Copy the Moxsoar binary and base config file
COPY --from=builder /go/src/github.com/abaumeister/moxsoar/moxsoar .
COPY --from=builder /go/src/github.com/abaumeister/moxsoar/moxsoar.yml .
# Copy the UI bundle
COPY --from=ui-builder /tmp/moxsoar-ui/build/ /etc/moxsoar/static 

# Start it up
EXPOSE 8000-8999
VOLUME /etc/moxsoar/data
VOLUME /etc/moxsoar/content
VOLUME /certs
CMD ["./moxsoar", "run", "--config", "./moxsoar.yml"]
