package nodes

import (
	"strings"
	"testing"

	"github.com/bengrewell/aether-webui/internal/provider"
	"github.com/bengrewell/aether-webui/internal/store"
)

func newTestProvider(t *testing.T) *Nodes {
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
	if len(descs) != 5 {
		t.Errorf("registered %d endpoints, want 5", len(descs))
	}
}

func TestNewProvider_EndpointPaths(t *testing.T) {
	p := newTestProvider(t)

	wantOps := map[string]string{
		"nodes-list":   "/api/v1/nodes",
		"nodes-get":    "/api/v1/nodes/{id}",
		"nodes-create": "/api/v1/nodes",
		"nodes-update": "/api/v1/nodes/{id}",
		"nodes-delete": "/api/v1/nodes/{id}",
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
// Create handler
// ---------------------------------------------------------------------------

func TestHandleCreate(t *testing.T) {
	p := newTestProvider(t)
	in := &NodeCreateInput{}
	in.Body.Name = "node1"
	in.Body.AnsibleHost = "10.0.0.1"
	in.Body.AnsibleUser = "ubuntu"
	in.Body.Password = "secret"
	in.Body.Roles = []string{"master"}

	out, err := p.handleCreate(t.Context(), in)
	if err != nil {
		t.Fatalf("handleCreate: %v", err)
	}
	if out.Body.ID == "" {
		t.Error("expected non-empty ID")
	}
	if out.Body.Name != "node1" {
		t.Errorf("Name = %q, want %q", out.Body.Name, "node1")
	}
	if out.Body.AnsibleHost != "10.0.0.1" {
		t.Errorf("AnsibleHost = %q, want %q", out.Body.AnsibleHost, "10.0.0.1")
	}
	if !out.Body.HasPassword {
		t.Error("expected HasPassword=true")
	}
	if out.Body.HasSudoPassword {
		t.Error("expected HasSudoPassword=false")
	}
	if len(out.Body.Roles) != 1 || out.Body.Roles[0] != "master" {
		t.Errorf("Roles = %v, want [master]", out.Body.Roles)
	}
}

func TestHandleCreate_MissingName(t *testing.T) {
	p := newTestProvider(t)
	in := &NodeCreateInput{}
	in.Body.AnsibleHost = "10.0.0.1"

	_, err := p.handleCreate(t.Context(), in)
	if err == nil {
		t.Fatal("expected error for missing name")
	}
}

func TestHandleCreate_MissingHost(t *testing.T) {
	p := newTestProvider(t)
	in := &NodeCreateInput{}
	in.Body.Name = "node1"

	_, err := p.handleCreate(t.Context(), in)
	if err == nil {
		t.Fatal("expected error for missing host")
	}
}

func TestHandleCreate_InvalidRole(t *testing.T) {
	p := newTestProvider(t)
	in := &NodeCreateInput{}
	in.Body.Name = "node1"
	in.Body.AnsibleHost = "10.0.0.1"
	in.Body.Roles = []string{"invalid_role"}

	_, err := p.handleCreate(t.Context(), in)
	if err == nil {
		t.Fatal("expected error for invalid role")
	}
	if !strings.Contains(err.Error(), "invalid role") {
		t.Errorf("error = %q, should mention 'invalid role'", err)
	}
}

// ---------------------------------------------------------------------------
// Get handler
// ---------------------------------------------------------------------------

func TestHandleGet(t *testing.T) {
	p := newTestProvider(t)

	// Create first.
	in := &NodeCreateInput{}
	in.Body.Name = "node1"
	in.Body.AnsibleHost = "10.0.0.1"
	in.Body.SSHKey = "ssh-rsa AAAA"
	created, err := p.handleCreate(t.Context(), in)
	if err != nil {
		t.Fatalf("handleCreate: %v", err)
	}

	out, err := p.handleGet(t.Context(), &NodeGetInput{ID: created.Body.ID})
	if err != nil {
		t.Fatalf("handleGet: %v", err)
	}
	if out.Body.Name != "node1" {
		t.Errorf("Name = %q, want %q", out.Body.Name, "node1")
	}
	if !out.Body.HasSSHKey {
		t.Error("expected HasSSHKey=true")
	}
}

func TestHandleGet_NotFound(t *testing.T) {
	p := newTestProvider(t)
	_, err := p.handleGet(t.Context(), &NodeGetInput{ID: "nonexistent"})
	if err == nil {
		t.Fatal("expected error for missing node")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %q, should mention 'not found'", err)
	}
}

// ---------------------------------------------------------------------------
// List handler
// ---------------------------------------------------------------------------

func TestHandleList_Empty(t *testing.T) {
	p := newTestProvider(t)
	out, err := p.handleList(t.Context(), nil)
	if err != nil {
		t.Fatalf("handleList: %v", err)
	}
	if len(out.Body) != 0 {
		t.Errorf("expected empty list, got %d", len(out.Body))
	}
}

func TestHandleList_WithNodes(t *testing.T) {
	p := newTestProvider(t)

	for _, name := range []string{"node1", "node2"} {
		in := &NodeCreateInput{}
		in.Body.Name = name
		in.Body.AnsibleHost = "10.0.0.1"
		if _, err := p.handleCreate(t.Context(), in); err != nil {
			t.Fatalf("handleCreate(%s): %v", name, err)
		}
	}

	out, err := p.handleList(t.Context(), nil)
	if err != nil {
		t.Fatalf("handleList: %v", err)
	}
	if len(out.Body) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(out.Body))
	}
}

