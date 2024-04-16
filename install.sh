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
		rm /usr/local/bin/zig
	fi
}

install_version() {
	version=$1

	echo "Installing Zig version: ${version}"
	ln -s "/opt/zig/zig-linux-x86_64-${version}/zig" /usr/local/bin/zig

	if [[ -f /usr/local/bin/zig ]]; then
		echo "Zig $(zig version) installed successfully."
		exit 0
	else
		echo "Zig installation failed."
		exit 1
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
}

main "$@"
