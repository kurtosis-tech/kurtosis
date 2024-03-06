FROM alpine:latest

WORKDIR /app

COPY . .


# The idea is to test if the build argument has been provided.
# The following command would not work:
# $ docker build -t image-build-package-test .
# Error: the build argument BUILD_ARG has not been provided
# While this command would work:
# $ docker build -t image-build-package-test --build-arg "BUILD_ARG=VALUE" .
ARG BUILD_ARG="DEFAULT_VALUE"
RUN if [ "$BUILD_ARG" != "VALUE" ]; then \
  echo "Error: the build argument BUILD_ARG has not been provided"; \
  exit 1; \
fi

CMD ["echo", "Hello, Kurtosis!"]
