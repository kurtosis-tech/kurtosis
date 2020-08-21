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

# Note that this CANNOT be an execution list else the variables won't be expanded
# See: https://stackoverflow.com/questions/40454470/how-can-i-use-a-variable-inside-a-dockerfile-cmd
CMD set -euo pipefail && ./kurtosis-api-server \
    --network-id=${NETWORK_ID} \
    --subnet-mask=${SUBNET_MASK} \
    --gateway-ip=${GATEWAY_IP} \
    --container-ip=${CONTAINER_IP} \
    --log-level=${LOG_LEVEL} 2>&1 | tee ${LOG_FILEPATH}
