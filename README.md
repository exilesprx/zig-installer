# Zig and ZLS Installer

> **BREAKING CHANGE - v4.0.0:** This installer now uses **user-local installation only** (no more sudo required). If you have an existing system-wide installation, see the [Migration Guide](#migrating-from-system-installation) below.

This program installs Zig and the Zig Language Server (ZLS) to your local user directory. You can choose to install both or just one of them.

**Platform Support:**

- **Linux:** Fully supported
- **macOS:** Experimental support (ARM64 and x86_64)

**Important:** Do NOT run this installer with sudo. It installs to your home directory and does not require root privileges.

## Installation Directories

The installer uses the following user-local directories:

- **Zig installations:** `~/.local/share/zig/`
- **ZLS installations:** `~/.local/share/zls/`
- **Symlinks (zig, zls):** `~/.local/bin/`

No system-wide directories (`/opt`, `/usr/local`) are used.

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

> **Note**: using `go install` does not add build metadata like version information. To include that, use the justfile or build with ldflags. See details in "Build-time Configuration" section.

3. Run the program with the desired options (no sudo required).

## Usage

```bash
./zig-installer [command] [OPTIONS]
```

### Commands

- `install`: Install Zig and/or ZLS (default if no command specified)
- `migrate`: Migrate from a system-wide installation to user-local installation
- `cleanup`: Interactively clean up old Zig versions
- `version`: Show version information about the installer
- `env`: Generate a template .env file

### Options

#### Installation Options

- `--zig-only`: Install only Zig
- `--zls-only`: Install only ZLS (Zig Language Server)
- `--version, -v <version>`: Specify Zig version to install (default: latest master)

#### Cleanup Options

- `--auto-cleanup`: Automatically cleanup old versions after install without prompting
- `--no-cleanup`: Disable auto-cleanup prompt after install
- `--keep-last <N>`: Keep last N versions when cleaning up

#### Output Control

- `--verbose`: Show detailed output during installation
- `--no-color`: Disable colored output

#### Configuration

- `--env <file>`: Specify a custom environment file (default: `.env`)
- `--settings`: Show current configuration settings

#### Logging

- `--log-file <file>`: Specify log file (default: `zig-install.log`)
- `--enable-log`: Enable/disable logging to file (enabled by default)

## Migrating from System Installation

If you previously used an older version of this installer that installed Zig to `/opt/zig` or `/usr/local/zig`, you have two options:

### Option 1: Use the Migrate Command (Recommended)

Run the migrate command to automatically remove your system installation:

```bash
./zig-installer migrate
```

This will:

1. Detect system installations of Zig and ZLS
2. Remove them from `/opt/zig`, `/usr/local/zig`, `/opt/zls`, `/usr/local/zls`
3. Remove symlinks from `/usr/local/bin/zig` and `/usr/local/bin/zls`
4. Prompt for your password if needed (via sudo)

After migration, run:

```bash
./zig-installer install
```

### Option 2: Automatic Migration During Install

When you run `install`, the installer will detect any existing system installation and offer three choices:

```bash
./zig-installer install
```

You'll see:

```
⚠ Found existing system-wide installation
? How would you like to proceed?
  > Migrate to user-local (remove system installation) [Recommended]
    Keep both (may cause PATH conflicts)
    Cancel installation
```

- **Migrate:** Removes system installation and proceeds with user-local install
- **Keep both:** Installs user-local version alongside system version (you'll get PATH priority warnings)
- **Cancel:** Exits without making changes

### Manual Migration

If you prefer to manually remove the system installation:

**Linux:**

```bash
sudo rm -rf /opt/zig /opt/zls
sudo rm -f /usr/local/bin/zig /usr/local/bin/zls
```

**macOS:**

```bash
sudo rm -rf /usr/local/zig /usr/local/zls /opt/zig /opt/zls
sudo rm -f /usr/local/bin/zig /usr/local/bin/zls
```

Then install the user-local version:

```bash
./zig-installer install
```

## PATH Configuration

After installation, ensure `~/.local/bin` is in your PATH. The installer will check this automatically and provide instructions if needed.

### Bash / Zsh

Add to `~/.bashrc` or `~/.zshrc`:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

Then reload your shell:

```bash
source ~/.bashrc  # or source ~/.zshrc
```

### Fish

Add to `~/.config/fish/config.fish`:

```fish
set -gx PATH $HOME/.local/bin $PATH
```

Then reload:

```fish
source ~/.config/fish/config.fish
```

### Verify PATH

Check that zig is available:

```bash
which zig
# Should output: /home/yourusername/.local/bin/zig

zig version
```

## Configuration

This program can be configured in two ways (in order of precedence):

1. **Command-line flags**: Options provided directly when running the program
2. **Configuration file**: Settings in an `.env` file

Before running the program, it will check for required dependencies like `wget`, `git`, `jq`, `minisign`, and `xz`. If any are missing, it will inform you so you can install them.

### Configuration File (.env)

You can create a `.env` file in the same directory as the executable in two ways:

1. Use the `env` command to create a template:
   ```bash
   ./zig-installer env
   ```
2. Use the `--generate-env` flag with the install command:
   ```bash
   ./zig-installer install --generate-env
   ```

You can view your current configuration settings at any time using the `--settings` flag:

```bash
./zig-installer install --settings
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

### Basic Installation

Install both Zig and ZLS (latest master):

```bash
./zig-installer install
```

Install only Zig:

```bash
./zig-installer install --zig-only
```

Install only ZLS (Zig Language Server):
_Note: You must have Zig installed in order to compile ZLS._

```bash
./zig-installer install --zls-only
```

### Version-Specific Installation

Install a specific version (both Zig and ZLS will be installed at this version):

```bash
./zig-installer install --version 0.11.0
```

Install only Zig at a specific version:

```bash
./zig-installer install --zig-only --version 0.11.0
```

Install only ZLS (will use current Zig version regardless of --version):

```bash
./zig-installer install --zls-only --version 0.11.0  # Note: version will be ignored
```

### Installation with Options

Install with verbose output and custom log file:

```bash
./zig-installer install --verbose --log-file custom.log
```

Install a specific version with auto-cleanup (keep last 2 versions):

```bash
./zig-installer install --version 0.13.0 --auto-cleanup --keep-last 2
```

Install without cleanup prompt:

```bash
./zig-installer install --version 0.13.0 --no-cleanup
```

### Cleanup Commands

Cleanup old versions interactively:

```bash
./zig-installer cleanup
```

Cleanup keeping last 3 versions:

```bash
./zig-installer cleanup --keep-last 3 --yes
```

Dry run to see what would be removed:

```bash
./zig-installer cleanup --dry-run
```

### Other Commands

Display the current settings:

```bash
./zig-installer install --settings
```

Generate a template .env file:

```bash
./zig-installer env
```

Show version information:

```bash
./zig-installer version
```

Migrate from system installation:

```bash
./zig-installer migrate
```

## Cleanup Old Versions

The installer provides powerful tools to manage disk space by removing old Zig versions.

### Manual Cleanup

Run the cleanup command to interactively select versions to remove:

```bash
./zig-installer cleanup
```

This will:

1. Scan for all installed Zig versions in `~/.local/share/zig/`
2. Display a table showing version, size, install date, and current status
3. Allow you to select which versions to remove (the current version cannot be removed)
4. Ask for confirmation before removing
5. Show how much disk space was freed

**Note:** The cleanup command only manages user-local installations (`~/.local/share/zig/`). If you have a system installation in `/opt/zig` or `/usr/local/zig`, the cleanup command will warn you and provide instructions for manual removal.

#### Cleanup Command Options

**Interactive Mode (Default)**

```bash
# Select versions manually from an interactive list
./zig-installer cleanup
```

**Dry Run Mode**

```bash
# See what would be removed without actually removing anything
./zig-installer cleanup --dry-run
```

**Auto-Cleanup with Keep-Last**

```bash
# Keep the last 3 versions, automatically remove older ones
./zig-installer cleanup --keep-last 3

# Skip confirmation prompt with --yes flag
./zig-installer cleanup --keep-last 3 --yes
```

### Auto-Cleanup After Install

**By default**, when you install a specific Zig version, the installer will **automatically prompt** you to clean up old versions after a successful installation.

```bash
# Install Zig 0.13.0 - will prompt for cleanup after install
./zig-installer install --version 0.13.0

# After successful installation, you'll see:
# ℹ Found 3 other installed versions (308 MB)
# ? Clean up old versions? [Y/n]:
```

#### Controlling Auto-Cleanup Behavior

**Disable Auto-Cleanup Prompt**

```bash
# Install without cleanup prompt (opt-out)
./zig-installer install --version 0.13.0 --no-cleanup
```

**Auto-Cleanup Without Prompting**

```bash
# Install and automatically show cleanup selection UI
./zig-installer install --version 0.13.0 --auto-cleanup

# Install and automatically keep last 2 versions
./zig-installer install --version 0.13.0 --auto-cleanup --keep-last 2
```

### Cleanup Examples

**Example 1: Interactive Cleanup**

```bash
$ ./zig-installer cleanup

==> Scanning for installed Zig versions...
  --> Success: Found 4 installed versions

┌─────────────────────────┬─────────┬──────────────┬─────────┐
│ Version                 │ Size    │ Install Date │ Current │
├─────────────────────────┼─────────┼──────────────┼─────────┤
│ 0.13.0                  │ 127 MB  │ 2024-01-15   │    ✓    │
│ 0.12.0                  │ 115 MB  │ 2023-11-10   │         │
│ 0.11.0                  │ 98 MB   │ 2023-08-05   │         │
│ 0.10.1                  │ 95 MB   │ 2023-05-20   │         │
└─────────────────────────┴─────────┴──────────────┴─────────┘

Total disk usage: 435 MB

? Select versions to remove (space to select, enter to confirm):
  [x] 0.12.0
  [ ] 0.11.0
  [x] 0.10.1

==> Removing versions...
  --> Success: Removed 0.12.0 (115 MB)
  --> Success: Removed 0.10.1 (95 MB)

✓ Cleanup completed successfully!
✓ Freed 210 MB of disk space
```

**Example 2: Auto-Cleanup After Install**

```bash
$ ./zig-installer install --version 0.13.0

==> Zig Installation
  --> Success: Zig 0.13.0 installed and configured

ℹ Found 3 other installed versions (308 MB)

? Clean up old versions? [Y/n]: y

[Interactive selection UI shown...]

==> Removing 2 versions...
  --> Success: Removed 0.12.0
  --> Success: Removed 0.10.1

✓ Freed 210 MB of disk space
```

**Example 3: Keep Last N Versions**

```bash
$ ./zig-installer cleanup --keep-last 2 --yes

==> Auto-cleanup mode (keeping last 2 versions)

Keeping: 0.13.0, 0.12.0
Removing: 0.11.0, 0.10.1

==> Removing versions...
  --> Success: Removed 0.11.0 (98 MB)
  --> Success: Removed 0.10.1 (95 MB)

✓ Cleanup completed successfully!
✓ Freed 193 MB of disk space
```

## Version Management

The installer manages Zig and ZLS versions in the following way:

- When using `--version`, both Zig and ZLS will be installed at the specified version to ensure compatibility
- When using `--zig-only` with `--version`, only Zig will be installed at the specified version
- When using `--zls-only` with `--version`, ZLS will be installed matching your current Zig version, ignoring the specified version
- If no version is specified, the latest master versions will be used

This versioning strategy ensures that Zig and ZLS remain compatible with each other.

## Frequently Asked Questions

### Why did the installer change to user-local only?

**User-local installation provides several benefits:**

1. **No sudo required:** Safer and more convenient - no risk of accidentally damaging system files
2. **Per-user versions:** Each user can have their own Zig versions without conflicts
3. **Standard practice:** Follows the same pattern as rustup, nvm, pyenv, and other modern language installers
4. **Easier cleanup:** No permission issues when removing old versions
5. **Better isolation:** Development tools shouldn't require system-wide installation

### What if I run the installer with sudo?

The installer will **reject sudo** and display an error message:

```
✗ Error: This installer should NOT be run with sudo

As of v4.0.0, zig-installer uses user-local installation only.
It installs to ~/.local/share/zig and does not require root privileges.

If you have an existing system-wide installation, please run:
  zig-installer migrate
```

### Can I still use my old system installation?

Yes, but it's not recommended. If you keep both:

- **PATH priority matters:** `~/.local/bin` should come before `/usr/local/bin` in your PATH
- **Confusion risk:** Having multiple installations can lead to using the wrong version
- **No cleanup support:** The installer's cleanup command won't manage system installations

The installer will warn you about PATH conflicts if you choose to keep both.

### How do I uninstall completely?

**Remove user-local installation:**

```bash
rm -rf ~/.local/share/zig ~/.local/share/zls
rm -f ~/.local/bin/zig ~/.local/bin/zls
```

**Remove system installation (if you have one):**

```bash
sudo rm -rf /opt/zig /opt/zls /usr/local/zig /usr/local/zls
sudo rm -f /usr/local/bin/zig /usr/local/bin/zls
```

### Does this work on macOS?

Yes, but **macOS support is experimental**. The installer will display a warning on macOS:

```
⚠ Warning: macOS support is experimental
  Please report any issues at: https://github.com/exilesprx/zig-install/issues
```

Both ARM64 (Apple Silicon) and x86_64 (Intel) are supported.

### What if ~/.local/bin is not in my PATH?

After installation, the installer automatically checks your PATH. If `~/.local/bin` is not found, it will show shell-specific instructions:

**For Bash/Zsh:**

```bash
export PATH="$HOME/.local/bin:$PATH"
```

**For Fish:**

```fish
set -gx PATH $HOME/.local/bin $PATH
```

You'll need to add this to your shell's config file and reload.

### Can I install to a custom directory?

Not currently. The installer uses `~/.local/share/zig` and `~/.local/share/zls` following the XDG Base Directory specification. This is a standard location for user-specific data files.

### What happens to my old versions after migration?

When you run `zig-installer migrate` or choose "Migrate" during install:

1. **System installations are removed:** All versions in `/opt/zig`, `/usr/local/zig`, etc.
2. **Symlinks are removed:** `/usr/local/bin/zig` and `/usr/local/bin/zls`
3. **User-local versions are preserved:** Any versions you've already installed to `~/.local/share/zig` remain

After migration, run `zig-installer install` to install the version you need.

### How do I check which installation I'm using?

```bash
which zig
# User-local: /home/yourusername/.local/bin/zig
# System-wide: /usr/local/bin/zig

zig version
```

If `which zig` shows `/usr/local/bin/zig`, you're still using a system installation.

## Notes

- This installer uses **user-local directories only** - no system-wide installation
- Do **NOT** run with sudo - the installer will reject it
- Configuration via .env file allows for easy customization without rebuilding
- Logging is enabled by default to `zig-install.log`, but can be configured or disabled
- The program performs automatic dependency checks before installation
- The program automatically checks if `~/.local/bin` is in your PATH
- Auto-cleanup is enabled by default when installing specific versions (can be disabled with `--no-cleanup`)
- The cleanup command protects the currently active version from accidental removal
- Multiple Zig versions can be installed side-by-side in `~/.local/share/zig/`
- The cleanup command only manages user-local installations

## Changelog

### v4.0.0 - Breaking Changes

**Major Changes:**

- **BREAKING:** Removed system-wide installation support - user-local only (`~/.local/`)
- **BREAKING:** Installer now rejects sudo - no root privileges required or allowed
- **BREAKING:** Binary renamed from `zig-install-{platform}` to `zig-installer`

**New Features:**

- Added `migrate` command to migrate from system installations
- Automatic detection of existing system installations with migration prompt
- PATH configuration detection with shell-specific instructions
- Warning when system installation exists alongside user-local installation

**Installation Directories Changed:**

- Old: `/opt/zig` (Linux), `/usr/local/zig` (macOS)
- New: `~/.local/share/zig` (all platforms)
- Old: `/opt/zls` (Linux), `/usr/local/zls` (macOS)
- New: `~/.local/share/zls` (all platforms)
- Old: `/usr/local/bin/zig` (symlink)
- New: `~/.local/bin/zig` (symlink)

**Cleanup Command Changes:**

- Cleanup now only manages `~/.local/share/zig/` versions
- Cleanup warns if system installation detected
- No longer requires sudo

**Migration Path:**
Run `zig-installer migrate` to remove system installations, or use the automatic migration prompt during `install`.

## Troubleshooting

### Migration fails with "permission denied"

If `zig-installer migrate` fails with permission errors, you may need to manually remove the system installation:

**Linux:**

```bash
sudo rm -rf /opt/zig /opt/zls /usr/local/bin/zig /usr/local/bin/zls
```

**macOS:**

```bash
sudo rm -rf /usr/local/zig /usr/local/zls /opt/zig /opt/zls /usr/local/bin/zig /usr/local/bin/zls
```

After manual cleanup, you can verify with:

```bash
ls -la /usr/local/bin/zig  # Should show "No such file or directory"
ls -la /opt/zig            # Should show "No such file or directory"
```

Then install the user-local version:

```bash
./zig-installer install
```

### "command not found" after migration

If you get "zig: command not found" after migration, ensure `~/.local/bin` is in your PATH. See the [PATH Configuration](#path-configuration) section above.

### Multiple zig versions found

If `which -a zig` shows multiple zig installations:

```bash
$ which -a zig
/home/user/.local/bin/zig       # User-local (preferred)
/usr/local/bin/zig              # System-wide (old)
```

The first one in your PATH will be used. To remove the system installation, see "Migration fails with permission denied" above.

## License

This project is licensed under the MIT License.
