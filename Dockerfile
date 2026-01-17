# 🏗️ Stage 1: The Builder
# We use a large image with all the Go tools to compile the code
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy dependency files first (for better caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the code
COPY . .

# Build the binary
# -o main : output filename
# ./cmd/api/main.go : input file
RUN go build -o main ./cmd/api/main.go

# 🚀 Stage 2: The Runner
# We switch to a tiny empty image (Alpine) to keep the file size small
FROM alpine:latest

WORKDIR /root/

# Copy the binary from the Builder stage
COPY --from=builder /app/main .
# Copy the .env file (optional, but useful for now)
COPY --from=builder /app/.env .

# Expose the port
EXPOSE 8080

# Command to run when container starts
CMD ["./main"]