package preflight

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

// registry is the ordered list of all preflight checks.
var registry = []Check{
	checkRequiredPackages(),
	checkSSHConfigured(),
	checkAetherUserConfigured(),
	checkNodeSSHReachable(),
}

// checkIndex maps check IDs to their position in the registry.
var checkIndex = buildIndex(registry)

func buildIndex(checks []Check) map[string]int {
	idx := make(map[string]int, len(checks))
	for i, c := range checks {
		idx[c.ID] = i
	}
	return idx
}

// requiredBinaries lists the executables that must be present in PATH.
// AptPkg/YumPkg allow distro-aware installation.
var requiredBinaries = []struct {
	Name   string // binary name to look up
	AptPkg string // Debian/Ubuntu package name
	YumPkg string // RHEL/Fedora package name
}{
	{"make", "make", "make"},
	{"ansible-playbook", "ansible", "ansible"},
}

// packageManagers lists known package manager binaries in preference order.
var packageManagers = []struct {
	Binary  string // binary name to look up
	Install string // install subcommand template (use %s for package list)
}{
	{"apt-get", "sudo apt-get install -y %s"},
	{"dnf", "sudo dnf install -y %s"},
	{"yum", "sudo yum install -y %s"},
}

// ---------------------------------------------------------------------------
// 1. required-packages
// ---------------------------------------------------------------------------

func checkRequiredPackages() Check {
	return Check{
		ID:          "required-packages",
		Name:        "Required Packages",
		Description: "Checks that required build and deployment tools (make, ansible) are installed.",
		Severity:    SeverityRequired,
		Category:    CategoryTooling,
		FixWarning:  "This will install system packages using the detected package manager (apt-get, dnf, or yum). On Debian/Ubuntu, the Ansible PPA is added to provide the ansible package.",
		RunCheck: func(ctx context.Context, deps CheckDeps) CheckResult {
			r := newResult("required-packages", "Required Packages",
				"Checks that required build and deployment tools (make, ansible) are installed.",
				SeverityRequired, CategoryTooling, true)
			r.FixWarning = "This will install system packages using the detected package manager (apt-get, dnf, or yum). On Debian/Ubuntu, the Ansible PPA is added to provide the ansible package."

			var found, missing []string
			for _, bin := range requiredBinaries {
				path, err := deps.LookPath(bin.Name)
				if err != nil {
					missing = append(missing, bin.Name)
				} else {
					found = append(found, fmt.Sprintf("%s (%s)", bin.Name, path))
				}
			}

			if len(missing) == 0 {
				r.Passed = true
				r.Message = fmt.Sprintf("all required packages found: %s", strings.Join(found, ", "))
				return r
			}

			r.Message = fmt.Sprintf("missing required packages: %s", strings.Join(missing, ", "))

			// Detect package manager for install hints.
			pm, _ := detectPackageManager(deps)
			missingPkgs := missingPackageNames(missing, pm)
			needsAnsible := containsBinary(missing, "ansible-playbook")
			if pm == "apt-get" && needsAnsible {
				r.Details = "Ansible requires adding the PPA on Debian/Ubuntu:\n" +
					"  sudo apt-get install -y software-properties-common\n" +
					"  sudo add-apt-repository --yes --update ppa:ansible/ansible\n" +
					"  sudo apt-get install -y " + strings.Join(missingPkgs, " ")
			} else if pm != "" {
				r.Details = fmt.Sprintf("Install with: sudo %s install -y %s", pm, strings.Join(missingPkgs, " "))
			} else {
				r.Details = "Install missing packages using your distribution's package manager:\n" +
					"  Debian/Ubuntu: sudo add-apt-repository --yes --update ppa:ansible/ansible && sudo apt-get install -y " + strings.Join(missingPkgs, " ") + "\n" +
					"  RHEL/Fedora:   sudo dnf install -y " + strings.Join(missingPkgs, " ")
			}
			return r
		},
		RunFix: func(ctx context.Context, deps CheckDeps) FixResult {
			r := FixResult{
				ID:      "required-packages",
				Warning: "This will install system packages using the detected package manager (apt-get, dnf, or yum). On Debian/Ubuntu, the Ansible PPA is added to provide the ansible package.",
			}

			// Find which packages are missing.
			var missing []string
			for _, bin := range requiredBinaries {
				if _, err := deps.LookPath(bin.Name); err != nil {
					missing = append(missing, bin.Name)
				}
			}
			if len(missing) == 0 {
				r.Applied = true
				r.Message = "all required packages already installed"
				return r
			}

			pm, pmPath := detectPackageManager(deps)
			if pm == "" {
				r.Error = "no supported package manager found (tried apt-get, dnf, yum)"
				r.Message = "fix failed — cannot determine how to install packages"
				return r
			}

			// Ansible is not in default Ubuntu/Debian repos; add the PPA first.
			if pm == "apt-get" && containsBinary(missing, "ansible-playbook") {
				if err := addAnsiblePPA(ctx, deps); err != nil {
					r.Error = fmt.Sprintf("failed to add Ansible PPA: %v", err)
					r.Message = "fix failed — could not configure Ansible repository"
					return r
				}
			}

			pkgs := missingPackageNames(missing, pm)
			args := []string{pmPath, "install", "-y"}
			args = append(args, pkgs...)
			if _, err := deps.RunCommand(ctx, "sudo", args...); err != nil {
				r.Error = fmt.Sprintf("failed to install packages: %v", err)
				r.Message = fmt.Sprintf("fix failed — sudo %s install -y %s", pm, strings.Join(pkgs, " "))
				return r
			}

			r.Applied = true
			r.Message = fmt.Sprintf("installed packages via %s: %s", pm, strings.Join(pkgs, ", "))
			return r
		},
	}
}

