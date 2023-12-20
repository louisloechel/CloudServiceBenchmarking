# Use the official Go image as a parent image
FROM golang:1.18
# Use the official Go image as a parent image
FROM golang:1.19

# Set the working directory inside the container
WORKDIR /app

# Copy the go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the protobuf generated files and server code to the container's workspace
COPY pb/ pb/
COPY server/ .

# Build the application
RUN go build -o main ./main.go

# Expose the gRPC port
EXPOSE 50051

# Run the client/main when the container starts
CMD ["./main"]
