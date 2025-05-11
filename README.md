# Zig and ZLS Installer

This program installs Zig and the Zig Language Server (ZLS). You can choose to install both or just one of them.

**Note:** This program must be run as root. Use `sudo` when executing the binary.

> **Platform Support:** Linux is fully supported. macOS builds are currently a work in progress.

## Prerequisites

Before running the program, ensure you have the following dependencies installed:

- `wget` (for downloading Zig binary)
- `git` (for downloading ZLS)
- `jq` (for parsing version information)
- `minisign` (for verifying Zig download)
- `xz` (for extracting archives)

## Installation

1. Clone the repository or download the pre-built binary for your platform.
2. If downloading the source code, build the program:

```bash
go build
```

Or use the justfile if available:

```bash
just build
```

3. Run the program with the desired options.

## Usage

```bash
sudo ./zig-install-{platform} [command] [OPTIONS]
```

### Commands

- `install`: Install Zig and/or ZLS (default if no command specified)
- `version`: Show version information about the installer
- `env`: Generate a template .env file

### Options

- `--zig-only`: Install only Zig
- `--zls-only`: Install only ZLS (Zig Language Server)
- `--verbose`: Show detailed output during installation
- `--no-color`: Disable colored output
- `--env <file>`: Specify a custom environment file (default: `.env`)
- `--settings`: Show current settings
- `--log-file <file>`: Specify log file (default: `zig-install.log`)
- `--enable-log`: Enable/disable logging to file
- `--version`: The version to install for Zig and ZLS

## Configuration

This program can be configured in two ways (in order of precedence):

1. **Command-line flags**: Options provided directly when running the program
2. **Configuration file**: Settings in an `.env` file

### Configuration File (.env)

You can create a `.env` file in the same directory as the executable or use the `env` command to create a template:

```
# Zig download and verification
ZIG_PUB_KEY=RWSGOq2NVecA2UPNdBUZykf1CCb147pkmdtYxgb3Ti+JO/wCYvhbAb/U
ZIG_DOWN_URL=https://ziglang.org/builds/
ZIG_INDEX_URL=https://ziglang.org/download/index.json
```

The values override the defaults.

Creating a `.env` file is optional, but it allows for easy customization without modifying the source code. It allows you to update settings in the event of:

1. The upstream Zig project rotates their signing keys and the hardcoded default is outdated
2. The upstream Zig project has moved to a new download URL

### Build-time Configuration (Linker Flags)

When building from source, you can also customize some defaults using linker flags:

```bash
go build -ldflags="-X 'github.com/exilesprx/zig-install/internal/config.Version=VERSION'"
```

The justfile in this project automatically reads from the `.env` file and sets the defaults during build.

## Examples

Install both Zig and ZLS:

```bash
sudo ./zig-install-linux-amd64 install
```

Install only Zig:

```bash
sudo ./zig-install-linux-amd64 install --zig-only
```

Install only ZLS (Zig Language Server):
_Note: You must have Zig installed in order to compile ZLS._

```bash
sudo ./zig-install-linux-amd64 install --zls-only
```

Display the current settings:

```bash
sudo ./zig-install-linux-amd64 install --settings
```

Generate a template .env file:

```bash
sudo ./zig-install-linux-amd64 install --generate-env
```

Show version information:

```bash
./zig-install-linux-amd64 version
```

Install specific version:

```bash
./zig-install-linux-amd64 --version 0.14.0 install
```

## Notes

- This program must be run as root as it installs software to system directories
- Configuration via .env file allows for easy customization without rebuilding

## License

This project is licensed under the MIT License.
