package nodes

import "time"

// ValidRoles are the Ansible inventory group names corresponding to hosts.ini sections.
var ValidRoles = map[string]bool{
	"master":  true,
	"worker":  true,
	"gnbsim":  true,
	"oai":     true,
	"ueransim": true,
	"srsran":  true,
	"oscric":  true,
	"n3iwf":   true,
}

// ManagedNode is the API-facing representation of a cluster node.
// Secrets are never returned; only boolean presence flags are exposed.
type ManagedNode struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	AnsibleHost     string    `json:"ansible_host"`
	AnsibleUser     string    `json:"ansible_user"`
	HasPassword     bool      `json:"has_password"`
	HasSudoPassword bool      `json:"has_sudo_password"`
	HasSSHKey       bool      `json:"has_ssh_key"`
	Roles           []string  `json:"roles"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// ---------------------------------------------------------------------------
// Huma I/O types
// ---------------------------------------------------------------------------

type ManagedNodeListOutput struct {
	Body []ManagedNode
}

type NodeGetInput struct {
	ID string `path:"id" doc:"Node ID"`
}

type NodeGetOutput struct {
	Body ManagedNode
}

type NodeCreateInput struct {
	Body struct {
		Name         string   `json:"name" doc:"Unique node name (Ansible inventory hostname)"`
		AnsibleHost  string   `json:"ansible_host" doc:"IP or hostname for SSH"`
		AnsibleUser  string   `json:"ansible_user,omitempty" doc:"SSH username"`
		Password     string   `json:"password,omitempty" doc:"SSH password"`
		SudoPassword string   `json:"sudo_password,omitempty" doc:"Sudo password"`
		SSHKey       string   `json:"ssh_key,omitempty" doc:"SSH private key"`
		Roles        []string `json:"roles,omitempty" doc:"Role assignments"`
	}
}

type NodeCreateOutput struct {
	Body ManagedNode
}

type NodeUpdateInput struct {
	ID   string `path:"id" doc:"Node ID"`
	Body struct {
		Name         *string  `json:"name,omitempty" doc:"Unique node name"`
		AnsibleHost  *string  `json:"ansible_host,omitempty" doc:"IP or hostname for SSH"`
		AnsibleUser  *string  `json:"ansible_user,omitempty" doc:"SSH username"`
		Password     *string  `json:"password,omitempty" doc:"SSH password (set to empty string to clear)"`
		SudoPassword *string  `json:"sudo_password,omitempty" doc:"Sudo password (set to empty string to clear)"`
		SSHKey       *string  `json:"ssh_key,omitempty" doc:"SSH private key (set to empty string to clear)"`
		Roles        []string `json:"roles,omitempty" doc:"Role assignments (replaces entire set)"`
	}
}

type NodeUpdateOutput struct {
	Body ManagedNode
}

type NodeDeleteInput struct {
	ID string `path:"id" doc:"Node ID"`
}

type NodeDeleteOutput struct {
	Body struct {
		Message string `json:"message"`
	}
}
