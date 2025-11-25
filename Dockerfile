# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o deployment-agent .

# Final stage - Use official Docker CLI image (includes compose plugin)
FROM docker:27-cli

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/deployment-agent .

# Expose port
EXPOSE 5000

# Run the application
CMD ["./deployment-agent"]
