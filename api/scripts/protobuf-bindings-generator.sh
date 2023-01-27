#!/usr/bin/env bash

# This script regenerates Go bindings corresponding to the .proto files that define the API container's API
# It requires the Golang Protobuf extension to the 'protoc' compiler, as well as the Golang gRPC extension

set -euo pipefail
# script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")"; pwd)"



# ==================================================================================================
#                                           Constants
# ==================================================================================================
# ------------------------------------------- Shared -----------------------------------------------
PROTOC_CMD="protoc"
PROTOBUF_FILE_EXT=".proto"
PROTOC_INSTALL_COMMAND="brew install protobuf"

GOLANG_LANG="golang"
TYPESCRIPT_LANG="typescript"
SUPPORTED_LANGS=(
    "${GOLANG_LANG}"
    "${TYPESCRIPT_LANG}"
)

# Each language's file extension so we know what files need deleting before we regenerate bindings
declare -A FILE_EXTENSIONS
FILE_EXTENSIONS["${GOLANG_LANG}"]=".go"
FILE_EXTENSIONS["${TYPESCRIPT_LANG}"]=".ts"

# "Schema" of the function provided as a value of this map:
# generate_XXX_bindings(input_abs_dirpath, output_abs_dirpath) where:
#  1. The input_abs_dirpath is an ABSOLUTE path to a directory containing .proto files
#  2. The output_abs_dirpath is an ABSOLUTE path to the directory to output the generated files to
declare -A generators
generators["${GOLANG_LANG}"]="generate_golang_bindings"
generators["${TYPESCRIPT_LANG}"]="generate_typescript_bindings"

# ------------------------------------------- Golang -----------------------------------------------
GO_MOD_FILEPATH_ENV_VAR="GO_MOD_FILEPATH"
GO_MOD_FILE_MODULE_KEYWORD="module"  # Keyword in the go.mod file for specifying the module

# ----------------------------------------- Typescript -----------------------------------------------
NODE_GRPC_TOOLS_PROTOC_BIN_FILENAME="grpc_tools_node_protoc"    # For some reason, Node gRPC has its own 'protoc' binary
NODE_GRPC_TOOLS_PROTOC_PLUGIN_BIN_FILENAME="grpc_tools_node_protoc_plugin"  # The name of the plugin binary that will be used by 'protoc'
NODE_GRPC_TOOLS_INSTALL_COMMAND="npm install -g grpc-tools"
WEB_GRPC_PROTOC_BIN_FILENAME="protoc-gen-grpc-web"
WEB_GRPC_INSTALL_COMMAND="brew install protoc-gen-grpc-web"

# ==================================================================================================
#                                           Main Logic
# ==================================================================================================

show_helptext() {
    echo "Usage: $(basename "${0}") input_dirpath output_dirpath lang"
    echo ""
    echo "  input_dirpath   The directory containing ${PROTOBUF_FILE_EXT} files to generate bindings for (must exist)"
    echo "  output_dirpath  The directory to output the bindings in (must exist and must not be '/')"
    echo "  lang            The language to generate the output bindings in (options: $(IFS="|"; echo "${SUPPORTED_LANGS[*]}"))"
    echo ""
}

input_dirpath="${1:-}"
output_dirpath="${2:-}"
lang="${3:-}"
if [ -z "${input_dirpath}" ]; then
    echo "Error: Input dirpath must not be empty" >&2
    show_helptext
    exit 1
fi
if ! [ -d "${input_dirpath}" ]; then
    echo "Error: Input directory '${input_dirpath}' doesn't exist" >&2
    show_helptext
    exit 1
fi
if [ -z "${output_dirpath}" ]; then
    echo "Error: Output dirpath must not be empty" >&2
    show_helptext
    exit 1
fi
if ! [ -d "${output_dirpath}" ]; then
    echo "Error: Output directory '${output_dirpath}' doesn't exist" >&2
    show_helptext
    exit 1
fi
if [ "${output_dirpath}" == "/" ]; then
    echo "Error: Output directory must not be '/'"
    show_helptext
    exit 1
fi
if [ -z "${lang}" ]; then
    echo "Error: Lang must not be empty" >&2
    show_helptext
    exit 1
fi
valid_lang="false"
for supported_lang in "${SUPPORTED_LANGS[@]}"; do
    if [ "${lang}" == "${supported_lang}" ]; then
        valid_lang="true"
        break
    fi
done
if ! "${valid_lang}" ; then
    echo "Error: invalid output lang '${lang}'" >&2
    show_helptext
    exit 1
fi

case "${lang}" in
    "${GOLANG_LANG}")
        if ! env | grep "${GO_MOD_FILEPATH_ENV_VAR}" >/dev/null; then
            echo "Error: Variable '${GO_MOD_FILEPATH_ENV_VAR}' must be set to the go.mod filepath when lang is '${GOLANG_LANG}'" >&2
            exit 1
        fi
        if ! [ -f "${!GO_MOD_FILEPATH_ENV_VAR}" ]; then
            echo "Error: Variable '${GO_MOD_FILEPATH_ENV_VAR}' doesn't point to a go.mod file" >&2
            exit 1
        fi
        ;;
    "${TYPESCRIPT_LANG}")
        ;;
    *)
        echo "Error: Unrecognized lang '${lang}'; this is a bug in this script (likely indicating that a new language was added but this case statement wasn't updated)" >&2
        exit 1
        ;;
