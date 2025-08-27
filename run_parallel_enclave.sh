#!/bin/zsh

# Parse command line arguments
SEQUENTIAL=false
num_services=3 # Default value

while [[ $# -gt 0 ]]; do
    case $1 in
        --sequential)
            SEQUENTIAL=true
            shift
            ;;
        --num-services)
            num_services="$2"
            shift 2
            ;;
        *)
            echo "Unknown option: $1"
            echo "Usage: $0 [--sequential] [--num-services <number>]"
            exit 1
            ;;
    esac
done
dkt enclave rm parallel-enclave --force

dkt enclave add --api-container-version $(get-devc-img "tag") --name parallel-enclave 

dkt run ~/craft/ethereum-package/ --enclave parallel-enclave --args-file ~/craft/ethereum-package/args.yml --non-blocking-tasks
# kt run ~/craft/ethereum-package/ --enclave parallel-enclave --args-file ~/craft/ethereum-package/args.yml --parallelism 1 --non-blocking-tasks # runs sequentially

# if [ "$SEQUENTIAL" = true ]; then
#     echo "Running in sequential mode..."
#     dkt run /Users/tewodrosmitiku/craft/sandbox/tests/ --enclave parallel-enclave --parallelism 1 --non-blocking-tasks
# else
#     echo "Running in parallel mode..."
#     dkt run /Users/tewodrosmitiku/craft/sandbox/tests/ --enclave parallel-enclave --non-blocking-tasks
# fi

# use docker to find the container id of the api container for parallel enclave
# then docker the file /tmp/dependency.dot on that containres to ~/Users/tewodrosmitiku/craft/graphs/dependency.dot
docker cp $(docker ps -q --filter "name=kurtosis-api--" | head -n 1):/tmp/dependency.dot ~/craft/graphs/dependency.dot
docker cp $(docker ps -q --filter "name=kurtosis-api--" | head -n 1):/tmp/execution.txt ~/craft/graphs/execution.txt
docker cp $(docker ps -q --filter "name=kurtosis-api--" | head -n 1):/tmp/instruction_durations.json ~/craft/graphs/instruction_durations.json
docker cp $(docker ps -q --filter "name=kurtosis-api--" | head -n 1):/tmp/run_starlark_package_cpu.prof ~/craft/sandbox/optimizations/run_starlark_package_cpu.prof
docker cp $(docker ps -q --filter "name=kurtosis-api--" | head -n 1):/tmp/run_starlark_package_block.prof ~/craft/sandbox/optimizations/run_starlark_package_block.prof
docker cp $(docker ps -q --filter "name=kurtosis-api--" | head -n 1):/tmp/run_starlark_package_mutex.prof ~/craft/sandbox/optimizations/run_starlark_package_mutex.prof


dot -Tpng ~/craft/graphs/dependency.dot -o ~/craft/graphs/graph.png

# Get API container logs
# save logs to a file with the current date and time
filename=$(date +%Y-%m-%d_%H-%M-%S)

if [ "$SEQUENTIAL" = true ]; then
    docker logs $(docker ps -q --filter "name=kurtosis-api--" | head -n 1) &> /Users/tewodrosmitiku/craft/kurtosis/${filename}-sequential-apilogs.txt
else
    docker logs $(docker ps -q --filter "name=kurtosis-api--" | head -n 1) &> /Users/tewodrosmitiku/craft/kurtosis/${filename}-parallel-apilogs.txt
fi

# Get service IDs from container names
# read_genesis_id=$(docker ps --format '{{.Names}}' | grep "read-genesis-validators-root" | sed 's/.*--\(.*\)/\1/')
# read_osaka_id=$(docker ps --format '{{.Names}}' | grep "read-osaka-time" | sed 's/.*--\(.*\)/\1/')
# check_if_osaka_id=$(docker ps --format '{{.Names}}' | grep "check-osaka-enabled" | sed 's/.*--\(.*\)/\1/')

# if [ "$SEQUENTIAL" = true ]; then
#     echo "read_genesis_id: ${read_genesis_id:0:10}" > /Users/tewodrosmitiku/craft/kurtosis/${filename}-sequential-times.txt
#     echo "read_osaka_id: ${read_osaka_id:0:10}" >> /Users/tewodrosmitiku/craft/kurtosis/${filename}-sequential-times.txt  
#     echo "check_if_osaka_id: ${check_if_osaka_id:0:10}" >> /Users/tewodrosmitiku/craft/kurtosis/${filename}-sequential-times.txt
# else
#     echo "read_genesis_id: ${read_genesis_id:0:10}" > /Users/tewodrosmitiku/craft/kurtosis/${filename}-parallel-times.txt
#     echo "read_osaka_id: ${read_osaka_id:0:10}" >> /Users/tewodrosmitiku/craft/kurtosis/${filename}-parallel-times.txt
#     echo "check_if_osaka_id: ${check_if_osaka_id:0:10}" >> /Users/tewodrosmitiku/craft/kurtosis/${filename}-parallel-times.txt
# fi

# # # Extract timing metrics from API logs
# if [ "$SEQUENTIAL" = true ]; then
#     source /Users/tewodrosmitiku/craft/kurtosis-parallelism-debugging/extract-times.sh /Users/tewodrosmitiku/craft/kurtosis/${filename}-sequential-apilogs.txt "${read_genesis_id}" "${read_osaka_id}" "${check_if_osaka_id}" &>> /Users/tewodrosmitiku/craft/kurtosis/${filename}-sequential-times.txt
# else 
#     source /Users/tewodrosmitiku/craft/kurtosis-parallelism-debugging/extract-times.sh /Users/tewodrosmitiku/craft/kurtosis/${filename}-parallel-apilogs.txt "${read_genesis_id}" "${read_osaka_id}" "${check_if_osaka_id}" &>> /Users/tewodrosmitiku/craft/kurtosis/${filename}-parallel-times.txt
# fi

if [ "$SEQUENTIAL" = true ]; then
    source /Users/tewodrosmitiku/craft/kurtosis-parallelism-debugging/extract-times.sh /Users/tewodrosmitiku/craft/kurtosis-parallelism-debugging/${filename}-sequential-apilogs.txt ${num_services} "sequential"
else 
    source /Users/tewodrosmitiku/craft/kurtosis-parallelism-debugging/extract-times.sh /Users/tewodrosmitiku/craft/kurtosis-parallelism-debugging/${filename}-parallel-apilogs.txt ${num_services} "parallel"
fi
