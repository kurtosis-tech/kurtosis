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
RUN CGO_ENABLED=0 GOOS=linux go build -o kurtosis-initializer initializer/main.go

# ============= Execution Stage ================
FROM docker:stable AS execution

# !!IMPORTANT!! Whenever you update these, make sure to modify https://github.com/kurtosis-tech/kurtosis-docs !!
ENV CLIENT_ID=""
ENV CLIENT_SECRET=""
ENV CUSTOM_ENV_VARS_JSON="{}"
ENV DO_LIST="false"
ENV KURTOSIS_LOG_LEVEL="info"
ENV PARALLELISM="4"
ENV SHOW_HELP="false"
ENV TEST_NAMES=""
ENV TEST_SUITE_LOG_LEVEL="debug"

WORKDIR /run

# Copy the code into the container
COPY --from=builder /build/kurtosis-initializer .

# !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! IMPORTANT !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
#  1. Keep these flags sorted alphabetically, to avoid duplicates (which will cause subtle problems)
#  2. Keep the flag name the exact same as the Docker environment variable, so main.go's usage function
#       continues to make sense for users
#  3. If you change any of these:
#       a. Update the flag-parsing in main.go
#       b. Update the build_and_run.sh script, which passes in these environment variables
# !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! IMPORTANT !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
# vvvvvvvvvvvvv Read the IMPORTANT section before modifying this! vvvvvvvvvvvvvvvvvvvvvvvv
CMD ["sh", "-c", "set -euo pipefail && ./kurtosis-initializer \
    --CLIENT_ID=${CLIENT_ID} \
    --CLIENT_SECRET=${CLIENT_SECRET} \
    --CUSTOM_ENV_VARS_JSON=${CUSTOM_ENV_VARS_JSON} \
    --DO_LIST=${DO_LIST} \
    --KURTOSIS_API_IMAGE=${KURTOSIS_API_IMAGE} \
    --KURTOSIS_LOG_LEVEL=${KURTOSIS_LOG_LEVEL} \
    --PARALLELISM=${PARALLELISM} \
    --SHOW_HELP=${SHOW_HELP} \
    --SUITE_EXECUTION_VOLUME=${SUITE_EXECUTION_VOLUME} \
    --TEST_NAMES=${TEST_NAMES} \
    --TEST_SUITE_IMAGE=${TEST_SUITE_IMAGE} \
    --TEST_SUITE_LOG_LEVEL=${TEST_SUITE_LOG_LEVEL}"]
# ^^^^^^^^^^^^^ Read the IMPORTANT section before modifying this! ^^^^^^^^^^^^^^^^^^^^^^^^

# Code notes about the above:
# - This is the "exec" form of the command to ensure that SIGTERM gets passed to the API server
# - We call "sh" as the first command so that environment variable expansion will still happen
