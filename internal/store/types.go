package store

import "time"

type Key struct {
	Namespace string
	ID        string
}

type Meta struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	Version   int64
	ExpiresAt *time.Time
}

type Item[T any] struct {
	Key  Key
	Meta Meta
	Data T
}

// Credentials

type Credential struct {
	ID        string            // stable identifier
	Provider  string            // e.g. "ssh", "k8s", "aws"
	Labels    map[string]string // optional metadata tags (not used for auth)
	Secret    []byte            // plaintext at API boundary; encrypted at rest
	UpdatedAt time.Time         // optional; store will set if zero
}

type CredentialInfo struct {
	ID        string
	Provider  string
	Labels    map[string]string
	UpdatedAt time.Time
}

// Nodes

type Node struct {
	ID           string
	Name         string   // Ansible inventory hostname (e.g. "node1")
	AnsibleHost  string   // IP or hostname for SSH
	AnsibleUser  string   // SSH username
	Password     []byte   // plaintext at API boundary; encrypted at rest
	SudoPassword []byte   // plaintext at API boundary; encrypted at rest
	SSHKey       []byte   // plaintext at API boundary; encrypted at rest
	Roles        []string // role assignments (master, worker, gnbsim, etc.)
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type NodeInfo struct {
	ID          string
	Name        string
	AnsibleHost string
	AnsibleUser string
	Roles       []string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Metrics

type Sample struct {
	Metric string
	TS     time.Time
	Value  float64
	Labels map[string]string // exact-match identity for now
	Unit   string
}

type TimeRange struct {
	From time.Time
	To   time.Time
}

type Agg uint8

const (
	AggRaw Agg = iota
	Agg10s
	Agg1m
	Agg5m
	Agg1h
)

type RangeQuery struct {
	Metric string
	Range  TimeRange

	// Exact match only in this minimal implementation.
	LabelsExact map[string]string

	// Reserved for later; ignored for now.
	GroupBy []string

	Agg       Agg
	MaxPoints int
}

type Point struct {
	TS    time.Time
	Value float64
}

type Series struct {
	Metric string
	Labels map[string]string
	Points []Point
}
