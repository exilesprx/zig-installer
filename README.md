# Zig and ZLS Installer

This script installs Zig and the Zig Language Server (ZLS). You can choose to install both or just one of them.

## Prerequisites

Before running the script, ensure you have the following dependencies installed:

- `Bash`
- `jq`
- `Wget` (for downloading Zig binary)
- `Git` (for downloading ZLS)
- `minisign` (for verifying Zig download)

## Installation

1. Clone the repository or download the script.
2. Make the script executable:

```bash
chomod +x install.sh
```

3. Run the script with the desired options.

## Usage

```bash
./install.sh [OPTIONS]
```

### Options

- `--zig-only`: Install only Zig.

- `--zls-only`: Install only ZLS (Zig Language Server).

- `-h`, `--help`: Display the help message and exit.

## Examples

Install both Zig and ZLS:

```bash
./install.sh
```

Install only Zig:

```bash
./install.sh --zig-only
```

Install only ZLS (Zig Language Server).  
 _Note: You must have Zig installed in order to compile ZLS._

```bash
./install.sh --zls-only
```

Display the help message:

```bash
./install.sh --help
```

## Notes

- This script assumes you have the necessary permissions to install software on your system.
- Make sure to run the script from a directory where you have write access.

## License

This project is licensed under the MIT License.
