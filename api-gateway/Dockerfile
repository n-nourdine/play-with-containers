# Start from Alpine for a minimal base image
FROM alpine:3.21

# Install build dependencies
RUN apk add --no-cache wget curl tar bash ca-certificates

# Set Go version
ENV GO_VERSION=1.21.13

# Download and install Go manually
RUN wget https://go.dev/dl/go1.21.13.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go1.21.13.linux-amd64.tar.gz && \
    rm go1.21.13.linux-amd64.tar.gz

# Set Go environment variables
ENV PATH=$PATH:/usr/local/go/bin
ENV GOPATH=/go
ENV PATH=$GOPATH/bin:$PATH

# Create app directories
RUN mkdir -p /app $GOPATH/src $GOPATH/bin

# Set working directory
WORKDIR /app

# Copy dependency files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application
COPY . .

# Build the application
RUN go build -o api-gateway .

# Expose the application port
EXPOSE 3000

# Run the application
CMD ["./api-gateway"]