#!/bin/bash
# Script to download and install DL, https://github.com/local-deploy/dl
# Usage: chmod +x ./install_dl.sh && ./install_dl.sh

# shellcheck disable=SC2059

set -e

GITHUB_REPO=local-deploy/dl
TMP_DIR=/tmp
CURRENT_DIR=${PWD}

RED='\033[31m'
GREEN='\033[32m'
YELLOW='\033[33m'
RESET='\033[0m'

if [[ $EUID -eq 0 ]]; then
  echo "This script must NOT be run with sudo/root. Please re-run without sudo" 1>&2
  exit 1
fi

# check docker
if ! docker --version >/dev/null 2>&1; then
  printf "${YELLOW}Docker is required for dl. Please see https://docs.docker.com/engine/install/ ${RESET}\n"
  exit 1
fi

if ! docker-compose --version >/dev/null 2>&1; then
  DOCKER_LEGACY=false
fi
if ! docker compose version >/dev/null 2>&1; then
  DOCKER_PLUGIN=false
fi

# check docker compose
if [[ $DOCKER_LEGACY = false ]] && [[ $DOCKER_PLUGIN = false ]]; then
  printf "${YELLOW}docker compose is required for dl. Please see https://docs.docker.com/compose/install/ ${RESET}\n"
  exit 1
fi

# check docker-compose version
if [[ $DOCKER_PLUGIN = false ]] && [[ $DOCKER_LEGACY != false ]]; then
  DOCKER_COMPOSE_LEGACY_MAJOR=$(docker-compose --version --short | cut -d'.' -f 1)
  if [[ "${DOCKER_COMPOSE_LEGACY_MAJOR}" -eq 1 ]]; then
    printf "${YELLOW}docker compose is required version 2. Please update https://docs.docker.com/compose/install/ ${RESET}\n"
    exit 1
  fi
fi

# check architecture
architecture=""
case $(uname -m) in
x86_64 | amd64) architecture="amd64" ;;
arm64 | aarch64 | armv8b | armv8l | aarch64_be) architecture="arm64" ;;
esac

if [[ $architecture == "" ]]; then
  printf "${RED}Sorry, your machine architecture %s is not currently supported${RESET}\n" "$(uname -m)" && exit 1
fi

# check OS
os=""
case $(uname) in
Linux) os="linux" ;;
Darwin) os="darwin" ;;
esac

if [[ $os == "" ]]; then
  printf "${RED}Sorry, this installer does not support %s platform at this time${RESET}\n" "$(uname)" && exit 1
fi

case $SHELL in
*/zsh)
  SHELL_RC=".zshrc"
  ;;
*/bash)
  SHELL_RC=".bashrc"
  ;;
*)
  printf "${RED}Sorry, your shell is not currently supported${RESET}\n"
  exit 1
  ;;
esac

if [ "$1" ]; then
  LATEST_RELEASE=$1
else
  LATEST_RELEASE=$(curl --silent "https://api.github.com/repos/$GITHUB_REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
fi

RELEASE_BASE_URL="https://github.com/${GITHUB_REPO}/releases/download/$LATEST_RELEASE"
TARBALL="dl-$LATEST_RELEASE-${os}-${architecture}.tar.gz"

printf "${GREEN}Downloading release %s${RESET}\n" "${LATEST_RELEASE}"

curl -fsSL "$RELEASE_BASE_URL/$TARBALL" -o "${TMP_DIR}/${TARBALL}" || (printf "${RED}Failed downloading %s/%s${RESET}\n" "${RELEASE_BASE_URL}" "${TARBALL}" && exit 1)

printf "${GREEN}Extract archive${RESET}\n"

cd $TMP_DIR
tar -xzf "$TARBALL"

if [ -d "$HOME/.config/dl" ]; then
  rm -rf "$HOME/.config/dl"
fi
if [ -f "$HOME/.local/bin/dl" ]; then
  rm -f "$HOME/.local/bin/dl"
fi

if [ ! -d "$HOME/.local/bin" ]; then
  mkdir -p "$HOME/.local/bin"
fi

if [[ ":$PATH:" != *":$HOME/.local/bin:"* ]]; then
  printf "\nPATH=\"\$HOME/.local/bin:\$PATH\"" >>"$HOME/$SHELL_RC" && PATH="$PATH:$HOME/.local/bin"
fi

mv dl "$HOME/.local/bin/dl"
chmod +x "$HOME/.local/bin/dl"

printf "${GREEN}Remove temp files${RESET}\n"

rm -f "${TMP_DIR}$TARBALL"

printf "${GREEN}DL is now installed. Run \"dl\" and \"dl version\" to verify your installation and see usage.${RESET}\n"

trap 'rm -f ${CURRENT_DIR}/install_dl.sh' EXIT
