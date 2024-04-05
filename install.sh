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

	echo "Downloading Zig version: ${version}"
	wget -P /opt/zig/ "https://ziglang.org/builds/zig-linux-x86_64-${version}.tar.xz"

	if [[ -f "/opt/zig/zig-linux-x86_64-${version}.tar.xz" ]]; then
		tar -xvf "/opt/zig/zig-linux-x86_64-${version}.tar.xz"
		rm "/opt/zig/zig-linux-x86_64-${version}.tar.xz"
	else
		echo "Zig download failed."
		exit 1
	fi
}

create_directories() {
	if [[ ! -d /opt/zig/bin ]]; then
		echo "Creating /opt/zig/bin directory."
		mkdir -p /opt/zig/bin
	fi
}

cleanup_old_installations() {
	if [[ -f /opt/zig/bin/zig ]]; then
		echo "Removing old Zig version $(zig version)."
		rm /opt/zig/bin/zig
	fi
}

install_version() {
	version=$1

	echo "Installing Zig version: ${version}"
	ln -s "/opt/zig/zig-linux-x86_64-${version}/zig" /opt/zig/bin/zig

	if [[ -f /opt/zig/bin/zig ]]; then
		echo "Zig $(zig version) installed successfully."
	else
		echo "Zig installation failed."
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
	create_directories
	cleanup_old_installations
	install_version "${version}"
}

main "$@"
