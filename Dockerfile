# syntax=docker/dockerfile:1

FROM golang:1.20-alpine as build
RUN apk add --no-cache gcc musl-dev
RUN apk --no-cache add ca-certificates curl bash xz-libs git

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
COPY config.json ./

RUN go mod download

COPY *.go ./

RUN go build -o main /app

FROM alpine:3.11.3
COPY --from=build /app ./


CMD ["./main"]
