# Build
FROM golang:latest as build

WORKDIR /app

COPY * .

RUN go mod download


#COPY config ./config
#USER nonroot:nonroot
WORKDIR /app/cmd/app
RUN go build -o ./server main.go

# Deploy
#FROM gcr.io/distroless/base-debian10
#WORKDIR /app
#COPY --from=build server ./
FROM ubuntu:20.04

RUN apt-get update
RUN apt-get install -y curl

COPY --from=build /app/server .
# COPY ./entrypoint.sh /usr/bin/entrypoint.sh

# ENTRYPOINT [ "entrypoint.sh" ]
CMD ["/server"]