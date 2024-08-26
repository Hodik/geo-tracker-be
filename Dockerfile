# Step 1: Build the Go app
FROM golang:1.22-alpine AS builder

WORKDIR /app



COPY go.mod go.sum ./

RUN go mod download

COPY . .

ENV GOARCH=amd64
ENV GOOS=linux

RUN go build -o main .

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/main .

COPY .env .