return_code=1
if ! git diff --name-only HEAD.."${TRAVIS_BRANCH}" | grep CHANGELOG.md; then
  echo "PR has no CHANGELOG entry. Please update the CHANGELOG!"
  return_code=1
else
  return_code=0
fi
exit "${return_code}"