// containsBinary reports whether name appears in the missing binaries list.
func containsBinary(missing []string, name string) bool {
	for _, m := range missing {
		if m == name {
			return true
		}
	}
	return false
}

// addAnsiblePPA installs software-properties-common and adds the Ansible PPA
// on Debian/Ubuntu systems. The --update flag on add-apt-repository runs
// apt update automatically after adding the PPA.
func addAnsiblePPA(ctx context.Context, deps CheckDeps) error {
	if _, err := deps.RunCommand(ctx, "sudo", "apt-get", "install", "-y", "software-properties-common"); err != nil {
		return fmt.Errorf("install software-properties-common: %w", err)
	}
	if _, err := deps.RunCommand(ctx, "sudo", "add-apt-repository", "--yes", "--update", "ppa:ansible/ansible"); err != nil {
		return fmt.Errorf("add-apt-repository: %w", err)
	}
	return nil
}

// detectPackageManager returns the name and path of the first available
// package manager, or empty strings if none is found.
func detectPackageManager(deps CheckDeps) (name, path string) {
	for _, pm := range packageManagers {
		if p, err := deps.LookPath(pm.Binary); err == nil {
			return pm.Binary, p
		}
	}
	return "", ""
}

// missingPackageNames maps binary names to distribution package names based
// on the detected package manager.
func missingPackageNames(missingBinaries []string, pm string) []string {
	isApt := pm == "apt-get"
	out := make([]string, 0, len(missingBinaries))
	for _, m := range missingBinaries {
		for _, bin := range requiredBinaries {
			if bin.Name == m {
				if isApt {
					out = append(out, bin.AptPkg)
				} else {
					out = append(out, bin.YumPkg)
				}
				break
			}
		}
	}
	return out
}

// ---------------------------------------------------------------------------
// 2. ssh-configured
// ---------------------------------------------------------------------------

func checkSSHConfigured() Check {
	return Check{
		ID:          "ssh-configured",
		Name:        "SSH Configuration",
		Description: "Checks that SSH password authentication is enabled on this host.",
		Severity:    SeverityRequired,
		Category:    CategoryAccess,
		FixWarning:  "Enabling SSH password authentication allows any user to log in with a password. Consider using key-based authentication for production environments.",
		RunCheck: func(ctx context.Context, deps CheckDeps) CheckResult {
			r := newResult("ssh-configured", "SSH Configuration",
				"Checks that SSH password authentication is enabled on this host.",
				SeverityRequired, CategoryAccess, true)
			r.FixWarning = "Enabling SSH password authentication allows any user to log in with a password. Consider using key-based authentication for production environments."

			enabled, source, err := parseSSHDPasswordAuth(deps)
			if err != nil {
				r.Error = err.Error()
				r.Message = "failed to read sshd configuration"
				return r
			}

			if enabled {
				r.Passed = true
				r.Message = fmt.Sprintf("PasswordAuthentication is enabled (set in %s)", source)
			} else {
				r.Message = fmt.Sprintf("PasswordAuthentication is disabled (set in %s)", source)
				r.Details = "SSH password authentication must be enabled for Aether node provisioning."
			}
			return r
		},
		RunFix: func(ctx context.Context, deps CheckDeps) FixResult {
			r := FixResult{
				ID:      "ssh-configured",
				Warning: "Enabling SSH password authentication allows any user to log in with a password. Consider using key-based authentication for production environments.",
			}

			dropIn := "/etc/ssh/sshd_config.d/99-aether-password-auth.conf"
			content := "PasswordAuthentication yes"
			cmd := fmt.Sprintf("echo '%s' | sudo tee %s > /dev/null", content, dropIn)
			if _, err := deps.RunCommand(ctx, "bash", "-c", cmd); err != nil {
				r.Error = fmt.Sprintf("failed to write drop-in config: %v", err)
				r.Message = "fix failed"
				return r
			}

			// Try both unit names: RHEL/Fedora use "sshd", Debian/Ubuntu use "ssh".
			restarted := false
			for _, unit := range []string{"sshd", "ssh"} {
				if _, err := deps.RunCommand(ctx, "sudo", "systemctl", "restart", unit); err == nil {
					restarted = true
					break
				}
			}
			if !restarted {
				r.Error = "wrote config but failed to restart SSH service (tried sshd and ssh units)"
				r.Message = "partial fix — config written but sshd not restarted"
				return r
			}

			r.Applied = true
			r.Message = fmt.Sprintf("wrote %s and restarted sshd", dropIn)
			return r
		},
	}
}

