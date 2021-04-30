# This script regenerates Go bindings corresponding to all .proto files inside this project
# It requires the Golang Protobuf extension to the 'protoc' compiler, as well as the Golang gRPC extension

set -euo pipefail
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")"; pwd)"
root_dirpath="$(dirname "${script_dirpath}")"

# ================================ CONSTANTS =======================================================
GO_MOD_FILENAME="go.mod"
GO_MOD_FILE_MODULE_KEYWORD="module"

# The name of the directory to contain generated Protobuf Go bindings
BINDINGS_OUTPUT_DIRNAME="bindings"

# Dirpaths, relative to the repo root, where protobuf files live
# The bindings for the Protobuf files will be generated inside a 'bindings' directory in this directory
PROTOBUF_RELATIVE_DIRPATHS=(
    "api_container/api"
    "test_suite/api"
)

# =============================== MAIN LOGIC =======================================================
go_mod_filepath="${root_dirpath}/${GO_MOD_FILENAME}"
if ! [ -f "${go_mod_filepath}" ]; then
    echo "Error: Could not get Go module name; no ${GO_MOD_FILENAME} found in root of repo" >&2
    exit 1
fi
go_module="$(grep "^${GO_MOD_FILE_MODULE_KEYWORD}" "${go_mod_filepath}" | awk '{print $2}')"
if [ "${go_module}" == "" ]; then
    echo "Error: Could not extract Go module from ${go_mod_filepath}" >&2
    exit 1
fi

for input_relative_dirpath in "${PROTOBUF_RELATIVE_DIRPATHS[@]}"; do
    input_abs_dirpath="${root_dirpath}/${input_relative_dirpath}"
    if ! [ -d "${input_abs_dirpath}" ]; then
        echo "Error: Dirpath '${input_abs_dirpath}' containing Protobuf files doesn't exist" >&2
        exit 1
    fi

    output_relative_dirpath="${input_relative_dirpath}/${BINDINGS_OUTPUT_DIRNAME}"

    output_abs_dirpath="${root_dirpath}/${output_relative_dirpath}"
    if [ "${output_abs_dirpath}/" == "/" ]; then
        echo "Error: Binding output dirpath cannot be empty!" >&2
        exit 1
    fi

    if ! mkdir -p "${output_abs_dirpath}"; then
        echo "Error: An error occurred creating output directory '${output_abs_dirpath}'" >&2
        exit 1
    fi

    if ! find ${output_abs_dirpath} -name '*.go' -delete; then
        echo "Error: An error occurred removing the existing protobuf-generated code from '${output_abs_dirpath}'" >&2
        exit 1
    fi

    fully_qualified_go_pkg="${go_module}/${output_relative_dirpath}"
    for protobuf_filepath in $(find "${input_abs_dirpath}" -name "*.proto"); do
        protobuf_filename="$(basename "${protobuf_filepath}")"

        # NOTE: When multiple people start developing on this, we won't be able to rely on using the user's local protoc because they might differ. We'll need to standardize by:
        #  1) Using protoc inside the API container Dockerfile to generate the output Go files (standardizes the output files for Docker)
        #  2) Using the user's protoc to generate the output Go files on the local machine, so their IDEs will work
        #  3) Tying the protoc inside the Dockerfile and the protoc on the user's machine together using a protoc version check
        #  4) Adding the locally-generated Go output files to .gitignore
        #  5) Adding the locally-generated Go output files to .dockerignore (since they'll get generated inside Docker)
        if ! protoc \
                -I="${input_abs_dirpath}" \
                --go_out="plugins=grpc:${output_abs_dirpath}" \
                `# Rather than specify the go_package in source code (which means all consumers of these protobufs would get it),` \
                `#  we specify the go_package here per https://developers.google.com/protocol-buffers/docs/reference/go-generated` \
                `# See also: https://github.com/golang/protobuf/issues/1272` \
                --go_opt="M${protobuf_filename}=${fully_qualified_go_pkg};$(basename "${fully_qualified_go_pkg}")" \
                "${protobuf_filepath}"; then
            echo "Error: An error occurred converting Protobuf file '${protobuf_filepath}' to bindings in directory '${output_abs_dirpath}'" >&2
            exit 1
        fi
        echo "Successfully generated bindings for Protobuf file '${protobuf_filepath}' in directory '${output_abs_dirpath}'"
    done
done
