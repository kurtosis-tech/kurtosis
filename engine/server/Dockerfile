FROM alpine:3.24

RUN apk update && apk add bash

ARG TARGETARCH

WORKDIR /run

ADD ./webapp ./webapp

COPY ./build/kurtosis-engine.$TARGETARCH ./kurtosis-engine

CMD ./kurtosis-engine
