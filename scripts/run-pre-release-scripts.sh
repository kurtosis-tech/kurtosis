#!/usr/bin/env bash
# 2021-07-08 WATERMARK, DO NOT REMOVE - This script was generated from the Kurtosis Bash script template

set -euo pipefail   # Bash "strict mode"
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root_dirpath="$(dirname "${script_dirpath}")"



# ==================================================================================================
#                                             Constants
# ==================================================================================================
PER_RELEASE_SCRIPTS_RELATIVE_FILEPATHS=(
  core/scripts/update-own-version-constants.sh
  engine/scripts/update-own-version-constants.sh
  cli/scripts/update-own-version-constants.sh
)



# ==================================================================================================
#                                       Arg Parsing & Validation
# ==================================================================================================
show_helptext_and_exit() {
    echo "Usage: $(basename "${0}")"
    exit 1
}

# ==================================================================================================
#                                             Main Logic
# ==================================================================================================
if ! docker_tag="$(kudet get-docker-tag)"; then
    echo "Error: Couldn't get the Docker image tag" >&2
    exit 1
fi


# if the docker_tag isn't dirty we add -dirty to it
# we do this as we could be in an uncommitted state locally
# also if we set the non-dirty version and then build any image
# it would pick the dirty version as the tag anyway as the repo has changed
# assuming -dirty from the beginning makes this easier
if [[ "${docker_tag}" != *"dirty"* ]]; then
  docker_tag="${docker_tag}-dirty"
fi

for pre_release_script_rel_filepath in "${PER_RELEASE_SCRIPTS_RELATIVE_FILEPATHS[@]}"; do
    pre_release_script_abs_filepath="${root_dirpath}/${pre_release_script_rel_filepath}"
    if ! bash "${pre_release_script_abs_filepath}" "${docker_tag}"; then
        echo "Error: Pre release script '${pre_release_script_abs_filepath}' failed" >&2
        exit 1
    fi
done
