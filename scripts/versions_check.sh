#!/usr/bin/env bash
set -euo pipefail   # Bash "strict mode"

# VERSION NUMBERS
# FOR GO WE EXPECT _AT LEAST_ THIS VERSION, BUT WE ARE OK WITH SUPERIOR VERSIONS
GO_VERSION=1.18

# FOR NODE, WE PIN THE EXACT VERSION NUMBER
NODE_VERSION=16.14


RED_BG=$(tput setab 1)
BLUE_BG=$(tput setab 4)
WHITE_FG=$(tput setaf 7)
NORMAL_BG=$(tput sgr0;)
BOLD=$(tput bold)
error=false;

echo "${BLUE_BG}${WHITE_FG}${BOLD}Starting Kurtosis Build...              ${NORMAL_BG}"

check_node_version() {
  if [ -f ~/.nvm/nvm.sh ]; then
    source ~/.nvm/nvm.sh
  elif command -v brew; then
    # https://docs.brew.sh/Manpage#--prefix-formula
    BREW_PREFIX=$(brew --prefix nvm)
    if [ -f "${BREW_PREFIX}/nvm.sh" ]; then
      source "${BREW_PREFIX}"/nvm.sh
    fi
  fi

  if ! command -v nvm &> /dev/null ; then
    echo "ERROR: unable to configure nvm"
    exit 1
  fi


  if ! nvm list "${NODE_VERSION}" &> /dev/null; then
    echo "${RED_BG}${WHITE_FG}${BOLD}node "${NODE_VERSION}" not installed. Please install it with ${NORMAL_BG}"
    echo "${RED_BG}${WHITE_FG}nvm install "${NODE_VERSION}"                                ${NORMAL_BG}"
    echo  ""
    error=true
  else
    echo "${BLUE_BG}${WHITE_FG}${BOLD}Node version "${NODE_VERSION}" found. ok    ${NORMAL_BG}"
  fi
}

version_lte() {
  if [ "$1" = "$2" ]; then
    echo 1
    return;
  fi
  # No need to reinvent the wheel. lets use sort with --check (C) and --version-sort options
  # those options are present since coreutils 7, released around 2009.
  $(printf '%s\n%s' "$1" "$2" | sort -C -V -u)  && echo 1 || echo 0
}

check_go_version() {
  local version=$(go version | { read -r _ _ v _; echo "${v#go}"; })
  if  [ "$(version_lte "${GO_VERSION}" "${version}")" !=  1 ]; then
    echo "${RED_BG}${WHITE_FG}${BOLD}GO "${GO_VERSION}" not installed. Found ${version}    ${NORMAL_BG}"
    error=true
  else
    echo "${BLUE_BG}${WHITE_FG}${BOLD}Minimum GO version "${GO_VERSION}" expected. Found ${version} ... ok    ${NORMAL_BG}"

  fi
}


# both check functions just a set a flag. In the unlikely probability there's a error running this script on some exotic
# environment like some really old version of cygwin or some other obscure old unix system, the build will proceed anyway
check_node_version
check_go_version

if "$error"; then
  echo exiting...
  exit 1
fi

nvm use $NODE_VERSION &> /dev/null



