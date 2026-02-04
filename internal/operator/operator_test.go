package operator

import (
	"errors"
	"testing"
)

func TestDomainConstants(t *testing.T) {
	tests := []struct {
		domain Domain
		want   string
	}{
		{DomainHost, "host"},
		{DomainKube, "kube"},
		{DomainAether, "aether"},
		{DomainExec, "exec"},
	}
	for _, tc := range tests {
		if string(tc.domain) != tc.want {
			t.Errorf("Domain %q: got %q, want %q", tc.domain, string(tc.domain), tc.want)
		}
	}
}

func TestOperationTypeConstants(t *testing.T) {
	tests := []struct {
		opType OperationType
		want   string
	}{
		{Action, "action"},
		{Query, "query"},
	}
	for _, tc := range tests {
		if string(tc.opType) != tc.want {
			t.Errorf("OperationType %q: got %q, want %q", tc.opType, string(tc.opType), tc.want)
		}
	}
}

func TestErrNotImplemented(t *testing.T) {
	if ErrNotImplemented == nil {
		t.Fatal("ErrNotImplemented should not be nil")
	}
	if ErrNotImplemented.Error() != "not implemented" {
		t.Errorf("ErrNotImplemented.Error() = %q, want %q", ErrNotImplemented.Error(), "not implemented")
	}
}

func TestErrNotImplementedIsComparable(t *testing.T) {
	err := ErrNotImplemented
	if !errors.Is(err, ErrNotImplemented) {
		t.Error("errors.Is(ErrNotImplemented, ErrNotImplemented) should be true")
	}
}

func TestOperatorHealthStruct(t *testing.T) {
	health := &OperatorHealth{
		Status:  "healthy",
		Message: "test message",
	}
	if health.Status != "healthy" {
		t.Errorf("Status = %q, want %q", health.Status, "healthy")
	}
	if health.Message != "test message" {
		t.Errorf("Message = %q, want %q", health.Message, "test message")
	}
}

func TestOperationStruct(t *testing.T) {
	op := Operation{
		Name:        "test-op",
		Type:        Action,
		Description: "A test operation",
	}
	if op.Name != "test-op" {
		t.Errorf("Name = %q, want %q", op.Name, "test-op")
	}
	if op.Type != Action {
		t.Errorf("Type = %q, want %q", op.Type, Action)
	}
	if op.Description != "A test operation" {
		t.Errorf("Description = %q, want %q", op.Description, "A test operation")
	}
}
