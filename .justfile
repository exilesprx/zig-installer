#!/usr/bin/env just --justfile

default: help

# Common build variables - using Just's variable support
module := "github.com/exilesprx/zig-install"
package := module + "/cmd"
config_package := module + "/internal/config"

# Helper recipe to setup build environment (not called directly)
_setup:
  #!/usr/bin/env bash
  set -euo pipefail
  VERSION=$(git describe --exact-match --tags 2>/dev/null || echo "$(git rev-parse --short=12 HEAD)-dev")
  COMMIT=$(git rev-parse HEAD)
  DATE=$(date)
  # Use environment variable directly or fall back to default if not set
  ZIG_KEY=${ZIG_PUB_KEY:-"RWSGOq2NVecA2UPNdBUZykf1CCb147pkmdtYxgb3Ti+JO/wCYvhbAb/U"}
  echo "-X {{package}}.Version=$VERSION -X {{package}}.Commit=$COMMIT -X '{{package}}.BuildDate=$DATE' -X '{{config_package}}.DefaultZigPubKey=$ZIG_KEY'"

# Display available commands
help:
  @echo "Available commands:"
  @echo "  build                   - Build for current platform"
  @echo "  build-all               - Build for all platforms (linux, windows, mac)"
  @echo "  build-linux             - Build for Linux (amd64, arm64)"
  @echo "  build-windows           - Build for Windows (amd64, arm64)"
  @echo "  build-mac               - Build for macOS (amd64, arm64)"

# Build for current platform
build:
  @echo "Building for current platform..."
  go build -ldflags="$(just _setup)"

# Build for a specific OS/ARCH combination
_build os arch suffix="":
  @echo "Building for {{os}}/{{arch}}..."
  GOOS={{os}} GOARCH={{arch}} go build -o zig-install-{{os}}-{{arch}}{{suffix}} -ldflags="$(just _setup)"

# Build for all platforms
build-all: build-linux build-windows build-mac
  @echo "All builds completed"

# Build for Linux
build-linux:
  @just _build linux amd64
  @just _build linux arm64

# Build for Windows
build-windows:
  # TODO: Update .env file to include the correct paths for Windows
  @just _build windows amd64 .exe
  @just _build windows arm64 .exe

# Build for macOS (using darwin as the OS name)
build-mac:
  # TODO: Update .env file to include the correct paths for macOS
  @just _build darwin amd64
  @just _build darwin arm64