esac


# ------------------------------------------- Golang -----------------------------------------------
generate_golang_bindings() {

    input_abs_dirpath="${1}"
    output_abs_dirpath="${2}"

    go_mod_expanded_abs_filepath="$(cd "$(dirname "${!GO_MOD_FILEPATH_ENV_VAR}")" && pwd)/$(basename "${!GO_MOD_FILEPATH_ENV_VAR}")"
    go_module="$(grep "^${GO_MOD_FILE_MODULE_KEYWORD}" "${go_mod_expanded_abs_filepath}" | awk '{print $2}')"
    if [ "${go_module}" == "" ]; then
        echo "Error: Could not extract Go module from file '${go_mod_expanded_abs_filepath}'" >&2
        exit 1
    fi
    go_mod_file_parent_dirpath="$(dirname "${go_mod_expanded_abs_filepath}")"

    # If the output dirpath isn't a child directory of the go.mod file's parent directory, this means that we can't construct the relative path
    if ! [[ "${output_abs_dirpath}" == "${go_mod_file_parent_dirpath}"* ]]; then
        echo "Error: The output dirpath '${output_abs_dirpath}' isn't a subdirectory of the go.mod file's parent directory '${go_mod_file_parent_dirpath}'" >&2
        return 1
    fi

    output_rel_dirpath_with_leading_slash="${output_abs_dirpath##"${go_mod_file_parent_dirpath}"}"
    output_rel_dirpath="${output_rel_dirpath_with_leading_slash##/}"

    if ! command -v "${PROTOC_CMD}" > /dev/null; then
        echo "Error: No '${PROTOC_CMD}' command found; you'll need to install it via '${PROTOC_INSTALL_COMMAND}'" >&2
        return 1
    fi

    go_out_flag="--go_out=${output_abs_dirpath}"
    go_grpc_out_flag="--go-grpc_out=${output_abs_dirpath}"

    fully_qualified_go_pkg="${go_module}/${output_rel_dirpath}"
    fully_qualified_go_pkg="${fully_qualified_go_pkg%%/}"
    for input_filepath in $(find "${input_abs_dirpath}" -type f -name "*${PROTOBUF_FILE_EXT}"); do
        # Rather than specify the go_package in source code (which means all consumers of these protobufs would get it),
        #  we specify the go_package here per https://developers.google.com/protocol-buffers/docs/reference/go-generated
        # See also: https://github.com/golang/protobuf/issues/1272
        go_module_flag="--go_opt=module=${fully_qualified_go_pkg}"
        go_grpc_module_flag="--go-grpc_opt=module=${fully_qualified_go_pkg}"

        # Way back in the day, GRPC's Go binding generation used to create an interface for servers to implement
        # When you added a new method in the .proto file, the interface would get a new method, your implementation of the
        # interface wouldn't have that method, and you'd get a compile error
        # The GRPC Go team decided they didn't like that adding new methods caused compile breaks, because it would break backwards
        # compatibility. They then required implementations of the interface to do this crazy struct-embedding thing such that new
        # methods in the .proto wouldn't cause a compile break.
        # We, and many other people, were very unhappy about this (compile breaks are always better than runtime breaks) and there
        # was a public outcry here: https://github.com/grpc/grpc-go/issues/3669
        # The Go team relented, and provided this flag to allow us to disable their requirement. We use the flag, because compile
        # breaks really are better than runtime breaks.
        go_grpc_unimplemented_servers_flag="--go-grpc_opt=require_unimplemented_servers=false"

        if ! "${PROTOC_CMD}" \
                -I="${input_abs_dirpath}" \
                "${go_out_flag}" \
                "${go_module_flag}" \
                "${go_grpc_out_flag}" \
                "${go_grpc_module_flag}" \
                "${go_grpc_unimplemented_servers_flag}" \
                "${input_filepath}"; then
            echo "Error: An error occurred generating Golang bindings for file '${input_filepath}'" >&2
            return 1
        fi
    done
}

