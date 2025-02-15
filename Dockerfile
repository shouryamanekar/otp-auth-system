# Use the official Golang image as a build stage
FROM golang:1.23 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum first to leverage caching
COPY go.mod go.sum ./
RUN go mod tidy

# Copy the entire project
COPY . .

# Install swag CLI to generate Swagger docs inside the container
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Generate Swagger docs inside the container
RUN swag init

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o otp-auth-system .

# Use a minimal base image for production
FROM alpine:latest

# Install CA certificates for HTTPS connections
RUN apk --no-cache add ca-certificates

# Set the working directory inside the container
WORKDIR /root/

# Copy the built binary from the builder stage
COPY --from=builder /app/otp-auth-system .
COPY --from=builder /app/docs /docs

# Ensure the binary has execution permissions
RUN chmod +x otp-auth-system

# Expose the application port
EXPOSE 8080

# Run the application
CMD ["./otp-auth-system"]
