package store

import "time"

type SaveOptions struct {
	CreateOnly      bool
	ExpectedVersion int64
	TTL             time.Duration
}

type LoadOptions struct {
	RequireFresh bool
}

type ListOptions struct {
	Prefix string // optional: only IDs starting with Prefix
	Limit  int
}

type SaveOption func(*SaveOptions)
type LoadOption func(*LoadOptions)
type ListOption func(*ListOptions)

func CreateOnly() SaveOption {
	return func(o *SaveOptions) { o.CreateOnly = true }
}

func ExpectedVersion(v int64) SaveOption {
	return func(o *SaveOptions) { o.ExpectedVersion = v }
}

func WithTTL(d time.Duration) SaveOption {
	return func(o *SaveOptions) { o.TTL = d }
}

func RequireFresh() LoadOption {
	return func(o *LoadOptions) { o.RequireFresh = true }
}

func WithPrefix(p string) ListOption {
	return func(o *ListOptions) { o.Prefix = p }
}

func WithLimit(n int) ListOption {
	return func(o *ListOptions) { o.Limit = n }
}
