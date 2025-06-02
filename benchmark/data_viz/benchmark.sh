#! /bin/zsh

# Parse command line arguments
REBUILD=false
ARGS_FILE=""
PACKAGE_ID="github.com/ethpandaops/ethereum-package"  # Default package
OUTPUT_DIR=""
while [[ $# -gt 0 ]]; do
  case $1 in
    -r|--rebuild)
      REBUILD=true
      shift
      ;;
    --args-file)
      ARGS_FILE="$2"
      shift 2
      ;;
    --package-id)
      PACKAGE_ID="$2"
      shift 2
      ;;
    --output-dir)
      OUTPUT_DIR="$2"
      shift 2
      ;;
    *)
      shift
      ;;
  esac
done

# Validate args file if provided
if [ -n "$ARGS_FILE" ] && [ ! -f "$ARGS_FILE" ]; then
  echo "Error: Args file '$ARGS_FILE' does not exist"
  exit 1
fi

# Rebuild API container if requested
if [ "$REBUILD" = true ]; then
  echo "Rebuilding API container..."
  cd ../../core/scripts/
  ./build.sh
  cd -
fi

dkt enclave add --name benchmark-test --api-container-version $(get-devc-img "tag")
echo "Running test"

# Construct dkt run command with args file if provided
DKT_RUN_CMD="dkt run ${PACKAGE_ID} --enclave benchmark-test"
if [ -n "$ARGS_FILE" ]; then
  DKT_RUN_CMD="$DKT_RUN_CMD --args-file $ARGS_FILE"
fi

eval $DKT_RUN_CMD
echo "Finished running test"

API_CONTAINER=$(docker ps --filter "name=kurtosis-api--" --format "{{.Names}}" | head -n 1)
if [ -z "$OUTPUT_DIR" ]; then
  data_directory_name="./data-${API_CONTAINER}"
else
  data_directory_name="$OUTPUT_DIR"
fi
mkdir -p ${data_directory_name}
docker cp ${API_CONTAINER}:/run/benchmark-data/ ${data_directory_name}
echo "Pulled benchmark data from enclave"

python3 benchmark_viz.py ${data_directory_name}
echo "Created visualizations"

# dkt enclave rm -f benchmark-test
echo "Removed enclave"
