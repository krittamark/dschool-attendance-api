# Build Stage
FROM golang:alpine AS builder

ENV GOTOOLCHAIN=auto
WORKDIR /app

# Install git, ca-certificates, tzdata
RUN apk add --no-cache git ca-certificates tzdata

# Copy dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build statically linked binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/bin/server cmd/server/main.go

# Production Runtime Stage
FROM alpine:3.20

WORKDIR /app

# Install ca-certificates and tzdata for Bangkok timezone
RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Bangkok

# Create data directory for api_keys.json
RUN mkdir -p /app/data

# Copy binary from builder
COPY --from=builder /app/bin/server /app/server

# Copy OpenAPI spec
COPY --from=builder /app/openapi.json /app/openapi.json
COPY --from=builder /app/openapi.yaml /app/openapi.yaml

# Expose standard Cloud Run port
ENV PORT=8080
EXPOSE 8080

# Run binary
CMD ["/app/server"]
