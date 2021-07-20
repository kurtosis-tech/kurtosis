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

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o kurtosis-api-server api_container/main.go

# ============= Execution Stage ================
FROM alpine:3.12 AS execution

WORKDIR /run

# Copy the code into the container
COPY --from=builder /build/kurtosis-api-server .

CMD ./kurtosis-api-server