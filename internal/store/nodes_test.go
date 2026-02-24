package store

import (
	"testing"
)

func newTestStore(t *testing.T) Client {
	t.Helper()
	ctx := t.Context()
	dbPath := t.TempDir() + "/test.db"
	st, err := New(ctx, dbPath)
	if err != nil {
		t.Fatalf("store.New: %v", err)
	}
	t.Cleanup(func() { st.Close() })
	return st
}

func TestUpsertNode_RoundTrip(t *testing.T) {
	st := newTestStore(t)
	ctx := t.Context()

	node := Node{
		ID:           "node-1",
		Name:         "master1",
		AnsibleHost:  "10.0.0.1",
		AnsibleUser:  "ubuntu",
		Password:     []byte("secret"),
		SudoPassword: []byte("sudopass"),
		SSHKey:       []byte("ssh-rsa AAAA"),
		Roles:        []string{"master", "worker"},
	}

	if err := st.UpsertNode(ctx, node); err != nil {
		t.Fatalf("UpsertNode: %v", err)
	}

	got, ok, err := st.GetNode(ctx, "node-1")
	if err != nil {
		t.Fatalf("GetNode: %v", err)
	}
	if !ok {
		t.Fatal("expected node to exist")
	}

	if got.Name != "master1" {
		t.Errorf("Name = %q, want %q", got.Name, "master1")
	}
	if got.AnsibleHost != "10.0.0.1" {
		t.Errorf("AnsibleHost = %q, want %q", got.AnsibleHost, "10.0.0.1")
	}
	if got.AnsibleUser != "ubuntu" {
		t.Errorf("AnsibleUser = %q, want %q", got.AnsibleUser, "ubuntu")
	}
	if string(got.Password) != "secret" {
		t.Errorf("Password = %q, want %q", got.Password, "secret")
	}
	if string(got.SudoPassword) != "sudopass" {
		t.Errorf("SudoPassword = %q, want %q", got.SudoPassword, "sudopass")
	}
	if string(got.SSHKey) != "ssh-rsa AAAA" {
		t.Errorf("SSHKey = %q, want %q", got.SSHKey, "ssh-rsa AAAA")
	}
	if len(got.Roles) != 2 {
		t.Fatalf("Roles = %v, want 2 elements", got.Roles)
	}
	// Roles are sorted by nodeRoles query.
	if got.Roles[0] != "master" || got.Roles[1] != "worker" {
		t.Errorf("Roles = %v, want [master worker]", got.Roles)
	}
	if got.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
	if got.UpdatedAt.IsZero() {
		t.Error("expected non-zero UpdatedAt")
	}
}

func TestUpsertNode_Update(t *testing.T) {
	st := newTestStore(t)
	ctx := t.Context()

	node := Node{
		ID:          "node-1",
		Name:        "master1",
		AnsibleHost: "10.0.0.1",
		AnsibleUser: "ubuntu",
		Roles:       []string{"master"},
	}
	if err := st.UpsertNode(ctx, node); err != nil {
		t.Fatalf("UpsertNode: %v", err)
	}

	// Update name and roles.
	node.Name = "master1-updated"
	node.AnsibleHost = "10.0.0.2"
	node.Roles = []string{"worker"}
	if err := st.UpsertNode(ctx, node); err != nil {
		t.Fatalf("UpsertNode (update): %v", err)
	}

	got, ok, err := st.GetNode(ctx, "node-1")
	if err != nil {
		t.Fatalf("GetNode: %v", err)
	}
	if !ok {
		t.Fatal("expected node to exist")
	}
	if got.Name != "master1-updated" {
		t.Errorf("Name = %q, want %q", got.Name, "master1-updated")
	}
	if got.AnsibleHost != "10.0.0.2" {
		t.Errorf("AnsibleHost = %q, want %q", got.AnsibleHost, "10.0.0.2")
	}
	if len(got.Roles) != 1 || got.Roles[0] != "worker" {
		t.Errorf("Roles = %v, want [worker]", got.Roles)
	}
}

func TestUpsertNode_Validation(t *testing.T) {
	st := newTestStore(t)
	ctx := t.Context()

	tests := []struct {
		name string
		node Node
	}{
		{"empty ID", Node{Name: "n", AnsibleHost: "h"}},
		{"empty Name", Node{ID: "id", AnsibleHost: "h"}},
		{"empty AnsibleHost", Node{ID: "id", Name: "n"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := st.UpsertNode(ctx, tt.node)
			if err != ErrInvalidArgument {
				t.Errorf("expected ErrInvalidArgument, got %v", err)
			}
		})
	}
}

func TestGetNode_NotFound(t *testing.T) {
	st := newTestStore(t)
	ctx := t.Context()

	_, ok, err := st.GetNode(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("GetNode: %v", err)
	}
	if ok {
		t.Error("expected ok=false for missing node")
	}
}

