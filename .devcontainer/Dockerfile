FROM mcr.microsoft.com/devcontainers/go:1-1.21-bullseye

ARG NODE_MAJOR=20

# Install Node.js, pnpm, Python toolchain, and protobuf compiler for schema work.
RUN curl -fsSL https://deb.nodesource.com/setup_${NODE_MAJOR}.x | bash - \
    && apt-get update \
    && apt-get install -y --no-install-recommends \
        nodejs \
        python3 \
        python3-pip \
        python3-venv \
        python3-dev \
        protobuf-compiler \
    && npm install -g pnpm@8 \
    && pip3 install --no-cache-dir poetry==1.7.1 \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

ENV PNPM_HOME=/usr/local/share/pnpm
ENV PATH="$PNPM_HOME:$PATH"

# Install buf to support protobuf linting and generation workflows.
RUN GOBIN=/usr/local/bin go install github.com/bufbuild/buf/cmd/buf@v1.30.0

