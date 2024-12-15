# Use a minimal base image with Go installed
FROM golang:1.22-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Install build tools and dependencies
RUN apk add --no-cache git

# Copy the Go modules manifest and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the application source code
COPY . .

# Build the application binary
RUN go build -o eth-balance-exporter .

# Final stage: Use a smaller image for the final container
FROM alpine:latest

# Set working directory in the container
WORKDIR /root/

# Install CA certificates to allow secure connections
RUN apk add --no-cache ca-certificates

# Copy the application binary from the builder stage
COPY --from=builder /app/eth-balance-exporter .

# Expose the application port
EXPOSE 8080

# Command to run the application
CMD ["./eth-balance-exporter"]
