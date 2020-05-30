# kurtosis
E2E Testing Harness for Ava

# Requirements

Golang version 1.13.x.   
Docker Engine running in your environment.

# Install

Clone this repository and cd into it.  
Run `./scripts/build.sh`. This will build the main binary and put it the `build/` directory of this repository.  

Clone [the test controller](https://github.com/kurtosis-tech/ava-test-controller) and run `docker build .` inside the directory.

# Usage

Run `./build/kurtosis -help` or `./build/kurtosis -h` to see command line usage.  
Example: `kurtosis --gecko-image-name=gecko-f290f73 --test-controller-image-name=YOURIMAGE`

# Architecture

Kurtosis builds a network of Gecko Docker images and runs a Docker container to run tests against it.
Both images must already exist in your Docker engine, and the names of both images are specified by a command line argument.  
Currently, the ports that the container will run on for HTTP and for staking on your host machine are hard-coded to the standard Gecko defaults - 9650 for HTTP, 9651 for staking.

# Helpful Tip

Create an alias in your shell .rc file to stop and clear all Docker containers created by Kurtosis in one line.  
Run this every time after you kill kurtosis, because the containers will hang around.  
One way to do this is as follows:

```
# alias for clearing kurtosis containers 
kurtosisclearall() {  docker rm $(docker stop $(docker ps -a -q --filter ancestor="$1" --format="{{.ID}}")) } 
alias kclear=kurtosisclearall
```

Usage:
```
export GECKO_IMAGE=gecko-684ca4e
# run kurtosis
./build/kurtosis -gecko-image-name="${GECKO_IMAGE}"
# ...kill kurtosis manually...
# clear the docker containers initialized by kurtosis
kclear ${GECKO_IMAGE} 
```

# TODO

* Ability to run spectators versus stakers
