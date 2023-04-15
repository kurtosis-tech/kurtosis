FROM alpine:3.17

RUN apk update && apk add tar

WORKDIR /run

COPY ./build/files-artifacts-expander  ./

CMD ./files-artifacts-expander
