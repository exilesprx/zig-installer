#!/bin/bash

check_version() {
	version=$1

	if [[ "${version}" == "$(zig version)" ]]; then
		echo "Zig ${version} is already installed."
		exit 0
	fi
}

download_version() {
	version=$1

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

download_zls() {

	if [[ -d /opt/zig/zls ]]; then
		echo "ZLS already exists. Fetching latest."
		cwd=$(pwd)
		cd /opt/zls || exit 1
		git pull
		cd "$cwd" || exit 1
	else
		echo "Downloading ZLS."
		sudo mkdir -p /opt/zls
		sudo chown -R "$(whoami)":"$(whoami)" /opt/zls
		git clone https://github.com/zigtools/zls.git /opt/zls
	fi
}

build_zls() {
	echo "Building ZLS."
	cwd=$(pwd)
	cd /opt/zls || exit 1
	zig build -Doptimize=ReleaseSafe
	cd "$cwd" || exit 1
}

install_zls() {
	if [[ ! -f /usr/local/bin/zls ]]; then
		echo "Installing ZLS."
		sudo ln -s /opt/zls/zig-out/bin/zls /usr/local/bin/zls
	fi
}

main() {
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
	download_zls
	build_zls
	install_zls

	exit 0
}

main "$@"
