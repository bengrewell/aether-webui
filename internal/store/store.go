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
