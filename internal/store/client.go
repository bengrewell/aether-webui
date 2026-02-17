package store

import (
	"context"
)

type Client struct {
	S Store
	C Codec
}

func Save[T any](c Client, ctx context.Context, key Key, v T, opts ...SaveOption) (Meta, error) {
	b, err := c.C.Marshal(v)
	if err != nil {
		return Meta{}, err
	}
	return c.S.Save(ctx, key, b, opts...)
}

func Load[T any](c Client, ctx context.Context, key Key, opts ...LoadOption) (Item[T], bool, error) {
	raw, ok, err := c.S.Load(ctx, key, opts...)
	if err != nil || !ok {
		return Item[T]{}, ok, err
	}

	var out T
	if err := c.C.Unmarshal(raw.Data, &out); err != nil {
		return Item[T]{}, false, err
	}

	return Item[T]{
		Key:  raw.Key,
		Meta: raw.Meta,
		Data: out,
	}, true, nil
}

func (c Client) GetSchemaVersion() (int, error) {
	return c.S.GetSchemaVersion()
}
