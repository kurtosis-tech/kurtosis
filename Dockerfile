# FROM debian:bookworm-slim as builder
# FROM docker.io/goreleaser/goreleaser:v1.24.0
FROM docker.io/golang:bookworm

RUN echo 'deb [trusted=yes] https://repo.goreleaser.com/apt/ /' | tee /etc/apt/sources.list.d/goreleaser.list
RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash -

RUN apt update && apt install -qy build-essential curl jq yq goreleaser nodejs git

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2

RUN go install github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen@latest

RUN curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- --default-toolchain stable -y
ENV PATH="$PATH:/root/.cargo/bin"


RUN npm i -g yarn

RUN npm i -g openapi-typescript

# COPY ./api /src/api
# COPY ./kurtosis_version /src/kurtosis_version
# COPY ./name_generator /src/name_generator
# COPY ./container-engine-lib /src/container-engine-lib
# COPY ./grpc-file-transfer /src/grpc-file-transfer
# COPY ./contexts-config-store /src/contexts-config-store
# COPY ./cli /src/cli
# COPY ./scripts /src/scripts
# COPY ./utils /src/utils
# COPY ./.git /src/.git
# COPY ./cloud /src/cloud

COPY . /src


WORKDIR /src

RUN ./scripts/generate-kurtosis-version.sh
RUN ./container-engine-lib/scripts/build.sh
RUN ./contexts-config-store/scripts/build.sh
RUN ./grpc-file-transfer/scripts/build.sh
RUN ./name_generator/scripts/build.sh
RUN ./api/scripts/build.sh
RUN ./cli/scripts/build.sh

# COPY ./cli/
# RUN ./cli/scripts/build.sh
# COPY ./cli /src/cli

# WORKDIR /src/cli

# RUN ./

# RUN go build -o /usr/local/bin/kurtosis ./cli/main.go

RUN cp ./cli/cli/dist/cli_linux_amd64_v1/kurtosis /usr/local/bin/kurtosis

# FROM debian:bookworm-slim as runner


# # COPY --from=builder 

ENTRYPOINT /usr/local/bin/kurtosis