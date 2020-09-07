set -euo pipefail
script_dirpath="$(cd "$(dirname "${0}")" && pwd)"

bash "${script_dirpath}/build_images.sh"
bash "${script_dirpath}/run.sh" ${*:-}
