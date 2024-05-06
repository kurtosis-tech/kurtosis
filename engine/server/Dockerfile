FROM alpine:3.17

RUN apk update && apk add bash

ARG TARGETARCH

WORKDIR /run

ENV KURTOSIS_ENGINE_LOCAL_RUN=true
ENV KURTOSIS_ENGINE_LOCAL_DOCKER_BRIDGE_IP=172.17.0.7

ADD ./webapp ./webapp

COPY ./build/kurtosis-engine.$TARGETARCH ./kurtosis-engine

CMD ./kurtosis-engine
