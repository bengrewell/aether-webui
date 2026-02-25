package store

import "context"

type Store interface {
	Close() error
	Health(ctx context.Context) error
	Migrate(ctx context.Context) error

	// Objects (generic payload)
	Save(ctx context.Context, key Key, payload []byte, opts ...SaveOption) (Meta, error)
	Load(ctx context.Context, key Key, opts ...LoadOption) (ItemBytes, bool, error)
	Delete(ctx context.Context, key Key) error
	List(ctx context.Context, namespace string, opts ...ListOption) ([]Key, error)

	// Credentials (typed)
	UpsertCredential(ctx context.Context, cred Credential) error
	GetCredential(ctx context.Context, id string) (Credential, bool, error)
	DeleteCredential(ctx context.Context, id string) error
	ListCredentials(ctx context.Context) ([]CredentialInfo, error)

	// Nodes (typed)
	UpsertNode(ctx context.Context, node Node) error
	GetNode(ctx context.Context, id string) (Node, bool, error)
	DeleteNode(ctx context.Context, id string) error
	ListNodes(ctx context.Context) ([]NodeInfo, error)

	// Actions
	InsertAction(ctx context.Context, rec ActionRecord) error
	UpdateActionResult(ctx context.Context, id string, result ActionResult) error
	GetAction(ctx context.Context, id string) (ActionRecord, bool, error)
	ListActions(ctx context.Context, filter ActionFilter) ([]ActionRecord, error)

	// Component state
	UpsertComponentState(ctx context.Context, cs ComponentState) error
	GetComponentState(ctx context.Context, component string) (ComponentState, bool, error)
	ListComponentStates(ctx context.Context) ([]ComponentState, error)

	// Metrics (typed)
	AppendSample(ctx context.Context, s Sample) error
	AppendSamples(ctx context.Context, samples []Sample) error
	QueryRange(ctx context.Context, q RangeQuery) ([]Series, error)

	// Schema
	GetSchemaVersion() (int, error)

	CompactMetrics(ctx context.Context) error
}

type ItemBytes struct {
	Key  Key
	Meta Meta
	Data []byte
}
