# Zig and ZLS Installer

This script installs Zig and the Zig Language Server (ZLS). You can choose to install both or just one of them.

## Usage

```sh
./install.sh [OPTIONS]
```

### Options

- `--zig-only`  
  Install only Zig.

- `--zls-only`  
  Install only ZLS (Zig Language Server).

- `-h`, `--help`  
  Display the help message and exit.

## Examples

Install both Zig and ZLS:

```sh
./install.sh
```

Install only Zig:

```sh
./install.sh --zig-only
```

Install only ZLS (Zig Language Server).  
 _Note: You must have Zig installed in order to compile ZLS._

```sh
./install.sh --zls-only
```

Display the help message:

```sh
./install.sh --help
```

## Requirements

- Bash
- Wget (for downloading Zig binary)
- Git (for downloading ZLS)

## Installation

1. Clone the repository or download the script.
2. Make the script executable:

```sh
chomod +x install.sh
```

3. Run the script with the desired options.

## Notes

- This script assumes you have the necessary permissions to install software on your system.
- Make sure to run the script from a directory where you have write access.

## License

This project is licensed under the MIT License.
