#!/usr/bin/env bash

REPO_OWNER="jgfranco17"
REPO_NAME="smokesweep"
DEFAULT_VERSION="latest"

get_latest_version() {
  curl --silent "https://api.github.com/repos/${REPO_OWNER}/${REPO_NAME}/releases/${DEFAULT_VERSION}" | \
    grep '"tag_name":' | \
    sed -E 's/.*"([^"]+)".*/\1/'
}

download_binary() {
  local version=$1
  local os=$2
  local arch=$3

  url="https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/${version}/smokesweep-${version}-${os}-${arch}.tar.gz"
  echo "Downloading SmokeSweep from $url"

  curl -L "$url" -o smokesweep.tar.gz || {
    echo "Error: Download failed. Please check the version and try again."
    exit 1
  }
}

install_binary() {
  sudo tar -xzf smokesweep.tar.gz -C /usr/local/bin smokesweep || {
    echo "Error: Installation failed."
    exit 1
  }
  chmod +x /usr/local/bin/smokesweep
  rm smokesweep.tar.gz
}

# =============== MAIN SCRIPT ===============

version="${1:-$DEFAULT_VERSION}"

# Detect OS and architecture
case "$(uname -s)" in
  Linux*) os="linux" ;;
  Darwin*) os="darwin" ;;
  *) echo "Error: Unsupported OS"; exit 1 ;;
esac

arch="$(uname -m)"
case "$arch" in
  x86_64) arch="amd64" ;;
  aarch64) arch="arm64" ;;
  *) echo "Error: Unsupported architecture"; exit 1 ;;
esac

# Resolve latest version if needed
if [ "$version" = "latest" ]; then
  version=$(get_latest_version)
  if [ -z "$version" ]; then
    echo "Error: Unable to fetch the latest version."
    exit 1
  fi
fi

echo "Installing SmokeSweep version $version for $os/$arch"

# Download and install
download_binary "$version" "$os" "$arch"
install_binary

echo "SmokeSweep installation complete!"
echo "You can now run 'smokesweep --version' to verify the installation."
