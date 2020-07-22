if ! git diff --name-only HEAD.."${TRAVIS_BRANCH}" | grep CHANGELOG.md; then
  echo "PR has no CHANGELOG entry. Please update the CHANGELOG!"
fi
