# ============= Build Stage ======================
FROM golang:1.15-alpine AS builder

WORKDIR /build

# Copy and download dependencies using go mod
# Because this is first, the dependencies will get cached as a Docker layer
COPY go.mod .
COPY go.sum .
RUN go mod download

RUN apt install -y protobuf-compiler=3.13.0-r2

# Copy the code into the container
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o kurtosis-api-server api_container/main.go

# ============= Execution Stage ================
FROM alpine:3.12 AS execution

WORKDIR /run

# Copy the code into the container
COPY --from=builder /build/kurtosis-api-server .

# !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! IMPORTANT !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
#              1. Need to change api_container_env_vars if any of these changes!!
#              2. Keep these sorted alphabetically, to ensure no duplicates
# !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! IMPORTANT !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
CMD ["sh", "-c", "set -euo pipefail && ./kurtosis-api-server \
    --log-level=\"${LOG_LEVEL}\" \
    --mode=\"${MODE}\" \
    --params-json=\"${PARAMS_JSON}\" \
"]
# Developer notes about the above:
# - This is the "exec" form of the command to ensure that SIGTERM gets passed to the API server
# - We call "sh" as the first command so that environment variable expansion will still happen
