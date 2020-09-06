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
RUN CGO_ENABLED=0 GOOS=linux go build -o kurtosis-initializer initializer/main.go

# ============= Execution Stage ================
FROM docker:stable AS execution

ENV DO_LIST="false"
ENV TEST_NAMES=""
ENV KURTOSIS_LOG_LEVEL="debug"
ENV TEST_SUITE_LOG_LEVEL="debug"
ENV KURTOSIS_API_IMAGE="kurtosistech/kurtosis-core_api:latest"
ENV PARALLELISM="4"
ENV CUSTOM_ENV_VARS_JSON="{}"

WORKDIR /run

# Copy the code into the container
COPY --from=builder /build/kurtosis-initializer .

# IMPORTANT NOTES:
#  1) The flag arguments are sorted alphabetically!!
#  2) This is the "exec" form of the command to ensure that SIGTERM gets passed to the API server
#  3) We call "sh" as the first command so that environment variable expansion will still happen
CMD ["sh", "-c", "set -euo pipefail && ./kurtosis-initializer \
    --custom-env-vars-json=${CUSTOM_ENV_VARS_JSON} \
    --kurtosis-api-image=${KURTOSIS_API_IMAGE} \
    --kurtosis-log-level=${KURTOSIS_LOG_LEVEL} \
    --list=${DO_LIST} \
    --parallelism=${PARALLELISM} \
    --suite-execution-volume=${SUITE_EXECUTION_VOLUME} \
    --test-names=${TEST_NAMES} \
    --test-suite-image=${TEST_SUITE_IMAGE} \
    --test-suite-log-level=${TEST_SUITE_LOG_LEVEL}"]
