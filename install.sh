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

check_user() {
  if [ "$(id -u)" -eq 0 ]; then
    echo "Please run this script as a non-root user."
    exit 1
  fi
}

check_dependencies() {
  local pkgs=("git" "wget" "jq" "minisign")
  for pkg in "${pkgs[@]}"; do
    if ! which "${pkg}" >/dev/null 2>&1; then
      echo "Missing $pkg. Please install it."
    fi
  done
}

zig_install() {
  version=$(wget -qO- https://ziglang.org/download/index.json | jq -r '.master.version')
  tarfile="zig-linux-x86_64-${version}.tar.xz"

  if [[ -z "${version}" ]]; then
    echo "Could not determine latest Zig version."
    exit 1
  else
    echo "Found latest Zig version: ${version}"
  fi

  check_version
  download_version
  verify_signature
  install_version
}

check_version() {
  if [[ "${version}" == "$(zig version)" ]]; then
    echo "Zig ${version} is already installed."
    exit 0
  fi
}

download_version() {
  if [[ ! -d /opt/zig ]]; then
    sudo mkdir -p /opt/zig
    sudo chown -R "$(whoami)":"$(whoami)" /opt/zig
  fi

  if [[ -f "/opt/zig/${tarfile}" ]]; then
    echo "Zig version ${version} already downloaded."
  elif wget -q --spider "https://ziglang.org/builds/${tarfile}"; then
    echo "Downloading Zig version: ${version}"
    wget -P /opt/zig/ "https://ziglang.org/builds/${tarfile}"
  else
    echo "Zig version ${version} not found."
    exit 1
  fi
}

verify_signature() {
  local output
  local pubkey="RWSGOq2NVecA2UPNdBUZykf1CCb147pkmdtYxgb3Ti+JO/wCYvhbAb/U"
  local search_string="Signature and comment signature verified"

  if [[ -f "/opt/zig/${tarfile}.minisig" ]]; then
    echo "Zig signature already downloaded."
  elif wget -q --spider "https://ziglang.org/builds/${tarfile}.minisig"; then
    echo "Downloading Zig signature"
    wget -P /opt/zig/ "https://ziglang.org/builds/${tarfile}.minisig"
  else
    echo "Zig signature not found. Cannot verify tar file."
    exit 1
  fi

  output=$(minisign -Vm "/opt/zig/${tarfile}" -P "${pubkey}")
  if [[ "$output" == *"$search_string"* ]]; then
    echo "Zig tar verified."
    rm "/opt/zig/${tarfile}.minisig"
  else
    echo "Zig tar verification failed. Removing files signature and tarfile."
    rm "/opt/zig/${tarfile}" "/opt/zig/${tarfile}.minisig"
    exit 1
  fi
}

install_version() {
  echo "Installing Zig version: ${version}"
  tar -xf "/opt/zig/${tarfile}" -C "/opt/zig/"
  rm "/opt/zig/${tarfile}"
  sudo ln -sf "/opt/zig/zig-linux-x86_64-${version}/zig" /usr/local/bin/zig

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
  echo "!! Sudo password may be required !!"

  check_user
  check_dependencies

  local cwd
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
