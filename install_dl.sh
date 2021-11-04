#!/bin/bash

# Script to download and install DL, https://github.com/local-deploy/dl
# Usage: install_dl.sh

set -e

GITHUB_REPO=local-deploy/dl
TMPDIR=/tmp

RED='\033[31m'
GREEN='\033[32m'
YELLOW='\033[33m'
RESET='\033[0m'
OS=$(uname)
USER=$(whoami)

if [[ $EUID -eq 0 ]]; then
  echo "This script must NOT be run with sudo/root. Please re-run without sudo." 1>&2
  exit 1
fi

uname_arch=$(uname -m)
if [ "$uname_arch" != "x86_64" ]; then
  printf "${RED}Sorry, your machine architecture %s is not currently supported.${RESET}\n" "${uname_arch}" && exit 1
fi

if [[ "$OS" == "Darwin" ]]; then
  BIN="dl_darwin_amd64"
elif [[ "$OS" == "Linux" ]]; then
  BIN="dl_linux_amd64"
else
  printf "${RED}Sorry, this installer does not support your platform at this time.${RESET}\n"
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

LATEST_RELEASE=$(curl --silent "https://api.github.com/repos/$GITHUB_REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

RELEASE_BASE_URL="https://github.com/${GITHUB_REPO}/releases/download/$LATEST_RELEASE"
TARBALL="dl-$LATEST_RELEASE.tar.gz"

curl -fsSL "$RELEASE_BASE_URL/$TARBALL" -o "${TMPDIR}/${TARBALL}" || (printf "${RED}Failed downloading %s/%s${RESET}\n" "${RELEASE_BASE_URL}" "${TARBALL}" && exit 1)

cd $TMPDIR
tar -xzf $TARBALL

if [ ! -d "$HOME/.local/bin" ]; then
  mkdir -p "$HOME/.local/bin"
fi
if [ ! -d "$HOME/.config/dl" ]; then
  mkdir -p "$HOME/.config/dl"
  mkdir -p "$HOME/.config/config-files"
fi

case ":$PATH:" in
  *:$HOME/.local/bin:*) ;;
  *) printf "\nexport \"PATH=\$PATH:$HOME/.local/bin\"" >>"$HOME/.bashrc" && PATH="$PATH:$HOME/.local/bin" ;;
esac

if [ -d "$HOME/.config/dl" ]; then
  rm -rf "$HOME/.config/dl"
fi
if [ -f "$HOME/.local/bin/dl" ]; then
  rm -f "$HOME/.local/bin/dl"
fi

mv "bin/$BIN" "$HOME/.local/bin/dl"
mv "config-files" "$HOME/.config/dl/"

chmod +x "$HOME/.local/bin/dl"

rm -f ${TMPDIR}$TARBALL

#if command -v brew >/dev/null; then
#  if [ -d "$(brew --prefix)/etc/bash_completion.d" ]; then
#    bash_completion_dir=$(brew --prefix)/etc/bash_completion.d
#    cp dl_bash_completion.sh $bash_completion_dir/dl
#    printf "${GREEN}Installed dl bash completions in $bash_completion_dir${RESET}\n"
#    rm dl_bash_completion.sh
#  else
#    printf "${YELLOW}Bash completion for dl was not installed. You may manually install /tmp/dl_bash_completion.sh in your bash_completion.d directory.${RESET}\n"
#  fi
#
#  if [ -d "$(brew --prefix)/share/zsh-completions" ] && [ -f dl_zsh_completion.sh ]; then
#    zsh_completion_dir=$(brew --prefix)/share/zsh-completions
#    cp dl_zsh_completion.sh $zsh_completion_dir/_dl
#    printf "${GREEN}Installed dl zsh completions in $zsh_completion_dir${RESET}\n"
#    rm dl_zsh_completion.sh
#  else
#    printf "${YELLOW}zsh completion for dl was not installed. You may manually install ${TMPDIR}/dl_zsh_completion.sh in your zsh-completions directory.${RESET}\n"
#  fi
#fi

#if command -v mkcert >/dev/null; then
#  printf "${YELLOW}Running mkcert -install, which may request your sudo password.'.${RESET}\n"
#  mkcert -install
#fi

printf "${GREEN}DL is now installed. Run \"dl\" and \"dl version\" to verify your installation and see usage.${RESET}\n"
