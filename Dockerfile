FROM golang:latest

ARG TARGETARCH

ENV ARCH=$TARGETARCH

RUN curl -sL https://github.com/goreleaser/goreleaser/releases/latest/download/goreleaser_Linux_x86_64.tar.gz | tar -xz -C /usr/local/bin

RUN go version && goreleaser --version

WORKDIR /app

# copy only folders needed for cli build
COPY ./.git ./.git

COPY ./cli/cli/ ./cli/cli

COPY ./scripts/ ./scripts/

COPY ./version.txt ./version.txt

COPY ./api ./api

COPY ./cloud ./cloud

COPY ./container-engine-lib ./container-engine-lib

COPY ./contexts-config-store ./contexts-config-store

COPY ./contexts-config-store ./contexts-config-store

COPY ./kurtosis_version ./kurtosis_version

COPY ./path-compression ./path-compression

COPY ./grpc-file-transfer ./grpc-file-transfer

COPY ./engine ./engine

COPY ./metrics-library ./metrics-library

RUN cd ./cli/cli/ && go mod tidy


RUN chmod u+x ./cli/cli/scripts/dev-img-entrypoint.sh

RUN ./cli/cli/scripts/build-dev-img.sh
# give  fake machine id
RUN echo 1234 > /etc/machine-id

ENTRYPOINT [ "./cli/cli/scripts/dev-img-entrypoint.sh" ]

