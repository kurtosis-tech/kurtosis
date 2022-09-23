FROM  alpine:3.16.2

RUN apk update && apk add bash

WORKDIR /run

COPY ./build/api-container ./

CMD ./api-container