package preflight

import (
	"context"
	"errors"
	"io/fs"
	"net"
	"os/user"
	"strings"
	"testing"
	"time"

	"github.com/bengrewell/aether-webui/internal/provider"
	"github.com/bengrewell/aether-webui/internal/store"
)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

// testDeps returns CheckDeps with all functions stubbed to fail by default.
func testDeps(t *testing.T) CheckDeps {
	t.Helper()
	errStub := errors.New("not stubbed")
	return CheckDeps{
		Log: nil,
		LookPath: func(string) (string, error) {
			return "", errStub
		},
		ReadFile: func(string) ([]byte, error) {
			return nil, errStub
		},
		ReadDir: func(string) ([]fs.DirEntry, error) {
			return nil, errStub
		},
		LookupUser: func(string) (*user.User, error) {
			return nil, errStub
		},
		RunCommand: func(ctx context.Context, name string, args ...string) ([]byte, error) {
			return nil, errStub
		},
		DialTimeout: func(string, string, time.Duration) (net.Conn, error) {
			return nil, errStub
		},
	}
}

func newTestProvider(t *testing.T) *Preflight {
	t.Helper()
	return NewProvider()
}

func newTestProviderWithStore(t *testing.T) *Preflight {
	t.Helper()
	ctx := t.Context()
	dbPath := t.TempDir() + "/test.db"
	st, err := store.New(ctx, dbPath)
	if err != nil {
		t.Fatalf("store.New: %v", err)
	}
	t.Cleanup(func() { st.Close() })
	return NewProvider(provider.WithStore(st))
}

// ---------------------------------------------------------------------------
// Constructor / registration tests
// ---------------------------------------------------------------------------

func TestNewProvider_ImplementsInterface(t *testing.T) {
	var _ provider.Provider = newTestProvider(t)
}

func TestNewProvider_EndpointCount(t *testing.T) {
	p := newTestProvider(t)
	descs := p.Base.Descriptors()
	if len(descs) != 3 {
		t.Errorf("registered %d endpoints, want 3", len(descs))
	}
}

func TestNewProvider_EndpointPaths(t *testing.T) {
	p := newTestProvider(t)

	wantOps := map[string]string{
		"preflight-list": "/api/v1/preflight",
		"preflight-get":  "/api/v1/preflight/{id}",
		"preflight-fix":  "/api/v1/preflight/{id}/fix",
	}

	descs := p.Base.Descriptors()
	for _, d := range descs {
		want, ok := wantOps[d.OperationID]
		if !ok {
			t.Errorf("unexpected operation %q", d.OperationID)
			continue
		}
		if d.HTTP.Path != want {
			t.Errorf("operation %q path = %q, want %q", d.OperationID, d.HTTP.Path, want)
		}
		delete(wantOps, d.OperationID)
	}
	for op := range wantOps {
		t.Errorf("missing operation %q", op)
	}
}

// ---------------------------------------------------------------------------
// Registry integrity tests
// ---------------------------------------------------------------------------

func TestRegistry_NoDuplicateIDs(t *testing.T) {
	seen := make(map[string]bool)
	for _, c := range registry {
		if seen[c.ID] {
			t.Errorf("duplicate check ID: %q", c.ID)
		}
		seen[c.ID] = true
	}
}

func TestRegistry_AllHaveRunCheck(t *testing.T) {
	for _, c := range registry {
		if c.RunCheck == nil {
			t.Errorf("check %q has nil RunCheck", c.ID)
		}
	}
}

func TestRegistry_IndexCoversAll(t *testing.T) {
	if len(checkIndex) != len(registry) {
		t.Errorf("checkIndex has %d entries, registry has %d", len(checkIndex), len(registry))
	}
	for _, c := range registry {
		if _, ok := checkIndex[c.ID]; !ok {
			t.Errorf("check %q missing from checkIndex", c.ID)
		}
	}
}

// ---------------------------------------------------------------------------
// required-packages check
// ---------------------------------------------------------------------------

func TestCheckRequiredPackages_AllFound(t *testing.T) {
	deps := testDeps(t)
	deps.LookPath = func(name string) (string, error) {
		switch name {
		case "make":
			return "/usr/bin/make", nil
		case "ansible-playbook":
			return "/usr/bin/ansible-playbook", nil
		}
		return "", errors.New("not found")
	}

	check := checkRequiredPackages()
	r := check.RunCheck(t.Context(), deps)

	if !r.Passed {
		t.Errorf("expected Passed=true, message=%q", r.Message)
	}
	if !strings.Contains(r.Message, "make") || !strings.Contains(r.Message, "ansible-playbook") {
		t.Errorf("message = %q, expected both package names", r.Message)
	}
}

