Running in CI
=============
Running Kurtosis on your local machine is nice, but the real power of the platform comes when it's executed as part of CI. This guide will walk you through setting up Kurtosis in your CI environment.

Machine-to-Machine Auth
-----------------------
While it's fine to prompt a user for their username and password when Kurtosis is run on the local machine, this method is insecure when executing Kurtosis on CI. Fortunately, our auth system provides [a system for handling this called "machine-to-machine auth"](https://auth0.com/docs/flows/client-credentials-flow). In the machine-to-machine flow, a client ID and a client secret are stored within the CI environment's secrets and passed in to every CI job. Kurtosis uses these credentials to retrieve a token from the auth provider, which allows Kurtosis execution to proceed.

The client ID and secret are created by the Kurtosis team on the backend, so if you don't have them already then shoot us a ping at [inquiries@kurtosistech.com](mailto:inquiries@kurtosistech.com) to discuss pricing!

**WARNING: Make sure you store your client ID and secret in a secure place! Anyone with the credentials could impersonate you to Kurtosis, which would use up any usage-based credits on your behalf.**

Using Client Credentials
------------------------
Now that you have your client credentials, you'll need to pass them in to your CI environment as environment variables. The route for doing so will be particular to your CI server, so do a Google search for "define YOUR_CI_TOOL secrets". Once done, you'll need to pass the appropriate environment variables to Kurtosis. The Kurtosis initializer takes in special flags for receiving these, so we recommend:

1. [Visiting the Kurtosis documentation on running a testsuite](./testsuite-details.md#running-a-testsuite) to see what Kurtosis initializer parameters to use and how to use them
2. Writing a wrapper script around `build_and_run.sh` that your CI will call to pass in the client ID and secret to Kurtosis

Client Credentials & Untrusted PRs
----------------------------------
Because the client credentials are secrets, they cannot be made available to builds of untrusted PRs (i.e. PRs from untrusted contributors, often submitted from forks) else there's a risk that the untrusted PR maliciously prints them. This means that CI builds for untrusted PRs will fail. This is a tough problem for the entire the open-source community, but a decent workaround exists for repos on Github:

1. Add the `actions-comment-run` Github Action to your repo [with these instructions](https://github.com/mieubrisse/actions-comment-run/tree/allowed-users-for-orgs#introduce-this-action). **WARNING:** If your repo is inside an organization, you should add the usernames of your repo administrators to the `allowed-users` config key. This is due to a bug in Github where "owners" of repos inside an organization don't actually get the Github `OWNER` role (see [this ticket](https://github.community/t/github-actions-have-me-as-contributor-role-when-im-owner/138933/9) for details).
1. Copy [the codeblock inside the "PR merge preview" header](https://github.com/mieubrisse/actions-comment-run/tree/allowed-users-for-orgs#pr-merge-preview)
1. [Store it as one of your "Saved Replies" in Github](https://github.com/mieubrisse/actions-comment-run/tree/allowed-users-for-orgs#tips-saved-replies)

Then, whenever an untrusted PR is submitted to your repo, a repo owner should:

1. Review the code carefully to make sure no secrets are being printed
1. Post a comment with the saved reply from above, which will trigger the Github Actions bot to create a copy of the changes on a new branch with a name like `actions-merge-preview/ORIGINAL-BRANCH-NAME`
1. Open a PR with this `actions-merge-preview/...` branch; because the repo owner is trusted, secrets will be passed to CI and the CI build will proceed
1. Inform the third-party contributor about the new PR so they can see the build status and make any necessary changes
1. If the third-party contributor pushes any more changes, re-review them for any secret-leaking and re-post the same saved reply to update the trusted PR

This allows you to require untrusted code to pass a review before it runs (thereby securing your secrets) while still preserving the safety of CI.
