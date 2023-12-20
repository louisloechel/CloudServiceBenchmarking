# Use the official Go image as a parent image
FROM golang:1.18

# Set the working directory inside the container
WORKDIR /server

# Copy the go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the local package files to the container's workspace
COPY server/ .

# Build the application
RUN go build -o server ./server/main.go

# Expose port 8080 to the outside world
EXPOSE 8080

# Run the server when the container starts
CMD ["./server"]
