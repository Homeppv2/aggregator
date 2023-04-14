# Build
FROM golang:1.20 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./
RUN go build -o /server main.go

# Deploy
FROM gcr.io/distroless/base-debian10
WORKDIR /app
COPY --from=build server ./
COPY config ./config
USER nonroot:nonroot
ENTRYPOINT [ "/app/server" ]