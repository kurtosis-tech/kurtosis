set -euo pipefail
# TODO debugging
echo "${CIRCLE_BRANCH}"
git diff --name-only HEAD.."${CIRCLE_BRANCH}"
if ! git diff --name-only HEAD.."${CIRCLE_BRANCH}" | grep CHANGELOG.md; then
  echo "PR has no CHANGELOG entry. Please update the CHANGELOG!"
  return_code=1
else
  return_code=0
fi
exit "${return_code}"
