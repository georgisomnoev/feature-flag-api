FROM golang:1.24 AS builder

ENV CGO_ENABLED=0 \
    GOOS=linux

WORKDIR /build

COPY go.mod go.sum ./

COPY vendor ./vendor
ENV GOFLAGS=-mod=vendor

COPY internal ./internal
COPY cmd ./cmd


RUN go build -o feature_flags_api ./cmd/feature_flags_api

FROM alpine:latest

WORKDIR /app
COPY --from=builder /build/feature_flags_api .
COPY certs ./certs

CMD ["./feature_flags_api"]