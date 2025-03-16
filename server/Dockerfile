# Use official image as a parent image
FROM golang:1.16.3-alpine3.13

# Set the working directory
WORKDIR /app

# Copy the current files to working directory
COPY . .

# Download and Install the dependencies
RUN go get -d -v ./...

# Build 
RUN go build -o main .

# Expose the Port
EXPOSE 8000

# Run the Executable
CMD ["./mserver"]