func TestCheckRequiredPackages_SomeMissing(t *testing.T) {
	deps := testDeps(t)
	deps.LookPath = func(name string) (string, error) {
		if name == "make" {
			return "/usr/bin/make", nil
		}
		return "", errors.New("not found")
	}

	check := checkRequiredPackages()
	r := check.RunCheck(t.Context(), deps)

	if r.Passed {
		t.Error("expected Passed=false")
	}
	if !strings.Contains(r.Message, "ansible-playbook") {
		t.Errorf("message = %q, expected missing package name", r.Message)
	}
	if r.Details == "" {
		t.Error("expected install instructions in Details")
	}
}

func TestCheckRequiredPackages_AllMissing(t *testing.T) {
	deps := testDeps(t)
	// LookPath defaults to error.

	check := checkRequiredPackages()
	r := check.RunCheck(t.Context(), deps)

	if r.Passed {
		t.Error("expected Passed=false")
	}
	if !strings.Contains(r.Message, "make") || !strings.Contains(r.Message, "ansible-playbook") {
		t.Errorf("message = %q, expected both missing package names", r.Message)
	}
	if !r.CanFix {
		t.Error("expected CanFix=true")
	}
}

func TestCheckRequiredPackages_FixWithApt(t *testing.T) {
	deps := testDeps(t)
	// make is missing, ansible-playbook is present.
	deps.LookPath = func(name string) (string, error) {
		switch name {
		case "ansible-playbook":
			return "/usr/bin/ansible-playbook", nil
		case "apt-get":
			return "/usr/bin/apt-get", nil
		}
		return "", errors.New("not found")
	}
	var ranCmd string
	deps.RunCommand = func(ctx context.Context, name string, args ...string) ([]byte, error) {
		ranCmd = name + " " + strings.Join(args, " ")
		return []byte("ok"), nil
	}

	check := checkRequiredPackages()
	r := check.RunFix(t.Context(), deps)

	if !r.Applied {
		t.Errorf("expected Applied=true, error=%q, message=%q", r.Error, r.Message)
	}
	if !strings.Contains(ranCmd, "apt-get") {
		t.Errorf("expected apt-get in command, got %q", ranCmd)
	}
	if !strings.Contains(ranCmd, "make") {
		t.Errorf("expected 'make' package in command, got %q", ranCmd)
	}
}

func TestCheckRequiredPackages_FixWithDnf(t *testing.T) {
	deps := testDeps(t)
	deps.LookPath = func(name string) (string, error) {
		if name == "dnf" {
			return "/usr/bin/dnf", nil
		}
		return "", errors.New("not found")
	}
	var ranCmd string
	deps.RunCommand = func(ctx context.Context, name string, args ...string) ([]byte, error) {
		ranCmd = name + " " + strings.Join(args, " ")
		return []byte("ok"), nil
	}

	check := checkRequiredPackages()
	r := check.RunFix(t.Context(), deps)

	if !r.Applied {
		t.Errorf("expected Applied=true, error=%q, message=%q", r.Error, r.Message)
	}
	if !strings.Contains(ranCmd, "dnf") {
		t.Errorf("expected dnf in command, got %q", ranCmd)
	}
}

func TestCheckRequiredPackages_FixNoPkgManager(t *testing.T) {
	deps := testDeps(t)
	// All LookPath calls fail (no binaries, no package manager).

	check := checkRequiredPackages()
	r := check.RunFix(t.Context(), deps)

	if r.Applied {
		t.Error("expected Applied=false when no package manager")
	}
	if r.Error == "" {
		t.Error("expected Error to be set")
	}
}

func TestCheckRequiredPackages_FixAllPresent(t *testing.T) {
	deps := testDeps(t)
	deps.LookPath = func(name string) (string, error) {
		return "/usr/bin/" + name, nil
	}

	check := checkRequiredPackages()
	r := check.RunFix(t.Context(), deps)

	if !r.Applied {
		t.Errorf("expected Applied=true when all present, error=%q", r.Error)
	}
	if !strings.Contains(r.Message, "already installed") {
		t.Errorf("message = %q, expected 'already installed'", r.Message)
	}
}

// ---------------------------------------------------------------------------
// ssh-configured check
// ---------------------------------------------------------------------------

