package provider

import (
	"context"
	"maps"

	"github.com/bengrewell/aether-webui/internal/operator"
)

// LocalProvider implements Provider for the local node.
type LocalProvider struct {
	id        NodeID
	operators map[operator.Domain]operator.Operator
}

// Option is a functional option for configuring a LocalProvider.
type Option func(*LocalProvider)

// NewLocalProvider creates a new LocalProvider with the given options.
func NewLocalProvider(opts ...Option) *LocalProvider {
	p := &LocalProvider{
		id:        LocalNode,
		operators: make(map[operator.Domain]operator.Operator),
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// WithOperator adds an operator to the provider.
func WithOperator(op operator.Operator) Option {
	return func(p *LocalProvider) {
		p.operators[op.Domain()] = op
	}
}

// ID returns the provider's node identifier.
func (p *LocalProvider) ID() NodeID {
	return p.id
}

// Operator returns the operator for the given domain, or nil if not available.
func (p *LocalProvider) Operator(domain operator.Domain) operator.Operator {
	return p.operators[domain]
}

// Operators returns all registered operators.
func (p *LocalProvider) Operators() map[operator.Domain]operator.Operator {
	result := make(map[operator.Domain]operator.Operator, len(p.operators))
	maps.Copy(result, p.operators)
	return result
}

// Health returns the aggregated health status of all operators.
func (p *LocalProvider) Health(ctx context.Context) (*ProviderHealth, error) {
	health := &ProviderHealth{
		Status:    "healthy",
		Message:   "all operators healthy",
		Operators: make(map[operator.Domain]string),
	}

	degraded := false
	unavailable := 0

	for domain, op := range p.operators {
		opHealth, err := op.Health(ctx)
		if err != nil {
			health.Operators[domain] = "error"
			unavailable++
			continue
		}
		health.Operators[domain] = opHealth.Status
		switch opHealth.Status {
		case "unavailable":
			unavailable++
		case "degraded":
			degraded = true
		}
	}

	if unavailable == len(p.operators) && len(p.operators) > 0 {
		health.Status = "unavailable"
		health.Message = "all operators unavailable"
	} else if unavailable > 0 || degraded {
		health.Status = "degraded"
		health.Message = "some operators degraded or unavailable"
	}

	return health, nil
}

// IsLocal returns true since this is a local provider.
func (p *LocalProvider) IsLocal() bool {
	return true
}
