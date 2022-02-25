FROM  envoyproxy/envoy-alpine:v1.21.0

RUN apk update && apk add bash

COPY ./envoy.yaml /etc/envoy/envoy.yaml
RUN chmod go+r /etc/envoy/envoy.yaml
ENV ENVOY_UID=0

WORKDIR /run

COPY ./build/kurtosis-engine ./run-engine-and-envoy.sh ./

CMD bash ./run-engine-and-envoy.sh