func TestCheckSSHConfigured_Enabled(t *testing.T) {
	deps := testDeps(t)
	deps.ReadFile = func(path string) ([]byte, error) {
		if path == "/etc/ssh/sshd_config" {
			return []byte("PasswordAuthentication yes\n"), nil
		}
		return nil, errors.New("not found")
	}
	deps.ReadDir = func(string) ([]fs.DirEntry, error) {
		return nil, errors.New("no dir")
	}

	check := checkSSHConfigured()
	r := check.RunCheck(t.Context(), deps)

	if !r.Passed {
		t.Errorf("expected Passed=true, message=%q", r.Message)
	}
	if !r.CanFix {
		t.Error("expected CanFix=true")
	}
}

func TestCheckSSHConfigured_Disabled(t *testing.T) {
	deps := testDeps(t)
	deps.ReadFile = func(path string) ([]byte, error) {
		if path == "/etc/ssh/sshd_config" {
			return []byte("PasswordAuthentication no\n"), nil
		}
		return nil, errors.New("not found")
	}
	deps.ReadDir = func(string) ([]fs.DirEntry, error) {
		return nil, errors.New("no dir")
	}

	check := checkSSHConfigured()
	r := check.RunCheck(t.Context(), deps)

	if r.Passed {
		t.Error("expected Passed=false")
	}
	if !strings.Contains(r.Message, "disabled") {
		t.Errorf("message = %q, expected 'disabled'", r.Message)
	}
}

func TestCheckSSHConfigured_DefaultNoDirective(t *testing.T) {
	deps := testDeps(t)
	deps.ReadFile = func(path string) ([]byte, error) {
		if path == "/etc/ssh/sshd_config" {
			return []byte("# Some comment\nPort 22\n"), nil
		}
		return nil, errors.New("not found")
	}
	deps.ReadDir = func(string) ([]fs.DirEntry, error) {
		return nil, errors.New("no dir")
	}

	check := checkSSHConfigured()
	r := check.RunCheck(t.Context(), deps)

	// OpenSSH default is yes.
	if !r.Passed {
		t.Errorf("expected Passed=true (OpenSSH default), message=%q", r.Message)
	}
}

func TestCheckSSHConfigured_ReadError(t *testing.T) {
	deps := testDeps(t)
	// ReadFile already returns error by default from testDeps.
	deps.ReadDir = func(string) ([]fs.DirEntry, error) {
		return nil, errors.New("no dir")
	}

	check := checkSSHConfigured()
	r := check.RunCheck(t.Context(), deps)

	if r.Passed {
		t.Error("expected Passed=false on read error")
	}
	if r.Error == "" {
		t.Error("expected Error to be set")
	}
}

func TestCheckSSHConfigured_DropInOverride(t *testing.T) {
	deps := testDeps(t)
	deps.ReadFile = func(path string) ([]byte, error) {
		switch path {
		case "/etc/ssh/sshd_config":
			return []byte("PasswordAuthentication yes\n"), nil
		case "/etc/ssh/sshd_config.d/50-cloud-init.conf":
			return []byte("PasswordAuthentication no\n"), nil
		default:
			return nil, errors.New("not found")
		}
	}
	deps.ReadDir = func(dir string) ([]fs.DirEntry, error) {
		if dir == "/etc/ssh/sshd_config.d" {
			return []fs.DirEntry{fakeDirEntry{name: "50-cloud-init.conf"}}, nil
		}
		return nil, errors.New("no dir")
	}

	check := checkSSHConfigured()
	r := check.RunCheck(t.Context(), deps)

	if r.Passed {
		t.Error("expected Passed=false (drop-in overrides main config)")
	}
	if !strings.Contains(r.Message, "50-cloud-init.conf") {
		t.Errorf("message = %q, expected drop-in file mention", r.Message)
	}
}

// ---------------------------------------------------------------------------
// findSSHDirective parser tests
// ---------------------------------------------------------------------------

