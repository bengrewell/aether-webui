package provider

import (
	"context"
	"testing"

	"github.com/bengrewell/aether-webui/internal/operator"
)

// mockOperator is a configurable mock for testing.
type mockOperator struct {
	domain    operator.Domain
	health    *operator.OperatorHealth
	healthErr error
}

func (m *mockOperator) Domain() operator.Domain {
	return m.domain
}

func (m *mockOperator) Health(ctx context.Context) (*operator.OperatorHealth, error) {
	return m.health, m.healthErr
}

func TestNewLocalProvider(t *testing.T) {
	p := NewLocalProvider()
	if p == nil {
		t.Fatal("NewLocalProvider() returned nil")
	}
	if p.ID() != LocalNode {
		t.Errorf("ID() = %q, want %q", p.ID(), LocalNode)
	}
}

func TestNewLocalProviderWithOperator(t *testing.T) {
	op := &mockOperator{domain: operator.DomainHost}
	p := NewLocalProvider(WithOperator(op))

	got := p.Operator(operator.DomainHost)
	if got != op {
		t.Errorf("Operator(DomainHost) = %v, want %v", got, op)
	}
}

func TestNewLocalProviderWithMultipleOperators(t *testing.T) {
	hostOp := &mockOperator{domain: operator.DomainHost}
	kubeOp := &mockOperator{domain: operator.DomainKube}
	aetherOp := &mockOperator{domain: operator.DomainAether}

	p := NewLocalProvider(
		WithOperator(hostOp),
		WithOperator(kubeOp),
		WithOperator(aetherOp),
	)

	if p.Operator(operator.DomainHost) != hostOp {
		t.Error("Operator(DomainHost) returned wrong operator")
	}
	if p.Operator(operator.DomainKube) != kubeOp {
		t.Error("Operator(DomainKube) returned wrong operator")
	}
	if p.Operator(operator.DomainAether) != aetherOp {
		t.Error("Operator(DomainAether) returned wrong operator")
	}
}

func TestLocalProviderOperatorNotFound(t *testing.T) {
	p := NewLocalProvider()
	got := p.Operator(operator.DomainHost)
	if got != nil {
		t.Errorf("Operator(DomainHost) = %v, want nil", got)
	}
}

func TestLocalProviderOperatorsReturnsDefensiveCopy(t *testing.T) {
	op := &mockOperator{domain: operator.DomainHost}
	p := NewLocalProvider(WithOperator(op))

	ops1 := p.Operators()
	ops2 := p.Operators()

	// Modify the first copy
	delete(ops1, operator.DomainHost)

	// The second copy and provider should be unaffected
	if len(ops2) != 1 {
		t.Errorf("ops2 length = %d, want 1", len(ops2))
	}
	if p.Operator(operator.DomainHost) != op {
		t.Error("modifying returned map affected provider's internal state")
	}
}

func TestLocalProviderIsLocal(t *testing.T) {
	p := NewLocalProvider()
	if !p.IsLocal() {
		t.Error("IsLocal() = false, want true")
	}
}

func TestLocalProviderHealthAllHealthy(t *testing.T) {
	hostOp := &mockOperator{
		domain: operator.DomainHost,
		health: &operator.OperatorHealth{Status: "healthy", Message: "ok"},
	}
	kubeOp := &mockOperator{
		domain: operator.DomainKube,
		health: &operator.OperatorHealth{Status: "healthy", Message: "ok"},
	}

	p := NewLocalProvider(WithOperator(hostOp), WithOperator(kubeOp))
	health, err := p.Health(context.Background())

	if err != nil {
		t.Fatalf("Health() error = %v", err)
	}
	if health.Status != "healthy" {
		t.Errorf("Status = %q, want %q", health.Status, "healthy")
	}
	if health.Message != "all operators healthy" {
		t.Errorf("Message = %q, want %q", health.Message, "all operators healthy")
	}
}

