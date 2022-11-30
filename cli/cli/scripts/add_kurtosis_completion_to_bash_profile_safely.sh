#!/bin/sh

VERSION=$1
ADD_COMPLETION_SCRIPT="add_kurtosis_completion_to_bash_profile_safely.sh"
BREW_PREFIX=$(brew --prefix)
NL='
'

bashProfile=$(cat ~/.bash_profile | sed '/bash_completion.d\/kurtosis/d')
echo "${bashProfile}${NL}${NL}# Leave this incantation as a single line, so that homebrew upgrades are smooth${NL}if [ -f ${BREW_PREFIX}/etc/bash_completion.d/kurtosis ]; then source ${BREW_PREFIX}/etc/bash_completion.d/kurtosis; fi${NL}" > ~/.bash_profile

echo "*** Did the above fail? ***"
echo "If yes, you will need to do this after homebrew finishes (once off):"
echo "${BREW_PREFIX}/Cellar/kurtosis-cli/${VERSION}/bin/${ADD_COMPLETION_SCRIPT}"