FROM  alpine:latest

RUN apk update && apk add bash

WORKDIR /run

COPY ./build/kurtosis-engine ./run-engine.sh ./

CMD bash ./run-engine.sh