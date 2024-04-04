# Build
FROM golang:latest as build

WORKDIR /app

COPY . .

RUN go mod download


#COPY config ./config
#USER nonroot:nonroot
RUN go build -o ./server ./cmd/app/main.go

# Deploy
#FROM gcr.io/distroless/base-debian10
#WORKDIR /app
#COPY --from=build server ./
FROM ubuntu:22.04

RUN apt update
# RUN apt-get install -y curl
# RUN apt install libc6

WORKDIR /app

# COPY ./entrypoint.sh /usr/bin/entrypoint.sh

# ENTRYPOINT [ "entrypoint.sh" ]
CMD ["/app/server"]