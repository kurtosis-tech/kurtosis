set -euo pipefail
root_dirpath="$(cd "$(dirname "${0}")" && pwd)"

echo "Generating API container code from protobufs..."
api_protobufs_dirpath="${root_dirpath}/api_container/api_protobufs"
if [ "$(ls -A ${api_protobufs_dirpath})" == "" ]; then
    echo "Error: No API container protobufs found; you'll need to initialize the protobufs submodule with 'git submodule init' then 'git submodule update'" >&2
    exit 1
fi
echo "Successfully generated API container code from protobufs"
