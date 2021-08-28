FROM alpine:3.12

WORKDIR /run

# Copy the code into the container
COPY internal-testsuite .

RUN find .

# Copy local static files inside the testsuite image
COPY internal_testsuite/static_files /static-files

CMD ./internal-testsuite
