#! /bin/zsh

# Parse command line arguments
REBUILD=false
ARGS_FILES=(
    "./tests/mev.yaml",
    # "./tests/mev-mock.yaml",
    # "./tests/mix-with-tools.yaml",
    # "./tests/mix-persistence.yaml",
    # "./tests/mix-public.yaml",
    # "./tests/minimal.yaml",
)

PACKAGE_ID="github.com/ethpandaops/ethereum-package"  # Default package
OUTPUT_DIR_PREFIX="non-parallel"

while [[ $# -gt 0 ]]; do
  case $1 in
    -r|--rebuild)
      REBUILD=true
      shift
      ;;
    --args-files)
      # Collect all args files until we hit another flag or end of arguments
      shift
      while [[ $# -gt 0 ]] && [[ ! "$1" =~ ^-- ]]; do
        ARGS_FILES+=("$1")
        shift
      done
      ;;
    --package-id)
      PACKAGE_ID="$2"
      shift 2
      ;;
    --output-dir-prefix)
      OUTPUT_DIR_PREFIX="$2"
      shift 2
      ;;
    *)
      shift
      ;;
  esac
done

# Validate args files
if [ ${#ARGS_FILES[@]} -eq 0 ]; then
  echo "Error: No args files provided. Use --args-files to specify one or more args files."
  exit 1
fi

for args_file in "${ARGS_FILES[@]}"; do
  if [ ! -f "$args_file" ]; then
    echo "Error: Args file '$args_file' does not exist"
    exit 1
  fi
done

# Run benchmark for each args file
for args_file in "${ARGS_FILES[@]}"; do
  echo "Running benchmark with args file: $args_file"
  
  # Generate output directory name based on args file name
  args_file_basename=$(basename "$args_file" .json)
  output_dir="${OUTPUT_DIR_PREFIX}-${args_file_basename}"
  
  # Construct benchmark command
  BENCHMARK_CMD="source benchmark.sh"
  if [ "$REBUILD" = true ]; then
    BENCHMARK_CMD="$BENCHMARK_CMD --rebuild"
  fi
  BENCHMARK_CMD="$BENCHMARK_CMD --args-file $args_file --package-id $PACKAGE_ID --output-dir $output_dir"
  
  echo "Running command: $BENCHMARK_CMD"
  eval $BENCHMARK_CMD
  
  echo "Completed benchmark for $args_file"
  echo "----------------------------------------"
done

echo "All benchmarks completed!"