#!/usr/bin/env just --justfile

set dotenv-load

default: help

# Common build variables - using Just's variable support
module := "github.com/exilesprx/zig-install"
package := module + "/cmd"
config_package := module + "/internal/config"
zig_path := "/opt/zig"
zls_path := "/opt/zls"

# Helper recipe to setup build environment (not called directly)
_setup:
  #!/usr/bin/env bash
  set -euo pipefail
  VERSION=$(git describe --exact-match --tags 2>/dev/null || echo "$(git rev-parse --short=12 HEAD)-dev")
  COMMIT=$(git rev-parse HEAD)
  DATE=$(date)
  # Use environment variable directly or fall back to default if not set
  echo "-X {{package}}.Version=$VERSION -X {{package}}.Commit=$COMMIT -X '{{package}}.BuildDate=$DATE'"

# Display available commands
help:
  @echo "Available commands:"
  @echo "  build                   - Build for current platform"
  @echo "  build-all               - Build for all platforms (linux, windows, mac)"
  @echo "  build-linux             - Build for Linux (amd64)"
  @echo "  build-mac               - Build for macOS (amd64)"
  @echo "  lint                    - Run golangci-lint on the codebase"

fmt:
    gofumpt -l -w .

lint:
    golangci-lint run -v ./...

clean:
  @echo "Cleaning build artifacts"
  rm -rf {{zls_path}}/.zig-cache {{zls_path}}/zig-out zig-install*

uninstall:
  @echo "Uninstalling zig-install and removing {{zig_path}} and {{zls_path}}"
  #!/usr/bin/env bash
  set -euo pipefail
  zig_target=$(realpath "$(command -v zig)" 2>/dev/null)
  zls_target=$(realpath "$(command -v zls)" 2>/dev/null)
  [ -n "$zig_target" ] && [ -e "$zig_target" ] && [[ "$zig_target" == "$zig_path"/* ]] && rm -rf "$zig_target"
  [ -n "$zls_target" ] && [ -e "$zls_target" ] && [[ "$zls_target" == "$zls_path"/* ]] && rm -rf "$zls_target"

# Build for current platform
build:
  @echo "Building for current platform..."
  go build -ldflags="$(just _setup)"

# Build for a specific OS/ARCH combination
_build os arch suffix="":
  @echo "Building for {{os}}/{{arch}}..."
  GOOS={{os}} GOARCH={{arch}} go build -o zig-install-{{os}}-{{arch}}{{suffix}} -ldflags="$(just _setup)"

# Build for all platforms
build-all: build-linux build-mac
  @echo "All builds completed"

# Build for Linux
build-linux:
  @just _build linux amd64

# Build for macOS (using darwin as the OS name)
build-mac:
  # TODO: Update .env file to include the correct paths for macOS
  @just _build darwin amd64
