#!/usr/bin/env bash
set -euo pipefail   # Bash "strict mode"

# VERSION NUMBERS
# FOR GO WE EXPECT _AT LEAST_ THIS VERSION, BUT WE ARE OK WITH SUPERIOR VERSIONS
REQUIRED_GO_VERSION=1.20

# FOR NODE, WE PIN THE EXACT VERSION NUMBER
REQUIRED_NODE_VERSION=20.11


RED_BG=$(tput setab 1)
BLUE_BG=$(tput setab 4)
WHITE_FG=$(tput setaf 7)
NORMAL_BG=$(tput sgr0;)
BOLD=$(tput bold)
error=false;

echo "${BLUE_BG}${WHITE_FG}${BOLD}Starting Kurtosis Build...${NORMAL_BG}"

check_node_version() {
  if ! command -v node &> /dev/null ; then
    echo "ERROR: node is not installed or not found in PATH"
    exit 1
  fi

  local local_node_version=$(node --version)
  # stripped_local_node_version should only contain node's {major.minor} versions to compare it with REQUIRED_NODE_VERSION.
  local stripped_local_node_version=$(echo "$local_node_version" | cut -d 'v' -f 2 | awk -F '.' '{print $1"."$2}')
  if [ "$(version_lte "${REQUIRED_NODE_VERSION}" "${stripped_local_node_version}")" != 1 ]; then
    echo "${RED_BG}${WHITE_FG}${BOLD}node "${REQUIRED_NODE_VERSION}" or higher not installed. Found ${stripped_local_node_version}${NORMAL_BG}"  
    exit 1
  else
    echo "${BLUE_BG}${WHITE_FG}${BOLD}Minimum node version "${REQUIRED_NODE_VERSION}" expected. Found ${stripped_local_node_version} ... ok${NORMAL_BG}"
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
  local local_go_version=$(go version | { read -r _ _ v _; echo "${v#go}"; })
  if  [ "$(version_lte "${REQUIRED_GO_VERSION}" "${local_go_version}")" !=  1 ]; then
    echo "${RED_BG}${WHITE_FG}${BOLD}GO "${REQUIRED_GO_VERSION}" not installed. Found ${local_go_version}${NORMAL_BG}"
    error=true
  else
    echo "${BLUE_BG}${WHITE_FG}${BOLD}Minimum GO version "${REQUIRED_GO_VERSION}" expected. Found ${local_go_version} ... ok${NORMAL_BG}"

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
