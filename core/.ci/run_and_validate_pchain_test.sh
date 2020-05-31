set -x

SCRIPTS_PATH=$(cd $(dirname "${BASH_SOURCE[0]}"); pwd)
KURTOSIS_PATH=$(dirname "${SCRIPTS_PATH}")

LATEST_KURTOSIS_TAG="kurtosistech/kurtosis:latest"
LATEST_CONTROLLER_TAG="kurtosistech/ava-test-controller:latest"
DEFAULT_GECKO_IMAGE="kurtosistech/gecko:latest"

bash "${KURTOSIS_PATH}"/scripts/build_image.sh ${LATEST_KURTOSIS_TAG}
docker pull ${LATEST_CONTROLLER_TAG}
docker pull ${DEFAULT_GECKO_IMAGE}

(docker run -v /var/run/docker.sock:/var/run/docker.sock \
--env DEFAULT_GECKO_IMAGE="${DEFAULT_GECKO_IMAGE}" \
--env TEST_CONTROLLER_IMAGE="${LATEST_CONTROLLER_TAG}" \
${LATEST_KURTOSIS_TAG}) &

kurtosis_pid=$!

sleep 15
docker ps -a
kill ${kurtosis_pid}

ACTUAL_EXIT_STATUS=$(docker ps -a --latest --filter ancestor=kurtosistech/ava-test-controller:latest --format="{{.Status}}")
EXPECTED_EXIT_STATUS="Exited \(0\).*"

echo "Exit status: ${ACTUAL_EXIT_STATUS}"

# Clear containers.
echo "Clearing kurtosis testnet containers."
docker rm $(docker stop $(docker ps -a -q --filter ancestor="${DEFAULT_GECKO_IMAGE}" --format="{{.ID}}")) >/dev/null

if [[ ${ACTUAL_EXIT_STATUS} =~ ${EXPECTED_EXIT_STATUS} ]]
then
  echo "Kurtosis test succeeded."
  exit 0
else
  echo "Kurtosis test failed."
  exit 1
fi