// parseSSHDPasswordAuth reads the main sshd_config and any drop-in files to
// determine the effective PasswordAuthentication setting. Returns (enabled, source_file, error).
// If no directive is found, defaults to yes (OpenSSH default) with source "default".
func parseSSHDPasswordAuth(deps CheckDeps) (bool, string, error) {
	const mainConfig = "/etc/ssh/sshd_config"

	// Start with OpenSSH default.
	enabled := true
	source := "default"

	// Parse main config.
	data, err := deps.ReadFile(mainConfig)
	if err != nil {
		return false, "", fmt.Errorf("read %s: %w", mainConfig, err)
	}

	if val, found := findSSHDirective(data, "PasswordAuthentication"); found {
		enabled = strings.EqualFold(val, "yes")
		source = mainConfig
	}

	// Parse drop-in files (last directive wins, matching sshd behavior).
	const dropInDir = "/etc/ssh/sshd_config.d"
	entries, err := deps.ReadDir(dropInDir)
	if err != nil {
		// Drop-in directory may not exist; that's fine.
		return enabled, source, nil
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".conf") {
			continue
		}
		path := filepath.Join(dropInDir, entry.Name())
		dropInData, err := deps.ReadFile(path)
		if err != nil {
			continue
		}
		if val, found := findSSHDirective(dropInData, "PasswordAuthentication"); found {
			enabled = strings.EqualFold(val, "yes")
			source = path
		}
	}

	return enabled, source, nil
}

// findSSHDirective scans sshd_config content for the last occurrence of a
// directive (case-insensitive), ignoring comments and blank lines.
func findSSHDirective(data []byte, directive string) (string, bool) {
	var lastVal string
	found := false
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 2 && strings.EqualFold(fields[0], directive) {
			lastVal = fields[1]
			found = true
		}
	}
	return lastVal, found
}

// ---------------------------------------------------------------------------
// 3. aether-user-configured
// ---------------------------------------------------------------------------

func checkAetherUserConfigured() Check {
	return Check{
		ID:          "aether-user-configured",
		Name:        "Aether User",
		Description: "Checks that the 'aether' system user exists with passwordless sudo.",
		Severity:    SeverityRequired,
		Category:    CategoryAccess,
		FixWarning:  "This will create a user 'aether' with a default password of 'aether' and NOPASSWD sudo access. Change the password after initial setup.",
		RunCheck: func(ctx context.Context, deps CheckDeps) CheckResult {
			r := newResult("aether-user-configured", "Aether User",
				"Checks that the 'aether' system user exists with passwordless sudo.",
				SeverityRequired, CategoryAccess, true)
			r.FixWarning = "This will create a user 'aether' with a default password of 'aether' and NOPASSWD sudo access. Change the password after initial setup."

			u, err := deps.LookupUser("aether")
			if err != nil {
				r.Message = "user 'aether' does not exist"
				return r
			}

			r.Message = fmt.Sprintf("user 'aether' exists (uid=%s)", u.Uid)

			// Check for sudoers file.
			const sudoersPath = "/etc/sudoers.d/aether"
			if _, err := deps.ReadFile(sudoersPath); err != nil {
				r.Details = fmt.Sprintf("user exists but %s not found — passwordless sudo may not be configured", sudoersPath)
				return r
			}

			r.Passed = true
			r.Message = fmt.Sprintf("user 'aether' exists (uid=%s) with sudoers configured", u.Uid)
			return r
		},
		RunFix: func(ctx context.Context, deps CheckDeps) FixResult {
			r := FixResult{
				ID:      "aether-user-configured",
				Warning: "This will create a user 'aether' with a default password of 'aether' and NOPASSWD sudo access. Change the password after initial setup.",
			}

			// Check if user already exists before attempting creation.
			_, lookupErr := deps.LookupUser("aether")
			userExisted := lookupErr == nil

			if !userExisted {
				if _, err := deps.RunCommand(ctx, "sudo", "useradd", "-m", "-s", "/bin/bash", "aether"); err != nil {
					r.Error = fmt.Sprintf("failed to create user: %v", err)
					r.Message = "fix failed"
					return r
				}

				// Set default password only for newly created users.
				cmd := "echo 'aether:aether' | sudo chpasswd"
				if _, err := deps.RunCommand(ctx, "bash", "-c", cmd); err != nil {
					r.Error = fmt.Sprintf("failed to set password: %v", err)
					r.Message = "user created but password not set"
					return r
				}
			}

			// Write sudoers file.
			sudoersContent := "aether ALL=(ALL) NOPASSWD: ALL"
			sudoersCmd := fmt.Sprintf("echo '%s' | sudo tee /etc/sudoers.d/aether > /dev/null", sudoersContent)
			if _, err := deps.RunCommand(ctx, "bash", "-c", sudoersCmd); err != nil {
				r.Error = fmt.Sprintf("failed to write sudoers file: %v", err)
				r.Message = "user created but sudoers not configured"
				return r
			}

			// Set permissions.
			if _, err := deps.RunCommand(ctx, "sudo", "chmod", "0440", "/etc/sudoers.d/aether"); err != nil {
				r.Error = fmt.Sprintf("failed to set sudoers permissions: %v", err)
				r.Message = "user and sudoers created but permissions not set"
				return r
			}

			r.Applied = true
			if userExisted {
				r.Message = "configured sudo access for existing user 'aether'"
			} else {
				r.Message = "created user 'aether' with sudo access and default password"
			}
			return r
		},
	}
}

