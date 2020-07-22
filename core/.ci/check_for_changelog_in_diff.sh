current_branch=$(git rev-parse --abbrev-ref HEAD)

# If current branch and travis branch are the same, we're running the "branch" check and it should pass.
if [[ "${current_branch}" == "${TRAVIS_BRANCH}" ]]; then
  exit 0
fi


return_code=1
if ! git diff --name-only HEAD.."${TRAVIS_BRANCH}" | grep CHANGELOG.md; then
  echo "PR has no CHANGELOG entry. Please update the CHANGELOG!"
  return_code=1
else
  return_code=0
fi
exit "${return_code}"

