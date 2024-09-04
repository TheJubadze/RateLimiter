# Use the official Golang image as the base
FROM golang:1.22 AS builder

# Continue with the rest of your build
WORKDIR /app
COPY . .

# Build your Go application
RUN go build -o rate-limiter ./cmd/server

FROM debian:stable-slim

# Set the working directory
WORKDIR /root

# Copy the pre-built binary from the previous stage
COPY --from=builder /app/rate-limiter .

# Install necessary packages for downloading and unpacking
RUN apt-get update && \
    apt-get install -y curl tar && \
    rm -rf /var/lib/apt/lists/*

# Set Goose version and architecture
ENV GOOSE_VERSION v3.22.0
ENV ARCH x86_64

# Download and install Goose precompiled binary
RUN curl -L https://github.com/pressly/goose/releases/download/$GOOSE_VERSION/goose_linux_$ARCH -o goose && \
    chmod +x goose && \
    mv goose /usr/local/bin/

# Install PostgreSQL client in the final stage
RUN apt-get update && apt-get install -y postgresql-client

# Copy configuration and migration files
COPY config.yaml /etc/rate-limiter/config.yaml
COPY migrations /migrations

# Entry point
CMD ["/rate-limiter"]
