# Multi-stage Dockerfile for CEX Go microservices.
# Build: docker build -f docker/Dockerfile --build-arg SERVICE_PATH=order-service/cmd/order-service -t cex/order-service .
FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG SERVICE_PATH
RUN test -n "$SERVICE_PATH"
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/service ./${SERVICE_PATH}

FROM alpine:3.19

RUN apk add --no-cache ca-certificates curl

COPY --from=builder /bin/service /bin/service

ENTRYPOINT ["/bin/service"]
