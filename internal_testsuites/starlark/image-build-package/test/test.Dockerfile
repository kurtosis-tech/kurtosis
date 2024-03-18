FROM alpine:latest

WORKDIR /app

COPY . .

CMD ["echo", "Hello, Kurtosis!"]
