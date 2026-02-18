package store

import "encoding/json"

type Codec interface {
	Marshal(v any) ([]byte, error)
	Unmarshal(data []byte, out any) error
}

// JSONCodec provides a default Codec implementation using encoding/json.
// It is suitable for storing "objects" payloads (settings/config/state) as JSON.
type JSONCodec struct{}

func (JSONCodec) Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

func (JSONCodec) Unmarshal(data []byte, out any) error {
	return json.Unmarshal(data, out)
}
