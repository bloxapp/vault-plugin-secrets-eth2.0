#
# STEP 1: Prepare environment
#
FROM golang:1.14-stretch AS preparer

RUN apt-get update                                                        && \
  DEBIAN_FRONTEND=noninteractive apt-get install -yq --no-install-recommends \
    curl git zip unzip wget g++ python gcc jq                                \
  && rm -rf /var/lib/apt/lists/*

RUN go version
RUN python --version

WORKDIR /go/src/github.com/bloxapp/key-vault/
COPY go.mod .
COPY go.sum .
RUN go mod download

#
# STEP 2: Build executable binary
#
FROM preparer AS builder

# Copy files and install app
COPY . .
RUN go get -d -v ./...
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags "-linkmode external -extldflags \"-static -lm\" -X main.Version=$(git describe --tags $(git rev-list --tags --max-count=1))" -o ethsign .

#
# STEP 3: Get vault image and copy the plugin
#
FROM vault:1.6.0 AS runner

# Download dependencies
RUN apk -v --update --no-cache add \
    bash ca-certificates curl openssl

WORKDIR /vault/plugins/

COPY --from=builder /go/src/github.com/bloxapp/key-vault/ethsign ./ethsign
COPY ./config/vault-config.json /vault/config/vault-config.json
COPY ./config/vault-config-tls.json /vault/config/vault-config-tls.json
COPY ./config/entrypoint.sh /vault/config/entrypoint.sh
COPY ./config/vault-tls.sh /vault/config/vault-tls.sh
COPY ./config/vault-init.sh /vault/config/vault-init.sh
COPY ./config/vault-unseal.sh /vault/config/vault-unseal.sh
COPY ./config/vault-plugin.sh /vault/config/vault-plugin.sh

RUN chown vault /vault/config/entrypoint.sh
RUN apk add jq

WORKDIR /

# Expose port 8200
EXPOSE 8200

# Run vault
CMD ["bash", "/vault/config/entrypoint.sh"]