# ------------------------------------------ TypeScript -----------------------------------------------
generate_typescript_bindings() {

    input_abs_dirpath="${1}"
    output_abs_dirpath="${2}"

    if ! node_protoc_bin_filepath="$(which "${NODE_GRPC_TOOLS_PROTOC_BIN_FILENAME}")"; then
        echo "Error: Couldn't find Node gRPC tools protoc binary '${NODE_GRPC_TOOLS_PROTOC_BIN_FILENAME}' on the PATH; have you installed the tools with '${NODE_GRPC_TOOLS_INSTALL_COMMAND}'?" >&2
        return 1
    fi
    if [ -z "${node_protoc_bin_filepath}" ]; then
        echo "Error: Got an empty filepath when looking for the Node gRPC tools protoc binary '${NODE_GRPC_TOOLS_PROTOC_BIN_FILENAME}'; have you installed the tools with '${NODE_GRPC_TOOLS_INSTALL_COMMAND}'?" >&2
        return 1
    fi
    if ! web_protoc_bin_filepath="$(which "${WEB_GRPC_PROTOC_BIN_FILENAME}")"; then
        echo "Error: Couldn't find gRPC Web tool protoc binary '${WEB_GRPC_PROTOC_BIN_FILENAME}' on the PATH; have you installed the tools with '${WEB_GRPC_INSTALL_COMMAND}'?" >&2
        return 1
    fi
    if [ -z "${web_protoc_bin_filepath}" ]; then
        echo "Error: Got an empty filepath when looking for the gRPC Web tool protoc binary '${WEB_GRPC_PROTOC_BIN_FILENAME}'; have you installed the tools with '${WEB_GRPC_INSTALL_COMMAND}'?" >&2
        return 1
    fi
    if ! node_protoc_plugin_bin_filepath="$(which "${NODE_GRPC_TOOLS_PROTOC_PLUGIN_BIN_FILENAME}")"; then
        echo "Error: Couldn't find Node gRPC tools protoc plugin binary '${NODE_GRPC_TOOLS_PROTOC_PLUGIN_BIN_FILENAME}' on the PATH; have you installed the tools with '${NODE_GRPC_TOOLS_INSTALL_COMMAND}'?" >&2
        return 1
    fi
    if [ -z "${node_protoc_plugin_bin_filepath}" ]; then
        echo "Error: Got an empty filepath when looking for the Node gRPC tools protoc plugin binary '${NODE_GRPC_TOOLS_PROTOC_PLUGIN_BIN_FILENAME}'; have you installed the tools with '${NODE_GRPC_TOOLS_PROTOC_PLUGIN_BIN_FILENAME}'?" >&2
        return 1
    fi

    for input_filepath in $(find "${input_abs_dirpath}" -type f -name "*${PROTOBUF_FILE_EXT}"); do
      # NOTE: Generating Node bindings
        if ! "${node_protoc_bin_filepath}" \
                -I="${input_abs_dirpath}" \
                "--js_out=import_style=commonjs,binary:${output_abs_dirpath}" \
                `# NOTE: we pass the grpc_js option to generate code using '@grpc/grpc-js', as the old 'grpc' package is deprecated` \
                "--grpc_out=grpc_js:${output_abs_dirpath}" \
                "--plugin=protoc-gen-grpc=${node_protoc_plugin_bin_filepath}" \
                `# NOTE: we pass mode=grpc-js to get Typescript definition files that use '@grpc/grpc-js' rather than 'grpc' `\
                "--ts_out=service=grpc-node,mode=grpc-js:${output_abs_dirpath}" \
                "${input_filepath}"; then
            echo "Error: An error occurred generating TypeScript Node bindings for file '${input_filepath}'" >&2
            return 1
        fi
        # NOTE: Generating Web bindings
        if ! "${node_protoc_bin_filepath}" \
                -I="${input_abs_dirpath}" \
                "--js_out=import_style=commonjs:${output_abs_dirpath}" \
                "--grpc-web_out=import_style=commonjs+dts,mode=grpcwebtext:${output_abs_dirpath}" \
                "${input_filepath}"; then
            echo "Error: An error occurred generating TypeScript Web bindings for file '${input_filepath}'" >&2
            return 1
        fi
    done
}

# ------------------------------------------ Shared Code-----------------------------------------------
input_abs_dirpath="$(cd "$(dirname "${input_dirpath}")" && pwd)/$(basename "${input_dirpath}")"
output_abs_dirpath="$(cd "$(dirname "${output_dirpath}")" && pwd)/$(basename "${output_dirpath}")"

file_ext="${FILE_EXTENSIONS["${lang}"]}"
if [ -z "${file_ext}" ]; then
    echo "Error: No file extension associated with lang '${lang}'; this is a bug in this script" >&2
    exit 1
fi
if ! find "${output_abs_dirpath}" -name "*${file_ext}" -delete; then
    echo "Error: An error occurred removing the existing ${lang} bindings at '${output_abs_dirpath}'" >&2
    exit 1
fi

generator_func="${generators["${lang}"]}"

# TODO When multiple people start developing on this, we won't be able to rely on using the user's local environment for generating bindings because the environments
# might differ across users
# We'll need to standardize by:
#  1) Using protoc inside the API container Dockerfile to generate the output Go files (standardizes the output files for Docker)
#  2) Using the user's protoc to generate the output Go files on the local machine, so their IDEs will work
#  3) Tying the protoc inside the Dockerfile and the protoc on the user's machine together using a protoc version check
#  4) Adding the locally-generated Go output files to .gitignore
#  5) Adding the locally-generated Go output files to .dockerignore (since they'll get generated inside Docker)
"${generator_func}" "${input_abs_dirpath}" "${output_abs_dirpath}"
