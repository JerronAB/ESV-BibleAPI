# Use a multi-stage build for Go
FROM golang:1.20 AS go_exec_builder

# Set the working directory
WORKDIR /app

# Copy the Go script into the container
COPY BibleServe.go .

# Build the Go script
RUN go build -o BibleServe BibleServe.go

# Use a minimal base image
FROM alpine:latest

# Install necessary dependencies
RUN apk add --no-cache libc6-compat

# Copy the built Go executable from the builder stage
COPY --from=go_exec_builder /app/BibleServe /app/BibleServe

# Copy the text file
COPY ESVBible.txt /app/ESVBible.txt

# Set the working directory
WORKDIR /app

# Expose the port your application uses
EXPOSE 80

# Run the Go server
CMD ["./BibleServe"]
