# Docker base image allows us to run docker-in-docker
FROM docker:stable

# Set necessary environmet variables needed for our image
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    # Defaults for images point to DockerHub endpoints.
    # These should be set in docker container run args
    DEFAULT_GECKO_IMAGE="avaplatform/gecko" \
    TEST_CONTROLLER_IMAGE="kurtosistech/ava-test-controller:latest"


# Move to working directory /build
WORKDIR /build

# Install Golang
RUN apk update && \
        apk upgrade && \
        apk add go

# Copy and download dependencies using go mod
COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy the code into the container
COPY . .

# Build the application
RUN go build -o kurtosis .

# Move to /dist directory as the place for resulting binary folder
WORKDIR /dist

# Copy binary from build to main folder
RUN cp /build/kurtosis .

# Command to run when starting the container
CMD /dist/kurtosis --gecko-image-name=${DEFAULT_GECKO_IMAGE} --test-controller-image-name=${TEST_CONTROLLER_IMAGE}
