### 0.1.5
* Update the quickstart docs with the new microservice examples in kurtosis-go 1.4.0

### 0.1.4
* Add link to website, with horizontable logo

### 0.1.3
* Update docs to match the service simplification refactor from kurtosis-go 1.3.0

### 0.1.2
* Add Discord link to all the tutorial pages

### 0.1.1
* Refactor the quickstart to contain step-by-step instructions to writing a testsuite (mostly ported from Kurtosis v0.1)
* Add CI to check links in Markdown documents

### 0.1.0
* Update the "Running in CI" instructions to point users to the `mieubrisse/actions-comment-run@allowed-users-for-orgs` fork, which has the necessary fix for running untrusted PRs with org-owned repos
* Replace list of initializer parameters with a pointer to use the `SHOW_HELP` initializer flag
* Update quickstart prerequisites with a Kurtosis login
* Don't run the validation job on `master` branch (should already have had it before PR is merged)
* Added changelog check to CircleCI
* Added instructions for running in CI
* Refactored entire onboarding flow to make it much, much simpler to quickstart!
* Added docs on the `CLIENT_ID` and `CLIENT_SECRET` Docker environment variables that can be passed to Kurtosis
* Updated docs with information about needing to bind-mount `$HOME/.kurtosis` -> `/kurtosis` for the initializer
* Added an initial version of the docs
