# Use the official Go image as a parent image
FROM golang:1.19

# Set the working directory inside the container
WORKDIR /app

# Copy the go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the protobuf generated files and client code to the container's workspace
COPY pb/ pb/
COPY client/ .

# Build the application
RUN go build -o client ./main.go

# No need to expose a port for the client

# Run the client/main when the container starts
CMD ["./client"]
