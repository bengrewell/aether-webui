package provider

import "testing"

func TestSetDegraded(t *testing.T) {
	b := New("test")

	b.SetDegraded("something broke")

	si := b.StatusInfo()
	if !si.Degraded {
		t.Error("expected Degraded=true after SetDegraded")
	}
	if si.DegradedReason != "something broke" {
		t.Errorf("DegradedReason = %q, want %q", si.DegradedReason, "something broke")
	}
}

func TestClearDegraded(t *testing.T) {
	b := New("test")
	b.SetDegraded("something broke")

	b.ClearDegraded()

	si := b.StatusInfo()
	if si.Degraded {
		t.Error("expected Degraded=false after ClearDegraded")
	}
	if si.DegradedReason != "" {
		t.Errorf("DegradedReason = %q, want empty", si.DegradedReason)
	}
}

func TestStatusInfo_DefaultNotDegraded(t *testing.T) {
	b := New("test")
	si := b.StatusInfo()
	if si.Degraded {
		t.Error("expected Degraded=false by default")
	}
	if si.DegradedReason != "" {
		t.Errorf("DegradedReason = %q, want empty by default", si.DegradedReason)
	}
}