func TestLocalProviderHealthSomeDegraded(t *testing.T) {
	hostOp := &mockOperator{
		domain: operator.DomainHost,
		health: &operator.OperatorHealth{Status: "healthy", Message: "ok"},
	}
	kubeOp := &mockOperator{
		domain: operator.DomainKube,
		health: &operator.OperatorHealth{Status: "degraded", Message: "partially available"},
	}

	p := NewLocalProvider(WithOperator(hostOp), WithOperator(kubeOp))
	health, err := p.Health(context.Background())

	if err != nil {
		t.Fatalf("Health() error = %v", err)
	}
	if health.Status != "degraded" {
		t.Errorf("Status = %q, want %q", health.Status, "degraded")
	}
	if health.Operators[operator.DomainHost] != "healthy" {
		t.Errorf("Operators[DomainHost] = %q, want %q", health.Operators[operator.DomainHost], "healthy")
	}
	if health.Operators[operator.DomainKube] != "degraded" {
		t.Errorf("Operators[DomainKube] = %q, want %q", health.Operators[operator.DomainKube], "degraded")
	}
}

func TestLocalProviderHealthSomeUnavailable(t *testing.T) {
	hostOp := &mockOperator{
		domain: operator.DomainHost,
		health: &operator.OperatorHealth{Status: "healthy", Message: "ok"},
	}
	kubeOp := &mockOperator{
		domain: operator.DomainKube,
		health: &operator.OperatorHealth{Status: "unavailable", Message: "not running"},
	}

	p := NewLocalProvider(WithOperator(hostOp), WithOperator(kubeOp))
	health, err := p.Health(context.Background())

	if err != nil {
		t.Fatalf("Health() error = %v", err)
	}
	if health.Status != "degraded" {
		t.Errorf("Status = %q, want %q", health.Status, "degraded")
	}
}

func TestLocalProviderHealthAllUnavailable(t *testing.T) {
	hostOp := &mockOperator{
		domain: operator.DomainHost,
		health: &operator.OperatorHealth{Status: "unavailable", Message: "down"},
	}
	kubeOp := &mockOperator{
		domain: operator.DomainKube,
		health: &operator.OperatorHealth{Status: "unavailable", Message: "down"},
	}

	p := NewLocalProvider(WithOperator(hostOp), WithOperator(kubeOp))
	health, err := p.Health(context.Background())

	if err != nil {
		t.Fatalf("Health() error = %v", err)
	}
	if health.Status != "unavailable" {
		t.Errorf("Status = %q, want %q", health.Status, "unavailable")
	}
	if health.Message != "all operators unavailable" {
		t.Errorf("Message = %q, want %q", health.Message, "all operators unavailable")
	}
}

func TestLocalProviderHealthOperatorError(t *testing.T) {
	hostOp := &mockOperator{
		domain:    operator.DomainHost,
		health:    nil,
		healthErr: operator.ErrNotImplemented,
	}
	kubeOp := &mockOperator{
		domain: operator.DomainKube,
		health: &operator.OperatorHealth{Status: "healthy", Message: "ok"},
	}

	p := NewLocalProvider(WithOperator(hostOp), WithOperator(kubeOp))
	health, err := p.Health(context.Background())

	if err != nil {
		t.Fatalf("Health() error = %v", err)
	}
	if health.Status != "degraded" {
		t.Errorf("Status = %q, want %q", health.Status, "degraded")
	}
	if health.Operators[operator.DomainHost] != "error" {
		t.Errorf("Operators[DomainHost] = %q, want %q", health.Operators[operator.DomainHost], "error")
	}
}

func TestLocalProviderHealthNoOperators(t *testing.T) {
	p := NewLocalProvider()
	health, err := p.Health(context.Background())

	if err != nil {
		t.Fatalf("Health() error = %v", err)
	}
	if health.Status != "healthy" {
		t.Errorf("Status = %q, want %q", health.Status, "healthy")
	}
}

func TestLocalProviderImplementsInterface(t *testing.T) {
	var _ Provider = (*LocalProvider)(nil)
}