// ---------------------------------------------------------------------------
// 4. node-ssh-reachable
// ---------------------------------------------------------------------------

func checkNodeSSHReachable() Check {
	return Check{
		ID:          "node-ssh-reachable",
		Name:        "Node SSH Reachability",
		Description: "Checks that all managed nodes are reachable via SSH (TCP port 22).",
		Severity:    SeverityInfo,
		Category:    CategoryNetwork,
		RunCheck: func(ctx context.Context, deps CheckDeps) CheckResult {
			r := newResult("node-ssh-reachable", "Node SSH Reachability",
				"Checks that all managed nodes are reachable via SSH (TCP port 22).",
				SeverityInfo, CategoryNetwork, false)

			// Empty path means no store was configured.
			if deps.Store.Path() == "" {
				r.Passed = true
				r.Message = "no store configured — skipping node reachability check"
				return r
			}

			nodes, err := deps.Store.ListNodes(ctx)
			if err != nil {
				r.Error = fmt.Sprintf("failed to list nodes: %v", err)
				r.Message = "unable to query node list"
				return r
			}

			if len(nodes) == 0 {
				r.Passed = true
				r.Message = "no managed nodes registered"
				return r
			}

			const dialTimeout = 5 * time.Second
			var details []string
			var reachable, unreachable, skipped int

			for _, node := range nodes {
				host := node.AnsibleHost
				if host == "" {
					details = append(details, fmt.Sprintf("  %s: SKIP (no ansible_host)", node.Name))
					skipped++
					continue
				}

				addr := host + ":22"
				conn, err := deps.DialTimeout("tcp", addr, dialTimeout)
				if err != nil {
					details = append(details, fmt.Sprintf("  %s (%s): UNREACHABLE — %v", node.Name, addr, err))
					unreachable++
				} else {
					conn.Close()
					details = append(details, fmt.Sprintf("  %s (%s): OK", node.Name, addr))
					reachable++
				}
			}

			r.Details = strings.Join(details, "\n")
			checked := reachable + unreachable
			if unreachable == 0 && checked > 0 {
				r.Passed = true
				if skipped > 0 {
					r.Message = fmt.Sprintf("%d node(s) reachable on port 22 (%d skipped, no ansible_host)", reachable, skipped)
				} else {
					r.Message = fmt.Sprintf("all %d node(s) reachable on port 22", reachable)
				}
			} else if checked == 0 {
				r.Passed = true
				r.Message = fmt.Sprintf("all %d node(s) skipped (no ansible_host configured)", skipped)
			} else {
				r.Message = fmt.Sprintf("%d of %d node(s) unreachable on port 22", unreachable, checked)
			}
			return r
		},
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func newResult(id, name, description string, severity Severity, category Category, canFix bool) CheckResult {
	return CheckResult{
		ID:          id,
		Name:        name,
		Description: description,
		Severity:    severity,
		Category:    category,
		CanFix:      canFix,
	}
}
