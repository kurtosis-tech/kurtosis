FROM alpine:3.24

RUN apk update && apk add tar

ARG TARGETARCH

WORKDIR /run

COPY ./build/files-artifacts-expander.$TARGETARCH  ./files-artifacts-expander

CMD ./files-artifacts-expander
