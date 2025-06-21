# Build stage
FROM golang:1.24.4-bullseye AS builder

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Set working directory
WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Copy templates directory
COPY --from=builder /app/templates ./templates

# Create static directory (in case it's needed)
RUN mkdir -p static

# Expose port
EXPOSE 8080

# Run the application
CMD ["./main"] 