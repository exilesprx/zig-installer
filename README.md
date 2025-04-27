# Zig and ZLS Installer

This program installs Zig and the Zig Language Server (ZLS). You can choose to install both or just one of them.

**Note:** This program must be run as root. Use `sudo` when executing the binary.

> **Platform Support:** Linux is fully supported. Windows and macOS builds are currently a work in progress.

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
sudo ./zig-install [command] [OPTIONS]
```

### Commands

- `install`: Install Zig and/or ZLS (default if no command specified)
- `version`: Show version information about the installer

### Options

- `--zig-only`: Install only Zig
- `--zls-only`: Install only ZLS (Zig Language Server)
- `--verbose`: Show detailed output during installation
- `--no-color`: Disable colored output
- `--env <file>`: Specify a custom environment file (default: `.env`)
- `--generate-env`: Generate a template .env file
- `--settings`: Show current settings
- `--log-file <file>`: Specify log file (default: `zig-install.log`)
- `--enable-log`: Enable/disable logging to file

## Configuration

This program can be configured in three ways (in order of precedence):

1. **Command-line flags**: Options provided directly when running the program
2. **Environment variables**: Variables with the prefix `ZIG_INSTALL_`
3. **Configuration file**: Settings in an `.env` file

### Available Environment Variables

All environment variables must be prefixed with `ZIG_INSTALL_`. For example, `ZIG_INSTALL_ZIG_DIR=/custom/path`.

| Environment Variable     | Description                                  | Default Value                                           |
|--------------------------|----------------------------------------------|---------------------------------------------------------|
| `ZIG_DIR`                | Directory to install Zig                     | `/opt/zig`                                              |
| `ZLS_DIR`                | Directory to install ZLS                     | `/opt/zls`                                              |
| `BIN_DIR`                | Directory for symlinks                       | `/usr/local/bin`                                        |
| `ZIG_PUB_KEY`            | Minisign public key to verify Zig download   | `RWSGOq2NVecA2UPNdBUZykf1CCb147pkmdtYxgb3Ti+JO/wCYvhbAb/U` |
| `ZIG_DOWN_URL`           | Base URL for Zig downloads                   | `https://ziglang.org/builds/`                           |
| `ZIG_INDEX_URL`          | URL for Zig version index                    | `https://ziglang.org/download/index.json`               |
| `ZIG_ONLY`               | Install only Zig (boolean)                   | `false`                                                 |
| `ZLS_ONLY`               | Install only ZLS (boolean)                   | `false`                                                 |
| `VERBOSE`                | Show detailed output (boolean)               | `false`                                                 |
| `NO_COLOR`               | Disable colored output (boolean)             | `false`                                                 |
| `LOG_FILE`               | File to log errors to                        | `zig-install.log`                                       |
| `ENABLE_LOG`             | Enable logging to file (boolean)             | `true`                                                  |

### Configuration File (.env)

You can create a `.env` file in the same directory as the executable or use the `--generate-env` flag to create a template:

```
# Zig & ZLS Installer Configuration
# Directories
ZIG_DIR=/opt/zig
ZLS_DIR=/opt/zls
BIN_DIR=/usr/local/bin

# Zig download and verification
ZIG_PUB_KEY=RWSGOq2NVecA2UPNdBUZykf1CCb147pkmdtYxgb3Ti+JO/wCYvhbAb/U
ZIG_DOWN_URL=https://ziglang.org/builds/
ZIG_INDEX_URL=https://ziglang.org/download/index.json
```

### Build-time Configuration (Linker Flags)

When building from source, you can also customize some defaults using linker flags:

```bash
go build -ldflags="-X 'github.com/exilesprx/zig-install/internal/config.DefaultZigPubKey=YOUR_KEY'"
```

The justfile in this project automatically reads the `ZIG_PUB_KEY` from the `.env` file if present and sets it as the default public key during build. This is useful when:

1. You need to use a custom or alternate signing key for Zig binaries
2. The upstream Zig project rotates their signing keys and the hardcoded default is outdated
3. You're in an air-gapped environment and need to verify Zig binaries against your organization's key
4. You're building a custom distribution of the installer with your own verification key

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

## Notes

- This program must be run as root as it installs software to system directories
- Configuration via .env file allows for easy customization without rebuilding

## License

This project is licensed under the MIT License.
