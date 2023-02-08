FROM --platform=linux/amd64 alpine:3.16.2

RUN apk update && apk add bash && apk add musl

WORKDIR /run

COPY ./build/kurtosis-engine ./

CMD ./kurtosis-engine
