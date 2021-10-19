# ============= Execution Stage ================
FROM alpine:3.12 AS execution

WORKDIR /run

# Copy the code into the container
COPY ./build/kurtosis-engine.bin .

# Execute Kurtosis Engine
CMD ./kurtosis-engine.bin
