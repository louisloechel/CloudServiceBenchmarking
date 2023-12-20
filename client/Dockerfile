# Use the official Go image as a parent image
FROM golang:1.18

# Set the working directory inside the container
WORKDIR /app

# Copy the go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the local package files to the container's workspace
COPY . .

# Build the application
RUN go build -o client

# Run the client when the container starts
CMD ["./client"]
