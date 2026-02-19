package endpoint

import "context"

type Semantics uint8

const (
	Read Semantics = iota
	Create
	Update
	Delete
	Action
)

type HTTPHint struct {
	Method string
	Path   string
}

type GRPCHint struct {
	Service string
	Method  string
}

type Descriptor struct {
	OperationID string    `json:"operation_id" yaml:"operation_id" mapstructure:"operation_id"`
	Semantics   Semantics `json:"semantic" yaml:"semantic" mapstructure:"semantic"`
	Summary     string    `json:"summary" yaml:"summary" mapstructure:"summary"`
	Description string    `json:"description" yaml:"description" mapstructure:"description"`
	Tags        []string  `json:"tags" yaml:"tags" mapstructure:"tags"`

	HTTP HTTPHint `json:"http" yaml:"http" mapstructure:"http"`
	GRPC GRPCHint `json:"grpc" yaml:"grpc" mapstructure:"grpc"`
}

// Transport-neutral “any endpoint”
type AnyEndpoint interface {
	Descriptor() Descriptor
}

type Endpoint[I any, O any] struct {
	Desc    Descriptor
	Handler func(ctx context.Context, in *I) (*O, error)
}

func (e Endpoint[I, O]) Descriptor() Descriptor { return e.Desc }
