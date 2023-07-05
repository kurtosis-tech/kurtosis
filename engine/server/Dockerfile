FROM --platform=linux/amd64 alpine:3.17

RUN apk update && apk add bash

WORKDIR /run

ADD ./dist ./dist

COPY ./build/kurtosis-engine ./

CMD ./kurtosis-engine
