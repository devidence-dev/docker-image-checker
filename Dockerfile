# Build stage
FROM golang:1.24.4-alpine3.22 AS builder

WORKDIR /app

# Install dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o docker-image-checker ./cmd/checker

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/docker-image-checker .

# Copy configuration files
COPY --from=builder /app/configs ./configs
COPY --from=builder /app/templates ./templates

# Create logs directory
RUN mkdir -p logs

# Run the application
CMD ["./docker-image-checker", "--daemon"]
