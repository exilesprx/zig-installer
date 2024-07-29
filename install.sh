#!/bin/bash

# Help function to display usage information
help() {
  echo "Usage: $0 [OPTIONS]"
  echo "Install Zig and ZLS (Zig Language Server) or only one of them."
  echo
  echo "Options:"
  echo "  --zig-only      Install only Zig"
  echo "  --zls-only      Install only ZLS (Zig Language Server)"
  echo "  -h, --help      Display this help message and exit"
  exit 0
}

zig_install() {
  version=$(wget -qO- https://ziglang.org/download/index.json | jq -r '.master.version')

  if [[ -z "${version}" ]]; then
    echo "Could not determine latest Zig version."
    exit 1
  else
    echo "Found latest Zig version: ${version}"
  fi

  check_version "${version}"
  download_version "${version}"
  cleanup_old_installations
  install_version "${version}"
}

check_version() {
  version=$1

  if [[ "${version}" == "$(zig version)" ]]; then
    echo "Zig ${version} is already installed."
    exit 0
  fi
}

download_version() {
  version=$1
  pub_key="RWSGOq2NVecA2UPNdBUZykf1CCb147pkmdtYxgb3Ti+JO/wCYvhbAb/U"
  search_string="Signature and comment signature verified"

  if [[ ! -d /opt/zig ]]; then
    sudo mkdir -p /opt/zig
    sudo chown -R "$(whoami)":"$(whoami)" /opt/zig
  fi

  if wget -q --spider "https://ziglang.org/builds/zig-linux-x86_64-${version}.tar.xz"; then
    echo "Downloading Zig version: ${version}"
    wget -P /opt/zig/ "https://ziglang.org/builds/zig-linux-x86_64-${version}.tar.xz"
  else
    echo "Zig version ${version} not found."
    exit 1
  fi

  shasum=$(wget -qO- https://ziglang.org/download/index.json | jq -r '.master.src.shasum')
  wget -P /opt/zig/ "https://ziglang.org/builds/zig-linux-x86_64-${version}.tar.xz.minisig"

  if [[ -z "${shasum}" ]]; then
    echo "Could not determine SHA-256 checksum for Zig version ${version}."
    rm "/opt/zig/zig-linux-x86_64-${version}.tar.xz" "/opt/zig/zig-linux-x86_64-${version}.tar.xz.minisig"
    exit 1
  fi

  cd /opt/zig || exit 1
  result=$(minisign -Vm "zig-linux-x86_64-${version}.tar.xz" -P "${pub_key}")

  if [[ "$result" == *"$search_string"* ]]; then
    echo "Zig download verified."
  else
    echo "Zig download verification failed."
    rm "/opt/zig/zig-linux-x86_64-${version}.tar.xz" "/opt/zig/zig-linux-x86_64-${version}.tar.xz.minisig"
    exit 1
  fi

  if [[ -f "/opt/zig/zig-linux-x86_64-${version}.tar.xz" ]]; then
    tar -xf "/opt/zig/zig-linux-x86_64-${version}.tar.xz" -C "/opt/zig/"
    rm "/opt/zig/zig-linux-x86_64-${version}.tar.xz"
  else
    echo "Zig download failed."
    exit 1
  fi
}

cleanup_old_installations() {
  if [[ -f /usr/local/bin/zig ]]; then
    echo "Removing old Zig version $(zig version)."
    sudo rm /usr/local/bin/zig
  fi
}

install_version() {
  version=$1

  echo "Installing Zig version: ${version}"
  sudo ln -s "/opt/zig/zig-linux-x86_64-${version}/zig" /usr/local/bin/zig

  if [[ -f /usr/local/bin/zig ]]; then
    echo "Zig $(zig version) installed successfully."
  else
    echo "Zig installation failed."
    exit 1
  fi
}

zls_install() {
  fetch_zls
  build_zls
  install_zls
}

fetch_zls() {

  if [[ -d /opt/zls ]]; then
    cd /opt/zls || exit 1
    git fetch
    if [[ $(git rev-list HEAD...origin/master --count) -gt 0 ]]; then
      echo "Fetching latest"
      git pull
    fi
  else
    echo "Fetching ZLS."
    sudo mkdir -p /opt/zls
    sudo chown -R "$(whoami)":"$(whoami)" /opt/zls
    git clone https://github.com/zigtools/zls.git /opt/zls
  fi
}

build_zls() {
  echo "Building ZLS."
  cd /opt/zls || exit 1
  zig build -Doptimize=ReleaseSafe
}

install_zls() {
  if [[ ! -f /usr/local/bin/zls ]]; then
    echo "Installing ZLS."
    sudo ln -s /opt/zls/zig-out/bin/zls /usr/local/bin/zls
  fi
}

main() {
  cwd=$(pwd)
  if [[ "$#" -eq 0 ]]; then
    zig_install
    zls_install
  elif [[ "$1" == "--zig-only" ]]; then
    zig_install
  elif [[ "$1" == "--zls-only" ]]; then
    zls_install
  elif [[ "$1" == "-h" || "$1" == "--help" ]]; then
    help
  else
    echo "Invalid option: $1"
    help
  fi
  cd "$cwd" || exit 1
  echo "Done!"
  exit 0
}

main "$@"
