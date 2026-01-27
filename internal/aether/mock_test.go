package aether

import (
	"context"
	"testing"
)

func TestNewMockProvider(t *testing.T) {
	provider := NewMockProvider("local")
	if provider == nil {
		t.Fatal("NewMockProvider returned nil")
	}
	if provider.Host() != "local" {
		t.Errorf("Expected host 'local', got '%s'", provider.Host())
	}
}

func TestNewMockProviderDefaultHost(t *testing.T) {
	provider := NewMockProvider("")
	if provider.Host() != "local" {
		t.Errorf("Expected default host 'local', got '%s'", provider.Host())
	}
}

func TestMockProvider_GetCoreConfig(t *testing.T) {
	provider := NewMockProvider("local")
	ctx := context.Background()

	config, err := provider.GetCoreConfig(ctx)
	if err != nil {
		t.Fatalf("GetCoreConfig returned error: %v", err)
	}
	if config == nil {
		t.Fatal("GetCoreConfig returned nil")
	}

	// Verify default values match expected
	if !config.Standalone {
		t.Error("Expected Standalone to be true")
	}
	if config.DataIface == "" {
		t.Error("DataIface is empty")
	}
	if config.Helm.ChartRef == "" {
		t.Error("Helm.ChartRef is empty")
	}
	if config.Helm.ChartVersion == "" {
		t.Error("Helm.ChartVersion is empty")
	}
	if config.UPF.AccessSubnet == "" {
		t.Error("UPF.AccessSubnet is empty")
	}
	if config.AMF.IP == "" {
		t.Error("AMF.IP is empty")
	}
}

func TestMockProvider_UpdateCoreConfig(t *testing.T) {
	provider := NewMockProvider("local")
	ctx := context.Background()

	newConfig := &CoreConfig{
		Standalone: false,
		DataIface:  "eth0",
		Helm: HelmConfig{
			ChartRef:     "custom/chart",
			ChartVersion: "1.0.0",
		},
	}

	err := provider.UpdateCoreConfig(ctx, newConfig)
	if err != nil {
		t.Fatalf("UpdateCoreConfig returned error: %v", err)
	}

	// Verify update took effect
	config, _ := provider.GetCoreConfig(ctx)
	if config.Standalone != false {
		t.Error("Standalone should be false after update")
	}
	if config.DataIface != "eth0" {
		t.Errorf("DataIface should be 'eth0', got '%s'", config.DataIface)
	}
}

func TestMockProvider_DeployCore(t *testing.T) {
	provider := NewMockProvider("local")
	ctx := context.Background()

	resp, err := provider.DeployCore(ctx)
	if err != nil {
		t.Fatalf("DeployCore returned error: %v", err)
	}
	if resp == nil {
		t.Fatal("DeployCore returned nil response")
	}
	if !resp.Success {
		t.Error("Expected Success to be true")
	}
	if resp.TaskID == "" {
		t.Error("TaskID is empty")
	}

	// Verify status changed
	status, _ := provider.GetCoreStatus(ctx)
	if status.State != StateDeployed {
		t.Errorf("Expected state Deployed, got %s", status.State)
	}
	if status.Host != "local" {
		t.Errorf("Expected host 'local', got '%s'", status.Host)
	}
}

func TestMockProvider_UndeployCore(t *testing.T) {
	provider := NewMockProvider("local")
	ctx := context.Background()

	resp, err := provider.UndeployCore(ctx)
	if err != nil {
		t.Fatalf("UndeployCore returned error: %v", err)
	}
	if !resp.Success {
		t.Error("Expected Success to be true")
	}

	// Verify status changed
	status, _ := provider.GetCoreStatus(ctx)
	if status.State != StateNotDeployed {
		t.Errorf("Expected state NotDeployed, got %s", status.State)
	}
}

func TestMockProvider_GetCoreStatus(t *testing.T) {
	provider := NewMockProvider("local")
	ctx := context.Background()

	status, err := provider.GetCoreStatus(ctx)
	if err != nil {
		t.Fatalf("GetCoreStatus returned error: %v", err)
	}
	if status == nil {
		t.Fatal("GetCoreStatus returned nil")
	}

	validStates := map[DeploymentState]bool{
		StateNotDeployed: true,
		StateDeploying:   true,
		StateDeployed:    true,
		StateFailed:      true,
		StateUndeploying: true,
	}
	if !validStates[status.State] {
		t.Errorf("Invalid state: %s", status.State)
	}
	if status.Host != "local" {
		t.Errorf("Expected host 'local', got '%s'", status.Host)
	}
}

func TestMockProvider_ListGNBs(t *testing.T) {
	provider := NewMockProvider("local")
	ctx := context.Background()

	list, err := provider.ListGNBs(ctx)
	if err != nil {
		t.Fatalf("ListGNBs returned error: %v", err)
	}
	if list == nil {
		t.Fatal("ListGNBs returned nil")
	}
	if len(list.GNBs) == 0 {
		t.Error("Expected at least one gNB")
	}

	for i, gnb := range list.GNBs {
		if gnb.ID == "" {
			t.Errorf("GNB[%d].ID is empty", i)
		}
		if gnb.Host == "" {
			t.Errorf("GNB[%d].Host is empty", i)
		}
		if gnb.Type == "" {
			t.Errorf("GNB[%d].Type is empty", i)
		}
	}
}

