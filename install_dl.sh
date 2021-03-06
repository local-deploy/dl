#!/bin/bash

# Script to download and install DL, https://github.com/local-deploy/dl
# Usage: chmod +x ./install_dl.sh && ./install_dl.sh

set -e

GITHUB_REPO=local-deploy/dl
TMPDIR=/tmp
CURRENT_DIR=${PWD}

RED='\033[31m'
GREEN='\033[32m'
YELLOW='\033[33m'
RESET='\033[0m'
OS=$(uname)

if [[ $EUID -eq 0 ]]; then
  echo "This script must NOT be run with sudo/root. Please re-run without sudo" 1>&2
  exit 1
fi

uname_arch=$(uname -m)
if [ "$uname_arch" != "x86_64" ]; then
  printf "${RED}Sorry, your machine architecture %s is not currently supported${RESET}\n" "${uname_arch}" && exit 1
fi

if [[ "$OS" == "Darwin" ]]; then
  BIN="dl_darwin_amd64"
elif [[ "$OS" == "Linux" ]]; then
  BIN="dl_linux_amd64"
else
  printf "${RED}Sorry, this installer does not support your platform at this time${RESET}\n"
  exit 1
fi

if ! docker --version >/dev/null 2>&1; then
  printf "${YELLOW}Docker is required for dl. Please see https://docs.docker.com/engine/install/ ${RESET}\n"
  exit 1
fi

if ! docker-compose --version >/dev/null 2>&1; then
  printf "${YELLOW}docker-compose is required for dl. Please see https://docs.docker.com/compose/install/ ${RESET}\n"
  exit 1
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

LATEST_RELEASE=$(curl --silent "https://api.github.com/repos/$GITHUB_REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

RELEASE_BASE_URL="https://github.com/${GITHUB_REPO}/releases/download/$LATEST_RELEASE"
TARBALL="dl-$LATEST_RELEASE.tar.gz"

printf "${GREEN}Downloading release %s${RESET}\n" "${LATEST_RELEASE}"

curl -fsSL "$RELEASE_BASE_URL/$TARBALL" -o "${TMPDIR}/${TARBALL}" || (printf "${RED}Failed downloading %s/%s${RESET}\n" "${RELEASE_BASE_URL}" "${TARBALL}" && exit 1)

printf "${GREEN}Extract archive${RESET}\n"

cd $TMPDIR
tar -xzf "$TARBALL"

if [ -d "$HOME/.config/dl/config-files" ]; then
  rm -rf "$HOME/.config/dl/config-files"
fi
if [ -f "$HOME/.local/bin/dl" ]; then
  rm -f "$HOME/.local/bin/dl"
fi
if [ -f "$HOME/.config/dl/config.yaml" ]; then
  sed -i "/version/c version: $LATEST_RELEASE" $HOME/.config/dl/config.yaml
fi

if [ ! -d "$HOME/.local/bin" ]; then
  mkdir -p "$HOME/.local/bin"
fi
if [ ! -d "$HOME/.config/dl" ]; then
  mkdir -p "$HOME/.config/dl/config-files"
fi

case ":$PATH:" in
*:$HOME/.local/bin:*) ;;
*) printf "\nexport \"PATH=\$PATH:$HOME/.local/bin\"" >>"$HOME/$SHELL_RC" && PATH="$PATH:$HOME/.local/bin" ;;
esac

mv "bin/$BIN" "$HOME/.local/bin/dl"
mv "config-files" "$HOME/.config/dl/config-files/"

chmod +x "$HOME/.local/bin/dl"

printf "${GREEN}Remove temp files${RESET}\n"

rm -f ${TMPDIR}$TARBALL

#if command -v mkcert >/dev/null; then
#  printf "${YELLOW}Running mkcert -install, which may request your sudo password.'.${RESET}\n"
#  mkcert -install
#fi

printf "${GREEN}DL is now installed. Run \"dl\" and \"dl version\" to verify your installation and see usage.${RESET}\n"

trap 'rm -f ${CURRENT_DIR}/install_dl.sh' EXIT
