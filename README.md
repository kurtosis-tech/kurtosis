TODO REPLACE WITH REPO NAME
===========================
TODO replace with repo description

TODO COMPLETE AND DELETE
------------------------
CONFIGURE GENERAL REPO SETTINGS
1. [ ] Check "always suggest updating pull request branches"
1. [ ] Check "allow auto-merge"
1. [ ] Check "automatically delete head branches"

SET UP KURTOSIS BRANCH PROTECTION
1. [ ] Under "Branches", create a new branch protection rule
1. [ ] Set the rule name to `master`
1. [ ] Check "Require pull request reviews before merging"
1. [ ] Check "Require approvals" (leaving it at 1)
1. [ ] Check "Allow specified actors to bypass required pull requests" and give the `kurtosisbot` user the permission
1. [ ] Check "Require status checks to pass before merging"
1. [ ] Check "Require branches to be up-to-date before merging" (NOTE: this prevents subtle bugs where two people change code in two separate branches, both their branches pass CI, but when merged they fail)
1. [ ] Add the status checks you want to pass (NOTE: if you have no CI/status checks for now, this is fine - just leave it empty)
1. [ ] Check "Require conversation resolution before merging" (NOTE: this is important as people sometimes forget comments)
1. [ ] Check "Include admins" at the bottom (admins can make mistakes too)
1. [ ] Select "Create" at the bottom

SET UP VERSIONING/RELEASING
1. [ ] Ask Kevin to make the Kurtosisbot RELEASER_TOKEN available to this repo so that the repo-releasing Github Action can use it (Kevin eeds to go into `kurtosis-tech` org settings and give the secret access to your new repo)
