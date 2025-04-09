package installer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/exilesprx/zig-install/internal/config"
	"github.com/exilesprx/zig-install/internal/logger"
	"github.com/exilesprx/zig-install/internal/tui"
)

// InstallZLS handles the ZLS installation process
func InstallZLS(p *tea.Program, config *config.Config, logger logger.ILogger, zigVersion string) error {
	// Prepare directories
	if err := os.MkdirAll(config.ZLSDir, 0755); err != nil {
		return fmt.Errorf("could not create directory %s: %w", config.ZLSDir, err)
	}

	// Get the username to set ownership
	user := os.Getenv("SUDO_USER")
	if user == "" {
		user = os.Getenv("USER")
	}

	// Set ownership
	if user != "" {
		if config.Verbose {
			p.Send(tui.DetailOutputMsg(fmt.Sprintf("Setting ownership of %s to %s", config.ZLSDir, user)))
		}

		cmd := exec.Command("chown", "-R", user+":"+user, config.ZLSDir)
		if output, err := cmd.CombinedOutput(); err != nil {
			if config.Verbose {
				p.Send(tui.DetailOutputMsg(fmt.Sprintf("Error: %s", output)))
			}
			return fmt.Errorf("could not set ownership of %s: %w", config.ZLSDir, err)
		} else if config.Verbose && len(output) > 0 {
			p.Send(tui.DetailOutputMsg(string(output)))
		}
	}

	// Determine if a matching ZLS tag exists for this Zig version
	zlsTag := convertToSemanticVersion(zigVersion)
	hasMatchingTag := checkZLSTagExists(zlsTag, logger)

	if hasMatchingTag {
		p.Send(tui.StatusMsg(fmt.Sprintf("Found matching ZLS tag for Zig %s", zigVersion)))
		logger.LogInfo("Found matching ZLS tag: %s", zlsTag)
		if config.Verbose {
			p.Send(tui.DetailOutputMsg(fmt.Sprintf("Found matching ZLS tag: %s", zlsTag)))
		}
	} else {
		p.Send(tui.StatusMsg(fmt.Sprintf("No matching ZLS tag for Zig %s, using master branch", zigVersion)))
		logger.LogInfo("No matching ZLS tag for Zig %s, using master branch", zigVersion)
		if config.Verbose {
			p.Send(tui.DetailOutputMsg("No matching ZLS tag, using master branch"))
		}
		zlsTag = ""
	}

	// Check if ZLS directory exists and update it
	if _, err := os.Stat(filepath.Join(config.ZLSDir, ".git")); err == nil {
		p.Send(tui.StatusMsg("Updating ZLS repository..."))

		// If we have a matching tag, checkout that specific tag
		if hasMatchingTag {
			if config.Verbose {
				p.Send(tui.DetailOutputMsg(fmt.Sprintf("Running: git fetch --tags && git checkout %s in %s", zlsTag, config.ZLSDir)))
			}

			// First fetch all tags
			fetchCmd := exec.Command("git", "fetch", "--tags")
			fetchCmd.Dir = config.ZLSDir

			output, err := fetchCmd.CombinedOutput()
			if err != nil {
				if config.Verbose {
					p.Send(tui.DetailOutputMsg(fmt.Sprintf("Error fetching tags: %s", output)))
				}
				return fmt.Errorf("could not fetch ZLS tags: %w", err)
			}

			if config.Verbose {
				p.Send(tui.DetailOutputMsg(string(output)))
			}

			// Then checkout the specific tag
			checkoutCmd := exec.Command("git", "checkout", zlsTag)
			checkoutCmd.Dir = config.ZLSDir

			output, err = checkoutCmd.CombinedOutput()
			if err != nil {
				if config.Verbose {
					p.Send(tui.DetailOutputMsg(fmt.Sprintf("Error checking out tag: %s", output)))
				}
				return fmt.Errorf("could not checkout ZLS tag %s: %w", zlsTag, err)
			}

			if config.Verbose {
				p.Send(tui.DetailOutputMsg(string(output)))
			}
		} else {
			// No matching tag, just pull the latest master
			if config.Verbose {
				p.Send(tui.DetailOutputMsg("Running: git checkout master && git pull in " + config.ZLSDir))
			}

			// First make sure we're on master
			checkoutCmd := exec.Command("git", "checkout", "master")
			checkoutCmd.Dir = config.ZLSDir

			output, err := checkoutCmd.CombinedOutput()
			if err != nil {
				if config.Verbose {
					p.Send(tui.DetailOutputMsg(fmt.Sprintf("Error checking out master: %s", output)))
				}
				// Try main branch if master fails
				checkoutCmd = exec.Command("git", "checkout", "main")
				checkoutCmd.Dir = config.ZLSDir

				output, err = checkoutCmd.CombinedOutput()
				if err != nil {
					if config.Verbose {
						p.Send(tui.DetailOutputMsg(fmt.Sprintf("Error checking out main: %s", output)))
					}
					return fmt.Errorf("could not checkout ZLS master/main branch: %w", err)
				}
			}

			if config.Verbose {
				p.Send(tui.DetailOutputMsg(string(output)))
			}

			// Then pull the latest changes
			cmd := exec.Command("git", "pull")
			cmd.Dir = config.ZLSDir

			output, err = cmd.CombinedOutput()
			if err != nil {
				if config.Verbose {
					p.Send(tui.DetailOutputMsg(fmt.Sprintf("Error updating repository: %s", output)))
				}
				return fmt.Errorf("could not update ZLS repository: %w", err)
			}

			if config.Verbose {
				p.Send(tui.DetailOutputMsg(string(output)))
			}
		}
	} else {
		// Clone the repository
		p.Send(tui.StatusMsg("Cloning ZLS repository..."))

		var cloneCmd *exec.Cmd

		if hasMatchingTag {
			if config.Verbose {
				p.Send(tui.DetailOutputMsg(fmt.Sprintf("Running: git clone -b %s --depth 1 https://github.com/zigtools/zls.git %s", zlsTag, config.ZLSDir)))
			}

			// Clone with specific tag/branch
			cloneCmd = exec.Command("git", "clone", "-b", zlsTag, "--depth", "1", "https://github.com/zigtools/zls.git", config.ZLSDir)
		} else {
			if config.Verbose {
				p.Send(tui.DetailOutputMsg(fmt.Sprintf("Running: git clone --depth 1 https://github.com/zigtools/zls.git %s", config.ZLSDir)))
			}

			// Clone default branch with depth 1
			cloneCmd = exec.Command("git", "clone", "--depth", "1", "https://github.com/zigtools/zls.git", config.ZLSDir)
		}

		output, err := cloneCmd.CombinedOutput()
		if err != nil {
			if config.Verbose {
				p.Send(tui.DetailOutputMsg(fmt.Sprintf("Error cloning repository: %s", output)))
			}
			return fmt.Errorf("could not clone ZLS repository: %w", err)
		}

		if config.Verbose {
			p.Send(tui.DetailOutputMsg(string(output)))
		}
	}

	// Build ZLS
	p.Send(tui.StatusMsg("Building ZLS..."))

	if config.Verbose {
		p.Send(tui.DetailOutputMsg(fmt.Sprintf("Running: zig build -Doptimize=ReleaseSafe in %s", config.ZLSDir)))
	}

	cmd := exec.Command("zig", "build", "-Doptimize=ReleaseSafe")
	cmd.Dir = config.ZLSDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		if config.Verbose {
			p.Send(tui.DetailOutputMsg(fmt.Sprintf("Error building ZLS: %s", output)))
		}
		return fmt.Errorf("could not build ZLS: %w", err)
	}

	if config.Verbose {
		p.Send(tui.DetailOutputMsg(string(output)))
	}

	// Create symbolic link
	zlsBinPath := filepath.Join(config.ZLSDir, "zig-out", "bin", "zls")
	linkPath := filepath.Join(config.BinDir, "zls")

	p.Send(tui.StatusMsg("Creating symbolic link..."))

	if config.Verbose {
		p.Send(tui.DetailOutputMsg(fmt.Sprintf("Creating symlink from %s to %s", zlsBinPath, linkPath)))
	}

	if _, err := os.Stat(linkPath); err == nil {
		os.Remove(linkPath)
		if config.Verbose {
			p.Send(tui.DetailOutputMsg(fmt.Sprintf("Removed existing symlink at %s", linkPath)))
		}
	}

	if err := os.Symlink(zlsBinPath, linkPath); err != nil {
		return fmt.Errorf("could not create symbolic link: %w", err)
	}

	if config.Verbose {
		p.Send(tui.DetailOutputMsg("Symlink created successfully"))
	}

	return nil
}

// checkZLSTagExists verifies if a specific tag exists in the ZLS repository
func checkZLSTagExists(tag string, logger logger.ILogger) bool {
	cmd := exec.Command("git", "ls-remote", "--tags", "https://github.com/zigtools/zls.git", tag)
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.LogError("Error checking ZLS tag: %v", err)
		return false
	}

	// If the output contains the tag, it exists
	return strings.Contains(string(output), tag)
}

// convertToSemanticVersion converts a Zig version like "0.12.0-dev.1234+hash" to semantic version format "0.12.0"
func convertToSemanticVersion(zigVersion string) string {
	// Extract the semantic version part (major.minor.patch)
	parts := strings.Split(zigVersion, "-")
	if len(parts) > 0 {
		// Further ensure we only have three components (major.minor.patch)
		versionParts := strings.Split(parts[0], ".")
		if len(versionParts) >= 3 {
			return fmt.Sprintf("%s.%s.%s", versionParts[0], versionParts[1], versionParts[2])
		}
		return parts[0]
	}
	return zigVersion
}
