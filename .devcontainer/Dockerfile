FROM golang:1.25.4-trixie

# Install basic dependencies
RUN apt update && \
    apt install -y \
        git \
        curl \
    && \
    rm -rf /var/lib/apt/lists/*

# Install task
RUN sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d v3.36.0

# Install golangci-lint
RUN sh -c "$(curl --location https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh)" -- v2.6.2
