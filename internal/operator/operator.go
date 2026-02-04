package operator

import (
	"context"
	"errors"
)

// Domain represents the operational domain of an operator.
type Domain string

const (
	DomainHost   Domain = "host"
	DomainKube   Domain = "kube"
	DomainAether Domain = "aether"
	DomainExec   Domain = "exec"
)

// Operator is the base interface for all domain-specific operators.
type Operator interface {
	Domain() Domain
	Health(ctx context.Context) (*OperatorHealth, error)
}

// OperatorHealth represents the health status of an operator.
type OperatorHealth struct {
	Status  string `json:"status"` // "healthy", "degraded", "unavailable"
	Message string `json:"message"`
}

// ErrNotImplemented is returned by stub operators for unimplemented methods.
var ErrNotImplemented = errors.New("not implemented")

// OperationType distinguishes between actions and queries.
type OperationType string

const (
	// Action represents an operation that modifies state.
	Action OperationType = "action"
	// Query represents an operation that only reads state.
	Query OperationType = "query"
)

// Operation defines a named operation that an operator can perform.
type Operation struct {
	Name        string
	Type        OperationType
	Description string
}

// Invocable is an interface for operators that support generic invocation.
type Invocable interface {
	// SupportedOperations returns the list of operations this operator supports.
	SupportedOperations() []Operation
	// Invoke executes a named operation with the given arguments.
	Invoke(ctx context.Context, opType OperationType, operation string, args ...any) (any, error)
}
