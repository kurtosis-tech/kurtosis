# !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! WARNING !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
#
#      Do not modify this file! It will get overwritten when you upgrade Kurtosis!
#
# !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! WARNING !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!

set -euo pipefail



# ============================================================================================
#                                      Constants
# ============================================================================================
# The directory where Kurtosis will store files it uses in between executions, e.g. access tokens
# Can make this configurable if needed
KURTOSIS_DIRPATH="${HOME}/.kurtosis"

KURTOSIS_CORE_TAG="{{.KurtosisCoreVersion}}"
KURTOSIS_DOCKERHUB_ORG="kurtosistech"
INITIALIZER_IMAGE="${KURTOSIS_DOCKERHUB_ORG}/kurtosis-core_initializer:${KURTOSIS_CORE_TAG}"
API_IMAGE="${KURTOSIS_DOCKERHUB_ORG}/kurtosis-core_api:${KURTOSIS_CORE_TAG}"

POSITIONAL_ARG_DEFINITION_FRAGMENTS=2



# ============================================================================================
#                                      Arg Parsing
# ============================================================================================
function print_help_and_exit() {
    echo ""
    echo "$(basename "${0}") {{.OneLinerHelpText}}"
    echo ""
    {{range $text := .LinewiseHelpText}}echo "{{$text}}"
    {{end}}
    echo ""
    exit 1  # Exit with an error code, so that if it gets accidentally called in parent scripts/CI it fails loudly
}



# ============================================================================================
#                                      Arg Parsing
# ============================================================================================
{{range $variable, $value := .DefaultValues}}{{$variable}}="{{$value}}"
{{end}}


POSITIONAL=()
while [ ${#} -gt 0 ]; do
    key="${1}"
    case "${key}" in
        {{range $flagArg := .FlagArgParsingData }}
        {{$flagArg.Flag}})
            {{if $flagArg.DoStoreTrue}}{{$flagArg.Variable}}="true"
            shift   # Shift to clear out the flag{{end}}
            {{if $flagArg.DoStoreValue}}{{$flagArg.Variable}}="${2}"
            shift   # Shift to clear out the flag
            shift   # Shift again to clear out the value{{end}}
            ;;
        {{end}}
        -*)
            echo "ERROR: Unrecognized flag '${key}'" >&2
            exit 1
            ;;
        *)
            POSITIONAL+=("${1}")
            shift
            ;;
    esac
done

# Restore positional parameters and assign them to variables
set -- "${POSITIONAL[@]}"
{{range $idx, $variable := .PositionalArgAssignment}}{{$variable}}="${{$idx}}"
{{end}}

if "${show_help}"; then
    print_help_and_exit
fi



# ============================================================================================
#                                    Arg Validation
# ============================================================================================
if [ "${#}" -ne {{.NumPositionalArgs}} ]; then
    echo "ERROR: Expected {{.NumPositionalArgs}} positional variables but got ${#}" >&2
    print_help_and_exit
fi

{{range $idx, $variable := .PositionalArgAssignment}}if [ -z "${{$variable}}" ]; then
    echo "ERROR: Variable '{{$variable}}' cannot be empty" >&2
    exit 1
fi{{end}}



# ============================================================================================
#                                    Main Logic
# ============================================================================================
{{- /* if this is a local dev branch of kurtosis-core, we skip the docker pull because it will always fail (because the images will by definition be local) */ -}}
{{- if .IsProductionRelease -}}
# Because Kurtosis X.Y.Z tags are normalized to X.Y so that minor patch updates are transparently 
#  used, we need to pull the latest API & initializer images
echo "Pulling latest versions of API & initializer image..."
if ! docker pull "${INITIALIZER_IMAGE}"; then
    echo "WARN: An error occurred pulling the latest version of the initializer image (${INITIALIZER_IMAGE}); you may be running an out-of-date version" >&2
else
    echo "Successfully pulled latest version of initializer image"
fi
if ! docker pull "${API_IMAGE}"; then
    echo "WARN: An error occurred pulling the latest version of the API image (${API_IMAGE}); you may be running an out-of-date version" >&2
else
    echo "Successfully pulled latest version of API image"
fi
{{- end}}

# Kurtosis needs a Docker volume to store its execution data in
# To learn more about volumes, see: https://docs.docker.com/storage/volumes/
sanitized_image="$(echo "${test_suite_image}" | sed 's/[^a-zA-Z0-9_.-]/_/g')"
suite_execution_volume="$(date +{{.VolumeTimestampDateFormat}})_${sanitized_image}"
if ! docker volume create "${suite_execution_volume}" > /dev/null; then
    echo "ERROR: Failed to create a Docker volume to store the execution files in" >&2
    exit 1
fi

if ! mkdir -p "${KURTOSIS_DIRPATH}"; then
    echo "ERROR: Failed to create the Kurtosis directory at '${KURTOSIS_DIRPATH}'" >&2
    exit 1
fi

docker run \
    `# The Kurtosis initializer runs inside a Docker container, but needs to access to the Docker engine; this is how to do it` \
    `# For more info, see the bottom of: http://jpetazzo.github.io/2015/09/03/do-not-use-docker-in-docker-for-ci/` \
    --mount "type=bind,source=/var/run/docker.sock,target=/var/run/docker.sock" \
    \
    `# Because the Kurtosis initializer runs inside Docker but needs to persist & read files on the host filesystem between execution,` \
    `#  the container expects the Kurtosis directory to be bind-mounted at the special "/kurtosis" path` \
    --mount "type=bind,source=${KURTOSIS_DIRPATH},target=/kurtosis" \
    \
    `# The Kurtosis initializer image requires the volume for storing suite execution data to be mounted at the special "/suite-execution" path` \
    --mount "type=volume,source=${suite_execution_volume},target=/suite-execution" \
    \
    `# Keep these sorted alphabetically` \
    --env CLIENT_ID="${client_id}" \
    --env CLIENT_SECRET="${client_secret}" \
    --env CUSTOM_PARAMS_JSON="${custom_params_json}" \
    --env DO_LIST="${do_list}" \
    --env KURTOSIS_API_IMAGE="${API_IMAGE}" \
    --env KURTOSIS_LOG_LEVEL="${kurtosis_log_level}" \
    --env PARALLELISM="${parallelism}" \
    --env SUITE_EXECUTION_VOLUME="${suite_execution_volume}" \
    --env TEST_NAMES="${test_names}" \
    --env TEST_SUITE_IMAGE="${test_suite_image}" \
    --env TEST_SUITE_LOG_LEVEL="${test_suite_log_level}" \
    \
    "${INITIALIZER_IMAGE}"
