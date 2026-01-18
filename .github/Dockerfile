# CI/CD Docker image for terraform-provider-danubedata
FROM golang:1.24-bookworm

# Install dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    git \
    gnupg \
    curl \
    unzip \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Install golangci-lint (v1.64+ required for Go 1.24)
ARG GOLANGCI_LINT_VERSION=v1.64.5
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \
    sh -s -- -b /usr/local/bin ${GOLANGCI_LINT_VERSION}

# Install goreleaser
ARG GORELEASER_VERSION=v2.5.1
RUN curl -sSfL https://github.com/goreleaser/goreleaser/releases/download/${GORELEASER_VERSION}/goreleaser_Linux_x86_64.tar.gz | \
    tar xz -C /usr/local/bin goreleaser

# Install Terraform (for acceptance tests)
ARG TERRAFORM_VERSION=1.10.3
RUN curl -sSfL https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip -o terraform.zip && \
    unzip terraform.zip -d /usr/local/bin && \
    rm terraform.zip

# Set up Go environment
ENV GOPATH=/go
ENV PATH="${GOPATH}/bin:${PATH}"
# CGO enabled for race detector support
ENV CGO_ENABLED=1

# Create workspace directory
WORKDIR /workspace

# Default command
CMD ["bash"]