func TestMockProvider_GetGNB(t *testing.T) {
	provider := NewMockProvider("local")
	ctx := context.Background()

	t.Run("existing gNB", func(t *testing.T) {
		gnb, err := provider.GetGNB(ctx, "gnb-0")
		if err != nil {
			t.Fatalf("GetGNB returned error: %v", err)
		}
		if gnb.ID != "gnb-0" {
			t.Errorf("Expected ID 'gnb-0', got '%s'", gnb.ID)
		}
	})

	t.Run("non-existent gNB", func(t *testing.T) {
		_, err := provider.GetGNB(ctx, "non-existent")
		if err == nil {
			t.Error("Expected error for non-existent gNB")
		}
	})
}

func TestMockProvider_CreateGNB(t *testing.T) {
	provider := NewMockProvider("local")
	ctx := context.Background()

	newGNB := &GNBConfig{
		ID:         "gnb-1",
		Name:       "Test gNB",
		Type:       "srsran",
		IP:         "10.0.0.100",
		Simulation: true,
	}

	t.Run("create new gNB", func(t *testing.T) {
		resp, err := provider.CreateGNB(ctx, newGNB)
		if err != nil {
			t.Fatalf("CreateGNB returned error: %v", err)
		}
		if !resp.Success {
			t.Error("Expected Success to be true")
		}

		// Verify gNB was created
		gnb, err := provider.GetGNB(ctx, "gnb-1")
		if err != nil {
			t.Fatalf("Failed to get created gNB: %v", err)
		}
		if gnb.Host != "local" {
			t.Errorf("Expected host 'local', got '%s'", gnb.Host)
		}
	})

	t.Run("create duplicate gNB", func(t *testing.T) {
		_, err := provider.CreateGNB(ctx, newGNB)
		if err == nil {
			t.Error("Expected error for duplicate gNB")
		}
	})
}

func TestMockProvider_UpdateGNB(t *testing.T) {
	provider := NewMockProvider("local")
	ctx := context.Background()

	t.Run("update existing gNB", func(t *testing.T) {
		updatedConfig := &GNBConfig{
			Name:       "Updated gNB",
			Type:       "ocudu",
			IP:         "10.0.0.200",
			Simulation: false,
		}

		err := provider.UpdateGNB(ctx, "gnb-0", updatedConfig)
		if err != nil {
			t.Fatalf("UpdateGNB returned error: %v", err)
		}

		gnb, _ := provider.GetGNB(ctx, "gnb-0")
		if gnb.Name != "Updated gNB" {
			t.Errorf("Expected Name 'Updated gNB', got '%s'", gnb.Name)
		}
	})

	t.Run("update non-existent gNB", func(t *testing.T) {
		err := provider.UpdateGNB(ctx, "non-existent", &GNBConfig{})
		if err == nil {
			t.Error("Expected error for non-existent gNB")
		}
	})
}

func TestMockProvider_DeleteGNB(t *testing.T) {
	provider := NewMockProvider("local")
	ctx := context.Background()

	t.Run("delete existing gNB", func(t *testing.T) {
		resp, err := provider.DeleteGNB(ctx, "gnb-0")
		if err != nil {
			t.Fatalf("DeleteGNB returned error: %v", err)
		}
		if !resp.Success {
			t.Error("Expected Success to be true")
		}

		// Verify gNB was deleted
		_, err = provider.GetGNB(ctx, "gnb-0")
		if err == nil {
			t.Error("Expected error when getting deleted gNB")
		}
	})

	t.Run("delete non-existent gNB", func(t *testing.T) {
		_, err := provider.DeleteGNB(ctx, "non-existent")
		if err == nil {
			t.Error("Expected error for non-existent gNB")
		}
	})
}

func TestMockProvider_GetGNBStatus(t *testing.T) {
	provider := NewMockProvider("local")
	ctx := context.Background()

	t.Run("existing gNB", func(t *testing.T) {
		status, err := provider.GetGNBStatus(ctx, "gnb-0")
		if err != nil {
			t.Fatalf("GetGNBStatus returned error: %v", err)
		}
		if status.ID != "gnb-0" {
			t.Errorf("Expected ID 'gnb-0', got '%s'", status.ID)
		}
	})

	t.Run("non-existent gNB", func(t *testing.T) {
		_, err := provider.GetGNBStatus(ctx, "non-existent")
		if err == nil {
			t.Error("Expected error for non-existent gNB")
		}
	})
}

func TestMockProvider_ListGNBStatuses(t *testing.T) {
	provider := NewMockProvider("local")
	ctx := context.Background()

	statuses, err := provider.ListGNBStatuses(ctx)
	if err != nil {
		t.Fatalf("ListGNBStatuses returned error: %v", err)
	}
	if statuses == nil {
		t.Fatal("ListGNBStatuses returned nil")
	}
	if len(statuses.GNBs) == 0 {
		t.Error("Expected at least one gNB status")
	}
}

func TestMockProviderImplementsInterface(t *testing.T) {
	var _ Provider = (*MockProvider)(nil)
}
