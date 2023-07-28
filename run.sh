#!/bin/bash

kurtosis clean -a
kurtosis engine stop

kurtosis run --cli-log-level=debug main.star