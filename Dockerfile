FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY . .

# Use standard Go build process for the builder stage
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o goose-connect

FROM debian:bookworm-slim

# Install system dependencies
RUN apt-get update && apt-get install -y \
    curl \
    git \
    procps \
    make \
    ca-certificates \
    bzip2 \
    libxcb1 \
    libdbus-1-3 \
    && rm -rf /var/lib/apt/lists/*


# Install mise for runtime management
RUN curl -fsSL https://mise.jdx.dev/install.sh | sh

# Set working directory
WORKDIR /root

# Copy kommon binary from builder
COPY --from=builder /app/goose-connect /usr/local/bin/

# Configure PATH for mise
ENV PATH="/root/.local/bin:/root/.mise/shims:${PATH}"

# Install common developer tools with mise
# These can be adjusted based on project needs
RUN /root/.local/bin/mise use --global golang@latest && \
    /root/.local/bin/mise use --global nodejs@lts && \
    /root/.local/bin/mise install

# Download and install Goose binary
RUN curl -fsSL https://github.com/block/goose/releases/download/stable/download_cli.sh -o install.sh \
    && chmod +x install.sh \
    && ./install.sh \
    && rm install.sh

# Create necessary directories for Goose
RUN mkdir -p /root/.config/goose /root/.local/share/goose

COPY assets/.goosehints /root/.config/goose/

# Verify installation
RUN goose --version

ENTRYPOINT ["goose-connect"]
