set -euo pipefail
script_dirpath="$(cd "$(dirname "${0}")" && pwd)"

bash "${script_dirpath}/build_initializer_binary.sh"
bash "${script_dirpath}/build_api_image.sh"
bash "${script_dirpath}/run.sh" ${*:-}
