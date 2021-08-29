FROM alpine:3.12

WORKDIR /run

COPY initializer .

# !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! IMPORTANT !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
#  1. Keep these flags sorted alphabetically, to avoid duplicates (which will cause subtle problems)
#  2. If you change any of these:
#       a. Update the flag-parsing in main.go
#       b. Update the template wrapper script, which passes in these environment variables
# !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! IMPORTANT !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
# vvvvvvvvvvvvv Read the IMPORTANT section before modifying this! vvvvvvvvvvvvvvvvvvvvvvvv
CMD ["sh", "-c", "set -euo pipefail && ./initializer \
    --CLIENT_ID=\"${CLIENT_ID}\" \
    --CLIENT_SECRET=\"${CLIENT_SECRET}\" \
    --CUSTOM_PARAMS_JSON=\"${CUSTOM_PARAMS_JSON}\" \
    --DO_LIST=\"${DO_LIST}\" \
    --EXECUTION_ID=\"${EXECUTION_ID}\" \
    --IS_DEBUG_MODE=\"${IS_DEBUG_MODE}\" \
    --KURTOSIS_API_IMAGE=\"${KURTOSIS_API_IMAGE}\" \
    --KURTOSIS_LOG_LEVEL=\"${KURTOSIS_LOG_LEVEL}\" \
    --PARALLELISM=\"${PARALLELISM}\" \
    --TEST_NAMES=\"${TEST_NAMES}\" \
    --TEST_SUITE_IMAGE=\"${TEST_SUITE_IMAGE}\" \
    --TEST_SUITE_LOG_LEVEL=\"${TEST_SUITE_LOG_LEVEL}\" \
"]
# ^^^^^^^^^^^^^ Read the IMPORTANT section before modifying this! ^^^^^^^^^^^^^^^^^^^^^^^^

# Code notes about the above:
# - This is the "exec" form of the command to ensure that SIGTERM gets passed to the API server
# - We call "sh" as the first command so that environment variable expansion will still happen
