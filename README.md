TODO REPLACE WITH REPO NAME
===========================
TODO replace with repo description

TODO COMPLETE AND DELETE
------------------------
CONFIGURE GENERAL REPO SETTINGS
1. [x] Check "always suggest updating pull request branches"
1. [x] Check "allow auto-merge"
1. [x] Check "automatically delete head branches"

SET UP KURTOSIS BRANCH PROTECTION
1. [x] Under "Branches", create a new branch protection rule
1. [x] Set the rule name to `master`
1. [x] Check "Require pull request reviews before merging"
1. [x] Check "Require approvals" (leaving it at 1)
1. [x] Check "Allow specified actors to bypass required pull requests" and give the `kurtosisbot` user the permission
1. [x] Check "Require status checks to pass before merging"
1. [x] Check "Require branches to be up-to-date before merging" (NOTE: this prevents subtle bugs where two people change code in two separate branches, both their branches pass CI, but when merged they fail)
1. [x] Add the status checks you want to pass (NOTE: if you have no CI/status checks for now, this is fine - just leave it empty)
1. [x] Check "Require conversation resolution before merging" (NOTE: this is important as people sometimes forget comments)
1. [x] Check "Include admins" at the bottom (admins can make mistakes too)
1. [x] Select "Create" at the bottom

SET UP VERSIONING/RELEASING
1. [ ] Ask Kevin to make the Kurtosisbot RELEASER_TOKEN available to this repo so that the repo-releasing Github Action can use it (Kevin eeds to go into `kurtosis-tech` org settings and give the secret access to your new repo)