// ---------------------------------------------------------------------------
// Update handler
// ---------------------------------------------------------------------------

func TestHandleUpdate(t *testing.T) {
	p := newTestProvider(t)

	in := &NodeCreateInput{}
	in.Body.Name = "node1"
	in.Body.AnsibleHost = "10.0.0.1"
	in.Body.Roles = []string{"master"}
	created, _ := p.handleCreate(t.Context(), in)

	newName := "node1-updated"
	newHost := "10.0.0.2"
	upd := &NodeUpdateInput{ID: created.Body.ID}
	upd.Body.Name = &newName
	upd.Body.AnsibleHost = &newHost
	upd.Body.Roles = []string{"worker"}

	out, err := p.handleUpdate(t.Context(), upd)
	if err != nil {
		t.Fatalf("handleUpdate: %v", err)
	}
	if out.Body.Name != "node1-updated" {
		t.Errorf("Name = %q, want %q", out.Body.Name, "node1-updated")
	}
	if out.Body.AnsibleHost != "10.0.0.2" {
		t.Errorf("AnsibleHost = %q, want %q", out.Body.AnsibleHost, "10.0.0.2")
	}
	if len(out.Body.Roles) != 1 || out.Body.Roles[0] != "worker" {
		t.Errorf("Roles = %v, want [worker]", out.Body.Roles)
	}
}

func TestHandleUpdate_NotFound(t *testing.T) {
	p := newTestProvider(t)
	upd := &NodeUpdateInput{ID: "nonexistent"}
	_, err := p.handleUpdate(t.Context(), upd)
	if err == nil {
		t.Fatal("expected error for missing node")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %q, should mention 'not found'", err)
	}
}

func TestHandleUpdate_PartialUpdate(t *testing.T) {
	p := newTestProvider(t)

	in := &NodeCreateInput{}
	in.Body.Name = "node1"
	in.Body.AnsibleHost = "10.0.0.1"
	in.Body.AnsibleUser = "ubuntu"
	in.Body.Roles = []string{"master"}
	created, _ := p.handleCreate(t.Context(), in)

	// Only update name; everything else should be preserved.
	newName := "node1-renamed"
	upd := &NodeUpdateInput{ID: created.Body.ID}
	upd.Body.Name = &newName

	out, err := p.handleUpdate(t.Context(), upd)
	if err != nil {
		t.Fatalf("handleUpdate: %v", err)
	}
	if out.Body.Name != "node1-renamed" {
		t.Errorf("Name = %q, want %q", out.Body.Name, "node1-renamed")
	}
	if out.Body.AnsibleHost != "10.0.0.1" {
		t.Error("AnsibleHost should be preserved")
	}
	if out.Body.AnsibleUser != "ubuntu" {
		t.Error("AnsibleUser should be preserved")
	}
	if len(out.Body.Roles) != 1 || out.Body.Roles[0] != "master" {
		t.Errorf("Roles should be preserved, got %v", out.Body.Roles)
	}
}

func TestHandleUpdate_InvalidRole(t *testing.T) {
	p := newTestProvider(t)

	in := &NodeCreateInput{}
	in.Body.Name = "node1"
	in.Body.AnsibleHost = "10.0.0.1"
	created, _ := p.handleCreate(t.Context(), in)

	upd := &NodeUpdateInput{ID: created.Body.ID}
	upd.Body.Roles = []string{"bogus"}

	_, err := p.handleUpdate(t.Context(), upd)
	if err == nil {
		t.Fatal("expected error for invalid role")
	}
}

// ---------------------------------------------------------------------------
// Delete handler
// ---------------------------------------------------------------------------

func TestHandleDelete(t *testing.T) {
	p := newTestProvider(t)

	in := &NodeCreateInput{}
	in.Body.Name = "node1"
	in.Body.AnsibleHost = "10.0.0.1"
	created, _ := p.handleCreate(t.Context(), in)

	out, err := p.handleDelete(t.Context(), &NodeDeleteInput{ID: created.Body.ID})
	if err != nil {
		t.Fatalf("handleDelete: %v", err)
	}
	if !strings.Contains(out.Body.Message, "deleted") {
		t.Errorf("message = %q, should mention 'deleted'", out.Body.Message)
	}

	// Verify node is gone.
	_, err = p.handleGet(t.Context(), &NodeGetInput{ID: created.Body.ID})
	if err == nil {
		t.Error("expected error getting deleted node")
	}
}

// ---------------------------------------------------------------------------
// Role validation
// ---------------------------------------------------------------------------

func TestValidateRoles_AllValid(t *testing.T) {
	for role := range ValidRoles {
		if err := validateRoles([]string{role}); err != nil {
			t.Errorf("validateRoles(%q): %v", role, err)
		}
	}
}

func TestValidateRoles_Invalid(t *testing.T) {
	if err := validateRoles([]string{"invalid"}); err == nil {
		t.Error("expected error for invalid role")
	}
}

func TestValidateRoles_Empty(t *testing.T) {
	if err := validateRoles(nil); err != nil {
		t.Errorf("validateRoles(nil): %v", err)
	}
}

// ---------------------------------------------------------------------------
// generateID
// ---------------------------------------------------------------------------

func TestGenerateID_Unique(t *testing.T) {
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id, err := generateID()
		if err != nil {
			t.Fatalf("generateID: %v", err)
		}
		if ids[id] {
			t.Fatalf("duplicate ID: %s", id)
		}
		ids[id] = true
	}
}
