# ============= Build Stage ======================
FROM golang:1.15-alpine AS builder
WORKDIR /build
# Copy and download dependencies using go mod
COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy the code into the container
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o kurtosis-api-server api_container/main.go

# ============= Execution Stage ================
FROM docker:stable AS execution
WORKDIR /run

# Copy the code into the container
COPY --from=builder /build/kurtosis-api-server .

# NOTE:
#  1) this is the "exec" form of the command to ensure that SIGTERM gets passed to the API server
#  2) we call "sh" as the first command so that variable expansion will still happen
CMD ["sh", "-c", "set -euo pipefail && ./kurtosis-api-server \
        --test-suite-container-id=${TEST_SUITE_CONTAINER_ID} \
        --network-id=${NETWORK_ID} \
        --subnet-mask=${SUBNET_MASK} \
        --gateway-ip=${GATEWAY_IP} \
        --log-level=${LOG_LEVEL}"]