func TestFindSSHDirective(t *testing.T) {
	tests := []struct {
		name      string
		data      string
		directive string
		wantVal   string
		wantFound bool
	}{
		{
			name:      "simple yes",
			data:      "PasswordAuthentication yes",
			directive: "PasswordAuthentication",
			wantVal:   "yes",
			wantFound: true,
		},
		{
			name:      "simple no",
			data:      "PasswordAuthentication no",
			directive: "PasswordAuthentication",
			wantVal:   "no",
			wantFound: true,
		},
		{
			name:      "commented out",
			data:      "#PasswordAuthentication yes",
			directive: "PasswordAuthentication",
			wantVal:   "",
			wantFound: false,
		},
		{
			name:      "case insensitive",
			data:      "passwordauthentication YES",
			directive: "PasswordAuthentication",
			wantVal:   "YES",
			wantFound: true,
		},
		{
			name:      "last directive wins",
			data:      "PasswordAuthentication yes\nPasswordAuthentication no",
			directive: "PasswordAuthentication",
			wantVal:   "no",
			wantFound: true,
		},
		{
			name:      "not found",
			data:      "Port 22\nUsePAM yes",
			directive: "PasswordAuthentication",
			wantVal:   "",
			wantFound: false,
		},
		{
			name:      "empty file",
			data:      "",
			directive: "PasswordAuthentication",
			wantVal:   "",
			wantFound: false,
		},
		{
			name:      "leading whitespace",
			data:      "  PasswordAuthentication yes",
			directive: "PasswordAuthentication",
			wantVal:   "yes",
			wantFound: true,
		},
		{
			name:      "tabs",
			data:      "\tPasswordAuthentication\tyes",
			directive: "PasswordAuthentication",
			wantVal:   "yes",
			wantFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, found := findSSHDirective([]byte(tt.data), tt.directive)
			if found != tt.wantFound {
				t.Errorf("found = %v, want %v", found, tt.wantFound)
			}
			if val != tt.wantVal {
				t.Errorf("val = %q, want %q", val, tt.wantVal)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// aether-user-configured check
// ---------------------------------------------------------------------------

func TestCheckAetherUserConfigured_ExistsWithSudoers(t *testing.T) {
	deps := testDeps(t)
	deps.LookupUser = func(name string) (*user.User, error) {
		if name == "aether" {
			return &user.User{Uid: "1001", Username: "aether"}, nil
		}
		return nil, errors.New("not found")
	}
	deps.ReadFile = func(path string) ([]byte, error) {
		if path == "/etc/sudoers.d/aether" {
			return []byte("aether ALL=(ALL) NOPASSWD: ALL\n"), nil
		}
		return nil, errors.New("not found")
	}

	check := checkAetherUserConfigured()
	r := check.RunCheck(t.Context(), deps)

	if !r.Passed {
		t.Errorf("expected Passed=true, message=%q", r.Message)
	}
	if !strings.Contains(r.Message, "1001") {
		t.Errorf("message = %q, expected uid mention", r.Message)
	}
}

func TestCheckAetherUserConfigured_ExistsWithoutSudoers(t *testing.T) {
	deps := testDeps(t)
	deps.LookupUser = func(name string) (*user.User, error) {
		if name == "aether" {
			return &user.User{Uid: "1001", Username: "aether"}, nil
		}
		return nil, errors.New("not found")
	}
	// ReadFile defaults to error for sudoers path.

	check := checkAetherUserConfigured()
	r := check.RunCheck(t.Context(), deps)

	if r.Passed {
		t.Error("expected Passed=false (no sudoers)")
	}
	if r.Details == "" {
		t.Error("expected Details about missing sudoers")
	}
}

func TestCheckAetherUserConfigured_Missing(t *testing.T) {
	deps := testDeps(t)
	// LookupUser defaults to error.

	check := checkAetherUserConfigured()
	r := check.RunCheck(t.Context(), deps)

	if r.Passed {
		t.Error("expected Passed=false")
	}
	if !strings.Contains(r.Message, "does not exist") {
		t.Errorf("message = %q, expected 'does not exist'", r.Message)
	}
}

// ---------------------------------------------------------------------------
// node-ssh-reachable check
// ---------------------------------------------------------------------------

func TestCheckNodeSSHReachable_NoStore(t *testing.T) {
	deps := testDeps(t)
	// Store is zero-value (empty path).

	check := checkNodeSSHReachable()
	r := check.RunCheck(t.Context(), deps)

	if !r.Passed {
		t.Error("expected Passed=true when no store")
	}
	if !strings.Contains(r.Message, "no store") {
		t.Errorf("message = %q, expected 'no store'", r.Message)
	}
}

func TestCheckNodeSSHReachable_NoNodes(t *testing.T) {
	p := newTestProviderWithStore(t)
	deps := DefaultDeps(p.Store(), p.Log())

	check := checkNodeSSHReachable()
	r := check.RunCheck(t.Context(), deps)

	if !r.Passed {
		t.Error("expected Passed=true when no nodes")
	}
	if !strings.Contains(r.Message, "no managed nodes") {
		t.Errorf("message = %q, expected 'no managed nodes'", r.Message)
	}
}

func TestCheckNodeSSHReachable_AllReachable(t *testing.T) {
	p := newTestProviderWithStore(t)
	ctx := t.Context()

	// Create a node.
	if err := p.Store().UpsertNode(ctx, store.Node{
		ID:          "n1",
		Name:        "node1",
		AnsibleHost: "10.0.0.1",
	}); err != nil {
		t.Fatalf("UpsertNode: %v", err)
	}

	deps := DefaultDeps(p.Store(), p.Log())
	deps.DialTimeout = func(network, addr string, timeout time.Duration) (net.Conn, error) {
		return &fakeConn{}, nil
	}

	check := checkNodeSSHReachable()
	r := check.RunCheck(ctx, deps)

	if !r.Passed {
		t.Errorf("expected Passed=true, message=%q", r.Message)
	}
}

func TestCheckNodeSSHReachable_SomeUnreachable(t *testing.T) {
	p := newTestProviderWithStore(t)
	ctx := t.Context()

	for _, n := range []store.Node{
		{ID: "n1", Name: "node1", AnsibleHost: "10.0.0.1"},
		{ID: "n2", Name: "node2", AnsibleHost: "10.0.0.2"},
	} {
		if err := p.Store().UpsertNode(ctx, n); err != nil {
			t.Fatalf("UpsertNode: %v", err)
		}
	}

	deps := DefaultDeps(p.Store(), p.Log())
	deps.DialTimeout = func(network, addr string, timeout time.Duration) (net.Conn, error) {
		if addr == "10.0.0.1:22" {
			return &fakeConn{}, nil
		}
		return nil, errors.New("connection refused")
	}

	check := checkNodeSSHReachable()
	r := check.RunCheck(ctx, deps)

	if r.Passed {
		t.Error("expected Passed=false with unreachable nodes")
	}
	if !strings.Contains(r.Details, "UNREACHABLE") {
		t.Errorf("details = %q, expected 'UNREACHABLE'", r.Details)
	}
}

// ---------------------------------------------------------------------------
// Fix function tests
// ---------------------------------------------------------------------------

func TestFixAetherUser_Success(t *testing.T) {
	deps := testDeps(t)
	deps.LookupUser = func(name string) (*user.User, error) {
		return nil, errors.New("not found")
	}
	deps.RunCommand = func(ctx context.Context, name string, args ...string) ([]byte, error) {
		return []byte("ok"), nil
	}

	check := checkAetherUserConfigured()
	r := check.RunFix(t.Context(), deps)

	if !r.Applied {
		t.Errorf("expected Applied=true, error=%q, message=%q", r.Error, r.Message)
	}
}

func TestFixAetherUser_CreateFails(t *testing.T) {
	deps := testDeps(t)
	deps.LookupUser = func(name string) (*user.User, error) {
		return nil, errors.New("not found")
	}
	deps.RunCommand = func(ctx context.Context, name string, args ...string) ([]byte, error) {
		return nil, errors.New("permission denied")
	}

	check := checkAetherUserConfigured()
	r := check.RunFix(t.Context(), deps)

	if r.Applied {
		t.Error("expected Applied=false")
	}
	if r.Error == "" {
		t.Error("expected Error to be set")
	}
}

func TestFixSSHConfigured_Success(t *testing.T) {
	deps := testDeps(t)
	deps.RunCommand = func(ctx context.Context, name string, args ...string) ([]byte, error) {
		return []byte("ok"), nil
	}

	check := checkSSHConfigured()
	r := check.RunFix(t.Context(), deps)

	if !r.Applied {
		t.Errorf("expected Applied=true, error=%q, message=%q", r.Error, r.Message)
	}
}

func TestFixSSHConfigured_RestartFails(t *testing.T) {
	deps := testDeps(t)
	deps.RunCommand = func(ctx context.Context, name string, args ...string) ([]byte, error) {
		// First call (tee) succeeds, second call (systemctl restart) fails.
		if name == "sudo" && len(args) > 0 && args[0] == "systemctl" {
			return nil, errors.New("systemctl failed")
		}
		return []byte("ok"), nil
	}

	check := checkSSHConfigured()
	r := check.RunFix(t.Context(), deps)

	if r.Applied {
		t.Error("expected Applied=false on restart failure")
	}
	if !strings.Contains(r.Error, "restart sshd") {
		t.Errorf("error = %q, expected restart sshd mention", r.Error)
	}
}

// ---------------------------------------------------------------------------
// Handler tests
// ---------------------------------------------------------------------------

func TestHandleListChecks(t *testing.T) {
	p := newTestProvider(t)
	out, err := p.handleListChecks(t.Context(), nil)
	if err != nil {
		t.Fatalf("handleListChecks: %v", err)
	}

	if out.Body.Total != len(registry) {
		t.Errorf("Total = %d, want %d", out.Body.Total, len(registry))
	}
	if len(out.Body.Results) != len(registry) {
		t.Errorf("Results count = %d, want %d", len(out.Body.Results), len(registry))
	}
	if out.Body.Passed+out.Body.Failed != out.Body.Total {
		t.Errorf("Passed(%d) + Failed(%d) != Total(%d)", out.Body.Passed, out.Body.Failed, out.Body.Total)
	}
}

func TestHandleGetCheck_Found(t *testing.T) {
	p := newTestProvider(t)
	out, err := p.handleGetCheck(t.Context(), &PreflightGetInput{ID: "required-packages"})
	if err != nil {
		t.Fatalf("handleGetCheck: %v", err)
	}
	if out.Body.ID != "required-packages" {
		t.Errorf("ID = %q, want %q", out.Body.ID, "required-packages")
	}
}

func TestHandleGetCheck_NotFound(t *testing.T) {
	p := newTestProvider(t)
	_, err := p.handleGetCheck(t.Context(), &PreflightGetInput{ID: "nonexistent"})
	if err == nil {
		t.Fatal("expected error for nonexistent check")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %q, should mention 'not found'", err)
	}
}

func TestHandleFixCheck_HasFix(t *testing.T) {
	p := newTestProvider(t)
	// ssh-configured has a fix; it will fail (no sudo) but the handler should not error.
	out, err := p.handleFixCheck(t.Context(), &PreflightFixInput{ID: "ssh-configured"})
	if err != nil {
		t.Fatalf("handleFixCheck: %v", err)
	}
	if out.Body.ID != "ssh-configured" {
		t.Errorf("ID = %q, want %q", out.Body.ID, "ssh-configured")
	}
}

func TestHandleFixCheck_NoFix(t *testing.T) {
	p := newTestProvider(t)
	_, err := p.handleFixCheck(t.Context(), &PreflightFixInput{ID: "node-ssh-reachable"})
	if err == nil {
		t.Fatal("expected error for check without fix")
	}
	if !strings.Contains(err.Error(), "no automated fix") {
		t.Errorf("error = %q, should mention 'no automated fix'", err)
	}
}

func TestHandleFixCheck_NotFound(t *testing.T) {
	p := newTestProvider(t)
	_, err := p.handleFixCheck(t.Context(), &PreflightFixInput{ID: "nonexistent"})
	if err == nil {
		t.Fatal("expected error for nonexistent check")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %q, should mention 'not found'", err)
	}
}

// ---------------------------------------------------------------------------
// Test fakes
// ---------------------------------------------------------------------------

// fakeDirEntry implements fs.DirEntry for testing.
type fakeDirEntry struct {
	name  string
	isDir bool
}

func (f fakeDirEntry) Name() string              { return f.name }
func (f fakeDirEntry) IsDir() bool               { return f.isDir }
func (f fakeDirEntry) Type() fs.FileMode         { return 0 }
func (f fakeDirEntry) Info() (fs.FileInfo, error) { return nil, nil }

// fakeConn implements net.Conn for testing DialTimeout.
type fakeConn struct{}

func (f *fakeConn) Read([]byte) (int, error)        { return 0, nil }
func (f *fakeConn) Write([]byte) (int, error)        { return 0, nil }
func (f *fakeConn) Close() error                     { return nil }
func (f *fakeConn) LocalAddr() net.Addr              { return fakeAddr{} }
func (f *fakeConn) RemoteAddr() net.Addr             { return fakeAddr{} }
func (f *fakeConn) SetDeadline(time.Time) error      { return nil }
func (f *fakeConn) SetReadDeadline(time.Time) error  { return nil }
func (f *fakeConn) SetWriteDeadline(time.Time) error { return nil }

type fakeAddr struct{}

func (fakeAddr) Network() string { return "tcp" }
func (fakeAddr) String() string  { return "fake" }
