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
2. If downloading the source code, build and install the program:

```bash
go install
```

Or use the justfile if available:

```bash
just install
```

> Note: using `go install` does not add build metadata like version information. To include that, use the justfile or build with ldflags. See details in "Build-time Configuration" section.

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
- `--settings`: Show current configuration settings
- `--log-file <file>`: Specify log file (default: `zig-install.log`)
- `--enable-log`: Enable/disable logging to file (enabled by default)
- `--version, -v <version>`: Specify Zig version to install (default: latest master)

## Configuration

This program can be configured in two ways (in order of precedence):

1. **Command-line flags**: Options provided directly when running the program
2. **Configuration file**: Settings in an `.env` file

Before running the program, it will check for required dependencies like `wget`, `git`, `jq`, `minisign`, and `xz`. If any are missing, it will inform you so you can install them.

### Configuration File (.env)

You can create a `.env` file in the same directory as the executable in two ways:

1. Use the `env` command to create a template:
   ```bash
   ./zig-install-linux-amd64 env
   ```
2. Use the `--generate-env` flag with the install command:
   ```bash
   sudo ./zig-install-linux-amd64 install --generate-env
   ```

You can view your current configuration settings at any time using the `--settings` flag:

```bash
./zig-install-linux-amd64 install --settings
```

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
go build -ldflags="-X 'github.com/exilesprx/zig-install/internal/config.Version=VERSION' 'github.com/exilesprx/zig-install/internal/config.Commit=COMMIT' 'github.com/exilesprx/zig-install/internal/config.BuildDate=DATE'"
```

The justfile in this project automatically sets the ldflags during build and is the recommended way to build.

## Examples

Install both Zig and ZLS (latest master):

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

Install a specific version (both Zig and ZLS will be installed at this version):

```bash
sudo ./zig-install-linux-amd64 install --version 0.11.0
```

Install only Zig at a specific version:

```bash
sudo ./zig-install-linux-amd64 install --zig-only --version 0.11.0
```

Install only ZLS (will use current Zig version regardless of --version):

```bash
sudo ./zig-install-linux-amd64 install --zls-only --version 0.11.0  # Note: version will be ignored
```

Install with verbose output and custom log file:

```bash
sudo ./zig-install-linux-amd64 install --verbose --log-file custom.log
```

Display the current settings:

```bash
./zig-install-linux-amd64 install --settings
```

Generate a template .env file:

```bash
./zig-install-linux-amd64 env
```

Show version information:

```bash
./zig-install-linux-amd64 version
```

## Version Management

The installer manages Zig and ZLS versions in the following way:

- When using `--version`, both Zig and ZLS will be installed at the specified version to ensure compatibility
- When using `--zig-only` with `--version`, only Zig will be installed at the specified version
- When using `--zls-only` with `--version`, ZLS will be installed matching your current Zig version, ignoring the specified version
- If no version is specified, the latest master versions will be used

This versioning strategy ensures that Zig and ZLS remain compatible with each other.

## Notes

- This program must be run as root as it installs software to system directories
- Configuration via .env file allows for easy customization without rebuilding
- Logging is enabled by default to `zig-install.log`, but can be configured or disabled
- The program performs automatic dependency checks before installation
- Both Zig and ZLS installations preserve file ownership for non-root users

## License

This project is licensed under the MIT License.
