#!/usr/bin/env bash
set -euo pipefail   # Bash "strict mode"

GO_VERSION=1.18
NODE_VERSION=16.14.0
RED_BG=$(tput setab 1)
WHITE_FG=$(tput setaf 7)
NORMAL_BG=$(tput sgr0;)
BOLD=$(tput bold)
error=false;

echo  " "

function parse_semver() {
    local token="$1"
    local major=0
    local minor=0
    local patch=0

    if egrep '^[0-9]+\.[0-9]+(\.[0-9]+)?' <<<"${token}" >/dev/null 2>&1 ; then
        # It has the correct syntax.
        local n=${token//[!0-9]/ }
        local a=(${n//\./ })
    fi

    echo "${a[@]}"
}

function compare_versions() {
  local expected=($(parse_semver "$1"))
  local found=($(parse_semver "$2"))
  local len=${#expected[@]}
  for (( i=0; i<len; i++ )) ; do
    if ! [ "${expected[$i]}" == "${found[$i]}" ]; then
      echo 1
    fi
  done
}

check_node_version() {
  if [ -f ~/.nvm/nvm.sh ]; then
    . ~/.nvm/nvm.sh
  elif command -v brew; then
    # https://docs.brew.sh/Manpage#--prefix-formula
    BREW_PREFIX=$(brew --prefix nvm)
    if [ -f "${BREW_PREFIX}/nvm.sh" ]; then
      . "${BREW_PREFIX}"/nvm.sh
    fi
  fi

  if ! command -v nvm &> /dev/null ; then
    echo "WARN: not able to configure nvm"
    exit 1
  fi


  if ! nvm list "${NODE_VERSION}" &> /dev/null; then
    echo "${RED_BG}${WHITE_FG}${BOLD}node "${NODE_VERSION}" not installed. Please install it with ${NORMAL_BG}"
    echo "${RED_BG}${WHITE_FG}nvm install "${NODE_VERSION}"                                ${NORMAL_BG}"
    echo  ""
    error=true
  fi
}

check_go_version() {
  local version=$(go version | { read -r _ _ v _; echo "${v#go}"; })
  local result="$(compare_versions ${GO_VERSION} "${version}")"
  if  [ "$result" == 1 ]; then
    echo "${RED_BG}${WHITE_FG}${BOLD}GO "${GO_VERSION}" not installed. Found ${version}    ${NORMAL_BG}"
    error=true
  fi
}

check_node_version
check_go_version

if "$error"; then
  echo  exiting...
  exit 1
fi

nvm use $NODE_VERSION



