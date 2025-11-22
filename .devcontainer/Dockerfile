FROM golang:1.25.4-trixie

# Install basic dependencies
RUN apt update && \
    apt install -y \
        curl \
        git \
        tar \
    && \
    rm -rf /var/lib/apt/lists/*

# Install task
RUN sh -c "$(curl -L https://taskfile.dev/install.sh)" -- -d v3.36.0

# Install golangci-lint
RUN sh -c "$(curl -L https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh)" -- v2.6.2

# Install gotestsum
RUN curl -L https://github.com/gotestyourself/gotestsum/releases/download/v1.13.0/gotestsum_1.13.0_linux_amd64.tar.gz -o gotestsum.tar.gz && \
    tar -xzf gotestsum.tar.gz -C /usr/local/bin gotestsum && \
    rm gotestsum.tar.gz

# Install jd
RUN curl -L https://github.com/josephburnett/jd/releases/download/v2.3.0/jd-amd64-linux -o /usr/local/bin/jd && \
    chmod +x /usr/local/bin/jd
