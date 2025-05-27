#! /bin/zsh

# set -e pipefail

dkt enclave add --name benchmark-test --api-container-version $(get-devc-img "tag")
echo "Running test"

go test -v /Users/tewodrosmitiku/craft/kurtosis/internal_testsuites/golang/testsuite/benchmark/benchmark_test.go
echo "Finished running test"

API_CONTAINER=$(docker ps --filter "name=kurtosis-api--" --format "{{.Names}}" | head -n 1)
data_directory_name="./data-${API_CONTAINER}"
docker cp ${API_CONTAINER}:/run/benchmark-data/ ${data_directory_name}
echo "Pulled benchmark data from enclave"

python3 benchmark_viz.py ${data_directory_name}
echo "Created visualizations"

dkt enclave rm --name benchmark-test
echo "Removed enclave"
