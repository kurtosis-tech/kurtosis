# ============= Build Stage ======================
FROM golang:1.15-alpine AS builder

WORKDIR /build

# Copy and download dependencies using go mod
# Because this is first, the dependencies will get cached as a Docker layer
COPY go.mod .
COPY go.sum .
RUN go mod download -x

# Copy the code into the container
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o kurtosis-initializer initializer/main.go

# ============= Execution Stage ================
FROM docker:stable AS execution

# !!IMPORTANT!! Whenever you update these, make sure to modify https://github.com/kurtosis-tech/kurtosis-docs !!
ENV CLIENT_ID=""
ENV CLIENT_SECRET=""
ENV CUSTOM_ENV_VARS_JSON="{}"
ENV DO_LIST="false"
ENV KURTOSIS_LOG_LEVEL="debug"
ENV PARALLELISM="4"
ENV TEST_NAMES=""
ENV TEST_SUITE_LOG_LEVEL="debug"

WORKDIR /run

# Copy the code into the container
COPY --from=builder /build/kurtosis-initializer .

# !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! IMPORTANT !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
#  1. Keep these flags sorted alphabetically!
#  2. If you make any changes here, you need to modify the docs at https://github.com/kurtosis-tech/kurtosis-docs
# !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! IMPORTANT !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
CMD ["sh", "-c", "set -euo pipefail && ./kurtosis-initializer \
    --client-id=${CLIENT_ID} \
    --client-secret=${CLIENT_SECRET} \
    --custom-env-vars-json=${CUSTOM_ENV_VARS_JSON} \
    --kurtosis-api-image=${KURTOSIS_API_IMAGE} \
    --kurtosis-log-level=${KURTOSIS_LOG_LEVEL} \
    --list=${DO_LIST} \
    --parallelism=${PARALLELISM} \
    --suite-execution-volume=${SUITE_EXECUTION_VOLUME} \
    --test-names=${TEST_NAMES} \
    --test-suite-image=${TEST_SUITE_IMAGE} \
    --test-suite-log-level=${TEST_SUITE_LOG_LEVEL}"]
# Code notes about the above:
# - This is the "exec" form of the command to ensure that SIGTERM gets passed to the API server
# - We call "sh" as the first command so that environment variable expansion will still happen
