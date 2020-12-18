![](./images/horizontal-logo.jpg)

[Kurtosis](https://www.kurtosistech.com) is a platform for running whole-system tests against distributed systems with the frequency and repeatability of unit tests.

The world is moving to microservices and systems are becoming increasingly complex. The more components a system has, the more [emergent phenomena](https://en.wikipedia.org/wiki/Emergence) and [unexpected outlier events](https://en.wikipedia.org/wiki/Black_swan_theory) that occur. More components equal more difficulty running a representative system, and more corners cut when testing. If nothing is done, testing will continue to decline and unpredictability will continue to rise. Engineers need a new tool to tame the complexity of the distributed age, that marries the ease and safety of unit tests with the representativeness of testing in production. Kurtosis is that tool.

Getting Started
---------------
You have two paths you can start with:

* For those who like to jump in and see things running, head over to [the quickstart instructions](./quickstart.md)
* For those who prefer to start at a high level, check out [the Kurtosis architecture docs](./architecture.md)

For Q&A, head over to the [Kurtosis Discord](https://discord.gg/6Jjp9c89z9) server.

Additional Documentation
------------------------

* [Debugging common failure scenarios](./debugging-failed-tests.md)
* [Running Kurtosis in CI](./running-in-ci.md)

Developer Notes
---------------
Stop running containers (replace `YOUR-IMAGE-NAME` with the name of the image of the containers you want to remove):
```
docker container ls    # See which Docker containers are left around - these will depend on the containers spun up
docker stop $(docker ps -a --quiet --filter ancestor="YOUR-IMAGE-NAME" --format="{{.ID}}")
```
