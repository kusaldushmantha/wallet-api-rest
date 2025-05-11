# Use Go 1.24 base image
FROM golang:1.24

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum to leverage layer caching
COPY go.mod go.sum ./

# Download Go module dependencies
RUN go mod tidy

# Copy the rest of the application code
COPY . .

# Build the Go app from the main.go file
RUN make build

# Expose the port your app listens on (adjust if needed)
EXPOSE 8080

# Run the compiled Go binary
CMD ["./wallet-app"]
