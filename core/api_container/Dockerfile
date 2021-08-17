# ============= Build Stage ======================
FROM golang:1.15-alpine AS builder

WORKDIR /build

# Copy and download dependencies using go mod
# Because this is first, the dependencies will get cached as a Docker layer
COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy the code into the container
COPY . .

# The API container gRPC server that will run in the container
RUN CGO_ENABLED=0 GOOS=linux go build -o kurtosis-api-server api_container/main.go

# The tiny CLI that we'll Docker exec to verify that the API container is up (we can't use a gRPC endpoint because
#  we don't have any guarantees of any other container being in the API container's enclave)
RUN CGO_ENABLED=0 GOOS=linux go build -o api-container-availability-waiter api_container_availability_waiter/main.go

# ============= Execution Stage ================
FROM alpine:3.12 AS execution

WORKDIR /run

COPY --from=builder /build/kurtosis-api-server .
COPY --from=builder /build/api-container-availability-waiter .

CMD ./kurtosis-api-server