func TestGetNode_EmptyID(t *testing.T) {
	st := newTestStore(t)
	ctx := t.Context()

	_, _, err := st.GetNode(ctx, "")
	if err != ErrInvalidArgument {
		t.Errorf("expected ErrInvalidArgument, got %v", err)
	}
}

func TestDeleteNode(t *testing.T) {
	st := newTestStore(t)
	ctx := t.Context()

	node := Node{
		ID:          "node-del",
		Name:        "todelete",
		AnsibleHost: "10.0.0.99",
		Roles:       []string{"master"},
	}
	if err := st.UpsertNode(ctx, node); err != nil {
		t.Fatalf("UpsertNode: %v", err)
	}

	if err := st.DeleteNode(ctx, "node-del"); err != nil {
		t.Fatalf("DeleteNode: %v", err)
	}

	_, ok, err := st.GetNode(ctx, "node-del")
	if err != nil {
		t.Fatalf("GetNode after delete: %v", err)
	}
	if ok {
		t.Error("expected node to be deleted")
	}
}

func TestDeleteNode_CascadeRoles(t *testing.T) {
	st := newTestStore(t)
	ctx := t.Context()

	node := Node{
		ID:          "node-cascade",
		Name:        "cascademe",
		AnsibleHost: "10.0.0.50",
		Roles:       []string{"master", "worker"},
	}
	if err := st.UpsertNode(ctx, node); err != nil {
		t.Fatalf("UpsertNode: %v", err)
	}

	if err := st.DeleteNode(ctx, "node-cascade"); err != nil {
		t.Fatalf("DeleteNode: %v", err)
	}

	// Roles should be gone (no orphans). Insert the same node again with no roles
	// and verify it has no leftover roles.
	node2 := Node{
		ID:          "node-cascade",
		Name:        "cascademe",
		AnsibleHost: "10.0.0.50",
	}
	if err := st.UpsertNode(ctx, node2); err != nil {
		t.Fatalf("UpsertNode (reinsert): %v", err)
	}
	got, _, _ := st.GetNode(ctx, "node-cascade")
	if len(got.Roles) != 0 {
		t.Errorf("expected no roles after cascade delete, got %v", got.Roles)
	}
}

func TestDeleteNode_EmptyID(t *testing.T) {
	st := newTestStore(t)
	ctx := t.Context()

	if err := st.DeleteNode(ctx, ""); err != ErrInvalidArgument {
		t.Errorf("expected ErrInvalidArgument, got %v", err)
	}
}

func TestListNodes(t *testing.T) {
	st := newTestStore(t)
	ctx := t.Context()

	// Start empty.
	list, err := st.ListNodes(ctx)
	if err != nil {
		t.Fatalf("ListNodes: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("expected empty list, got %d", len(list))
	}

	// Add two nodes.
	for _, n := range []Node{
		{ID: "n1", Name: "beta", AnsibleHost: "10.0.0.2", Roles: []string{"worker"}},
		{ID: "n2", Name: "alpha", AnsibleHost: "10.0.0.1", Roles: []string{"master"}},
	} {
		if err := st.UpsertNode(ctx, n); err != nil {
			t.Fatalf("UpsertNode(%s): %v", n.ID, err)
		}
	}

	list, err = st.ListNodes(ctx)
	if err != nil {
		t.Fatalf("ListNodes: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(list))
	}

	// Ordered by name.
	if list[0].Name != "alpha" {
		t.Errorf("first node = %q, want %q", list[0].Name, "alpha")
	}
	if list[1].Name != "beta" {
		t.Errorf("second node = %q, want %q", list[1].Name, "beta")
	}

	// Roles populated.
	if len(list[0].Roles) != 1 || list[0].Roles[0] != "master" {
		t.Errorf("alpha roles = %v, want [master]", list[0].Roles)
	}
}

func TestUpsertNode_NilSecrets(t *testing.T) {
	st := newTestStore(t)
	ctx := t.Context()

	node := Node{
		ID:          "node-nosec",
		Name:        "nosecrets",
		AnsibleHost: "10.0.0.3",
	}
	if err := st.UpsertNode(ctx, node); err != nil {
		t.Fatalf("UpsertNode: %v", err)
	}

	got, ok, _ := st.GetNode(ctx, "node-nosec")
	if !ok {
		t.Fatal("expected node to exist")
	}
	if got.Password != nil {
		t.Errorf("Password = %v, want nil", got.Password)
	}
	if got.SudoPassword != nil {
		t.Errorf("SudoPassword = %v, want nil", got.SudoPassword)
	}
	if got.SSHKey != nil {
		t.Errorf("SSHKey = %v, want nil", got.SSHKey)
	}
}

func TestMigrationCount(t *testing.T) {
	st := newTestStore(t)
	ctx := t.Context()

	// Verify 3 migrations were applied by counting rows.
	var count int
	err := st.s.(*db).conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM schema_migrations`).Scan(&count)
	if err != nil {
		t.Fatalf("count migrations: %v", err)
	}
	if count != 3 {
		t.Errorf("migration count = %d, want 3", count)
	}
}
