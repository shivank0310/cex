FROM golang:1.21-alpine AS builder
WORKDIR /src
COPY docker/stubs/health/main.go .
RUN CGO_ENABLED=0 go build -o /bin/health main.go

FROM alpine:3.19
RUN apk add --no-cache ca-certificates curl
COPY --from=builder /bin/health /bin/health
ENTRYPOINT ["/bin/health"]
