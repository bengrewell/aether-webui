package store

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// Client wraps a Store and Codec for typed object persistence.
type Client struct {
	s    Store
	c    Codec
	path string
}

// New opens (or creates) a SQLite-backed store at path and returns a
// ready-to-use Client. The parent directory is created automatically.
func New(ctx context.Context, path string, opts ...Option) (Client, error) {
	cfg := defaults()
	for _, o := range opts {
		o(&cfg)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return Client{}, fmt.Errorf("store: create directory: %w", err)
	}
	d, err := openDB(ctx, dbConfig{
		Path:          path,
		BusyTimeout:   cfg.busyTimeout,
		Crypter:       cfg.crypter,
		MetricsMaxAge: cfg.metricsMaxAge,
	})
	if err != nil {
		return Client{}, err
	}
	if err := d.Migrate(ctx); err != nil {
		d.Close()
		return Client{}, fmt.Errorf("store: migrate: %w", err)
	}
	return Client{s: d, c: JSONCodec{}, path: path}, nil
}

// Path returns the filesystem path of the backing database.
func (c Client) Path() string { return c.path }

// Close closes the underlying database connection.
func (c Client) Close() error {
	return c.s.Close()
}

// Health runs a lightweight liveness check against the store.
func (c Client) Health(ctx context.Context) error {
	return c.s.Health(ctx)
}

// Delete removes the object stored under key.
func (c Client) Delete(ctx context.Context, key Key) error {
	return c.s.Delete(ctx, key)
}

func Save[T any](c Client, ctx context.Context, key Key, v T, opts ...SaveOption) (Meta, error) {
	b, err := c.c.Marshal(v)
	if err != nil {
		return Meta{}, err
	}
	return c.s.Save(ctx, key, b, opts...)
}

func Load[T any](c Client, ctx context.Context, key Key, opts ...LoadOption) (Item[T], bool, error) {
	raw, ok, err := c.s.Load(ctx, key, opts...)
	if err != nil || !ok {
		return Item[T]{}, ok, err
	}

	var out T
	if err := c.c.Unmarshal(raw.Data, &out); err != nil {
		return Item[T]{}, false, err
	}

	return Item[T]{
		Key:  raw.Key,
		Meta: raw.Meta,
		Data: out,
	}, true, nil
}

// UpsertNode creates or updates a node in the store.
func (c Client) UpsertNode(ctx context.Context, node Node) error {
	return c.s.UpsertNode(ctx, node)
}

// GetNode retrieves a node by ID, returning decrypted secrets.
func (c Client) GetNode(ctx context.Context, id string) (Node, bool, error) {
	return c.s.GetNode(ctx, id)
}

// DeleteNode removes a node and its role assignments.
func (c Client) DeleteNode(ctx context.Context, id string) error {
	return c.s.DeleteNode(ctx, id)
}

// ListNodes returns all nodes without secrets.
func (c Client) ListNodes(ctx context.Context) ([]NodeInfo, error) {
	return c.s.ListNodes(ctx)
}

// InsertAction records a new action execution in the action history.
func (c Client) InsertAction(ctx context.Context, rec ActionRecord) error {
	return c.s.InsertAction(ctx, rec)
}

// UpdateActionResult updates the result of a completed action.
func (c Client) UpdateActionResult(ctx context.Context, id string, result ActionResult) error {
	return c.s.UpdateActionResult(ctx, id, result)
}

// GetAction retrieves a single action record by ID.
func (c Client) GetAction(ctx context.Context, id string) (ActionRecord, bool, error) {
	return c.s.GetAction(ctx, id)
}

// ListActions returns action records matching the filter criteria.
func (c Client) ListActions(ctx context.Context, filter ActionFilter) ([]ActionRecord, error) {
	return c.s.ListActions(ctx, filter)
}

// UpsertComponentState creates or updates the state of a component.
func (c Client) UpsertComponentState(ctx context.Context, cs ComponentState) error {
	return c.s.UpsertComponentState(ctx, cs)
}

// GetComponentState retrieves the state of a single component.
func (c Client) GetComponentState(ctx context.Context, component string) (ComponentState, bool, error) {
	return c.s.GetComponentState(ctx, component)
}

// ListComponentStates returns the state of all tracked components.
func (c Client) ListComponentStates(ctx context.Context) ([]ComponentState, error) {
	return c.s.ListComponentStates(ctx)
}

func (c Client) GetSchemaVersion() (int, error) {
	return c.s.GetSchemaVersion()
}

// AppendSamples writes metric samples to the time-series store.
func (c Client) AppendSamples(ctx context.Context, samples []Sample) error {
	return c.s.AppendSamples(ctx, samples)
}

// QueryRange queries metric samples over a time range.
func (c Client) QueryRange(ctx context.Context, q RangeQuery) ([]Series, error) {
	return c.s.QueryRange(ctx, q)
}
