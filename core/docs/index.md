![](./images/horizontal-logo.jpg)

[Kurtosis](https://www.kurtosistech.com) is a platform for running whole-system tests against distributed systems with the frequency and repeatability of unit tests.

The world is moving to microservices and systems are becoming increasingly complex. The more components a system has, the more [emergent phenomena](https://en.wikipedia.org/wiki/Emergence) and [unexpected outlier events](https://en.wikipedia.org/wiki/Black_swan_theory) that occur. More components equal more difficulty running a representative system, and more corners cut when testing. If nothing is done, testing will continue to decline and unpredictability will continue to rise. Engineers need a new tool to tame the complexity of the distributed age, that marries the ease and safety of unit tests with the representativeness of testing in production. Kurtosis is that tool.

Getting Started
---------------
You have two paths you can start with:

* For those who like to jump in and see things running, head over to [the quickstart instructions][1]
* For those who prefer to start at a high level, check out [the Kurtosis architecture docs][2]

For Q&A, head over to the [Kurtosis Discord](https://discord.gg/6Jjp9c89z9) server.

Documentation Index
------------------------

* [Quickstart][1]
* [Building & Running](./building-and-running.md)
* [Testsuite Customization](./testsuite-customization.md)
* [Debugging common failure scenarios](./debugging-failed-tests.md)
* [Architecture][2]
* [Advanced Usage](./advanced-usage.md)
* [Running Kurtosis in CI](./running-in-ci.md)
* [Versioning & upgrading](./versioning-and-upgrading.md)
* [Changelog](./changelog.md) 

Local Development Tips
----------------------
### Docker Volumes
Kurtosis will create several Docker volumes during the course of normal operation. These are intentionally not deleted after testsuite execution finishes, so that you can explore the volumes for debugging information. This means Docker volumes will slowly accumulate on your system, and you'll want to periodically clear them out. To do so, you can use `docker volume ls` to view existing volumes and identify the ones you'd like to remove, then `docker volume rm THE_VOLUME_ID` to remove them.

### Docker Containers
Much like Docker volumes, Kurtosis will create Docker containers on your system but won't delete them so it doesn't destroy debugging information. This means you'll also need to clear these out periodically. You can use `docker container ls -a` to view all containers (including stopped ones), and `docker container rm THE_CONTAINER_ID` to remove the ones you'd prefer to get rid of.

### Stop all running containers
A handy tip to stop all running containers of a certain image (replace `YOUR-IMAGE-NAME` with the name of the image of the containers you want to remove):

```
docker container ls    # See which Docker containers are left around - these will depend on the containers spun up
docker stop $(docker ps -a --quiet --filter ancestor="YOUR-IMAGE-NAME" --format="{{.ID}}")
```

[1]: https://github.com/kurtosis-tech/kurtosis-libs/tree/master#testsuite-quickstart
[2]: ./architecture.md
