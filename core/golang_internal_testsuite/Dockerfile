FROM alpine:3.12

WORKDIR /run

# Copy the code into the container
COPY golang-internal-testsuite .

# Copy local static files inside the testsuite image
COPY golang_internal_testsuite/static_files /static-files

CMD ./golang-internal-testsuite
