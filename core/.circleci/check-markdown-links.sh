set -euo pipefail
script_dirpath="$(cd "$(dirname "${0}")" && pwd)"
root_dirpath="$(dirname "${script_dirpath}")"

# Inspired by https://github.com/open-telemetry/opentelemetry-collector/pull/1156/files/2244e61f4dd0378deffc00d939edf6f800687dcf
exit_code=0
for filepath in $(find "${root_dirpath}" -iname '*.md' | sort); do
    markdown-link-check -qv "${filepath}" || exit_code=1
    # Wait to scan files so that we don't overload github with requests which may result in 429 responses
    sleep 2
done

exit "${exit_code}"
