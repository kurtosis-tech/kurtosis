---
title: Metrics Philosophy
sidebar_label: Metrics Philosophy
---

Kurtosis is a small startup, which means that understanding how users are using the product is vital to our success. To this end, we've built the capability for Kurtosis to send product analytics metrics.

However, user metrics are abused heavily in today's world - data is collected without the ability to disable analytics, intentionally deanonymized, and sold to third parties. We hate it as much as we're guessing you do.

It was therefore important to us to collect our product analytic metrics ethically. Concretely, this means that we've made our metrics:

1. Private: we will **never** give or sell your data to third parties
1. Anonymized: your user ID is a hash, so we don't know who you are
1. Obfuscated: potentially-sensitive parameters (e.g. enclave IDs) are hashed as well
1. Opt-out: Kurtosis allows you to [easily switch off analytics](../cli-reference/analytics-disable.md), even [in CI](../guides/running-in-ci.md)

If that sounds fair to you, we'd really appreciate you helping us get the data to make our product better. In exchange, you have our word that we'll honor the trust you've placed in us by continuing to fulfill the metrics promises above.
