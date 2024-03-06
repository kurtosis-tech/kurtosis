FROM alpine:latest

WORKDIR /app

COPY . .

ARG BUILD_ARG="DEFAULT_VALUE"
RUN if [ "$BUILD_ARG" != "VALUE" ]; then \
  echo "Error: the build argument BUILD_ARG has not been provided"; \
  exit 1; \
fi

CMD ["echo", "Hello, Kurtosis!"]
