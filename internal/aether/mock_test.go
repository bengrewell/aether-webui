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

func TestMockProvider_ListCores(t *testing.T) {
	provider := NewMockProvider("local")
	ctx := context.Background()

	list, err := provider.ListCores(ctx)
	if err != nil {
		t.Fatalf("ListCores returned error: %v", err)
	}
	if list == nil {
		t.Fatal("ListCores returned nil")
	}
	if len(list.Cores) == 0 {
		t.Error("Expected at least one core")
	}

	for i, core := range list.Cores {
		if core.ID == "" {
			t.Errorf("Core[%d].ID is empty", i)
		}
		if core.Name == "" {
			t.Errorf("Core[%d].Name is empty", i)
		}
	}
}

func TestMockProvider_GetCore(t *testing.T) {
	provider := NewMockProvider("local")
	ctx := context.Background()

	t.Run("existing core", func(t *testing.T) {
		config, err := provider.GetCore(ctx, "core-0")
		if err != nil {
			t.Fatalf("GetCore returned error: %v", err)
		}
		if config == nil {
			t.Fatal("GetCore returned nil")
		}
		if config.ID != "core-0" {
			t.Errorf("Expected ID 'core-0', got '%s'", config.ID)
		}
		if config.Name == "" {
			t.Error("Name is empty")
		}
		if config.Helm.ChartRef == "" {
			t.Error("Helm.ChartRef is empty")
		}
	})

	t.Run("non-existent core", func(t *testing.T) {
		_, err := provider.GetCore(ctx, "non-existent")
		if err == nil {
			t.Error("Expected error for non-existent core")
		}
	})
}

func TestMockProvider_DeployCore(t *testing.T) {
	provider := NewMockProvider("local")
	ctx := context.Background()

	t.Run("deploy with nil config uses defaults", func(t *testing.T) {
		resp, err := provider.DeployCore(ctx, nil)
		if err != nil {
			t.Fatalf("DeployCore returned error: %v", err)
		}
		if resp == nil {
			t.Fatal("DeployCore returned nil response")
		}
		if !resp.Success {
			t.Error("Expected Success to be true")
		}
		if resp.ID == "" {
			t.Error("ID is empty")
		}
		if resp.TaskID == "" {
			t.Error("TaskID is empty")
		}

		// Verify core was created with generated ID
		core, err := provider.GetCore(ctx, resp.ID)
		if err != nil {
			t.Fatalf("Failed to get deployed core: %v", err)
		}
		if core.Name == "" {
			t.Error("Name should have been generated")
		}
	})

	t.Run("deploy with config and custom name", func(t *testing.T) {
		customConfig := &CoreConfig{
			Name:       "My Custom Core",
			Standalone: true,
			DataIface:  "eth1",
			Helm: HelmConfig{
				ChartRef:     "custom/chart",
				ChartVersion: "2.0.0",
			},
		}

		resp, err := provider.DeployCore(ctx, customConfig)
		if err != nil {
			t.Fatalf("DeployCore returned error: %v", err)
		}
		if !resp.Success {
			t.Error("Expected Success to be true")
		}
		if resp.ID == "" {
			t.Error("Expected generated ID")
		}

		// Verify config was applied
		core, _ := provider.GetCore(ctx, resp.ID)
		if core.Name != "My Custom Core" {
			t.Errorf("Expected Name 'My Custom Core', got '%s'", core.Name)
		}
		if core.DataIface != "eth1" {
			t.Errorf("Expected DataIface 'eth1', got '%s'", core.DataIface)
		}

		// Verify status reflects new version
		status, _ := provider.GetCoreStatus(ctx, resp.ID)
		if status.Version != "2.0.0" {
			t.Errorf("Expected version '2.0.0', got '%s'", status.Version)
		}
	})

	t.Run("deploy duplicate core fails", func(t *testing.T) {
		duplicateConfig := &CoreConfig{
			ID: "core-0", // Already exists
		}
		_, err := provider.DeployCore(ctx, duplicateConfig)
		if err == nil {
			t.Error("Expected error for duplicate core")
		}
	})
}

func TestMockProvider_UpdateCore(t *testing.T) {
	provider := NewMockProvider("local")
	ctx := context.Background()

	t.Run("update existing core", func(t *testing.T) {
		newConfig := &CoreConfig{
			Name:       "Updated Core",
			Standalone: false,
			DataIface:  "eth0",
			Helm: HelmConfig{
				ChartRef:     "custom/chart",
				ChartVersion: "1.0.0",
			},
		}

		err := provider.UpdateCore(ctx, "core-0", newConfig)
		if err != nil {
			t.Fatalf("UpdateCore returned error: %v", err)
		}

		// Verify update took effect
		config, _ := provider.GetCore(ctx, "core-0")
		if config.Name != "Updated Core" {
			t.Errorf("Name should be 'Updated Core', got '%s'", config.Name)
		}
		if config.DataIface != "eth0" {
			t.Errorf("DataIface should be 'eth0', got '%s'", config.DataIface)
		}
	})

	t.Run("update non-existent core", func(t *testing.T) {
		err := provider.UpdateCore(ctx, "non-existent", &CoreConfig{})
		if err == nil {
			t.Error("Expected error for non-existent core")
		}
	})
}

func TestMockProvider_UndeployCore(t *testing.T) {
	provider := NewMockProvider("local")
	ctx := context.Background()

	t.Run("undeploy existing core", func(t *testing.T) {
		resp, err := provider.UndeployCore(ctx, "core-0")
		if err != nil {
			t.Fatalf("UndeployCore returned error: %v", err)
		}
		if !resp.Success {
			t.Error("Expected Success to be true")
		}

		// Verify core was removed
		_, err = provider.GetCore(ctx, "core-0")
		if err == nil {
			t.Error("Expected error when getting undeployed core")
		}
	})

	t.Run("undeploy non-existent core", func(t *testing.T) {
		_, err := provider.UndeployCore(ctx, "non-existent")
		if err == nil {
			t.Error("Expected error for non-existent core")
		}
	})
}

func TestMockProvider_GetCoreStatus(t *testing.T) {
	provider := NewMockProvider("local")
	ctx := context.Background()

	t.Run("existing core", func(t *testing.T) {
		status, err := provider.GetCoreStatus(ctx, "core-0")
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
	})

	t.Run("non-existent core", func(t *testing.T) {
		_, err := provider.GetCoreStatus(ctx, "non-existent")
		if err == nil {
			t.Error("Expected error for non-existent core")
		}
	})
}

func TestMockProvider_ListCoreStatuses(t *testing.T) {
	provider := NewMockProvider("local")
	ctx := context.Background()

	statuses, err := provider.ListCoreStatuses(ctx)
	if err != nil {
		t.Fatalf("ListCoreStatuses returned error: %v", err)
	}
	if statuses == nil {
		t.Fatal("ListCoreStatuses returned nil")
	}
	if len(statuses.Cores) == 0 {
		t.Error("Expected at least one core status")
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

func TestMockProvider_DeployGNB(t *testing.T) {
	provider := NewMockProvider("local")
	ctx := context.Background()

	t.Run("deploy with full config", func(t *testing.T) {
		newGNB := &GNBConfig{
			Name:       "Test gNB",
			Type:       "srsran",
			IP:         "10.0.0.100",
			Simulation: true,
		}

		resp, err := provider.DeployGNB(ctx, newGNB)
		if err != nil {
			t.Fatalf("DeployGNB returned error: %v", err)
		}
		if !resp.Success {
			t.Error("Expected Success to be true")
		}
		if resp.ID == "" {
			t.Error("Expected generated ID")
		}

		// Verify gNB was created
		gnb, err := provider.GetGNB(ctx, resp.ID)
		if err != nil {
			t.Fatalf("Failed to get deployed gNB: %v", err)
		}
		if gnb.Host != "local" {
			t.Errorf("Expected host 'local', got '%s'", gnb.Host)
		}
		if gnb.Name != "Test gNB" {
			t.Errorf("Expected Name 'Test gNB', got '%s'", gnb.Name)
		}
	})

	t.Run("deploy with nil config uses defaults", func(t *testing.T) {
		resp, err := provider.DeployGNB(ctx, nil)
		if err != nil {
			t.Fatalf("DeployGNB returned error: %v", err)
		}
		if !resp.Success {
			t.Error("Expected Success to be true")
		}
		if resp.ID == "" {
			t.Error("Expected generated ID")
		}

		// Verify gNB has generated name
		gnb, _ := provider.GetGNB(ctx, resp.ID)
		if gnb.Name == "" {
			t.Error("Expected Name to have generated value")
		}
	})

	t.Run("deploy with partial config applies defaults", func(t *testing.T) {
		partialGNB := &GNBConfig{
			Name: "Partial Config gNB",
			// Type, IP, Docker config will use defaults
		}

		resp, err := provider.DeployGNB(ctx, partialGNB)
		if err != nil {
			t.Fatalf("DeployGNB returned error: %v", err)
		}
		if !resp.Success {
			t.Error("Expected Success to be true")
		}

		// Verify defaults were applied
		gnb, _ := provider.GetGNB(ctx, resp.ID)
		if gnb.Type == "" {
			t.Error("Expected Type to have default value")
		}
		if gnb.IP == "" {
			t.Error("Expected IP to have default value")
		}
	})

	t.Run("deploy duplicate gNB fails", func(t *testing.T) {
		duplicateGNB := &GNBConfig{
			ID: "gnb-0", // Already exists
		}
		_, err := provider.DeployGNB(ctx, duplicateGNB)
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

func TestMockProvider_UndeployGNB(t *testing.T) {
	provider := NewMockProvider("local")
	ctx := context.Background()

	t.Run("undeploy existing gNB", func(t *testing.T) {
		resp, err := provider.UndeployGNB(ctx, "gnb-0")
		if err != nil {
			t.Fatalf("UndeployGNB returned error: %v", err)
		}
		if !resp.Success {
			t.Error("Expected Success to be true")
		}

		// Verify gNB was removed
		_, err = provider.GetGNB(ctx, "gnb-0")
		if err == nil {
			t.Error("Expected error when getting undeployed gNB")
		}
	})

	t.Run("undeploy non-existent gNB", func(t *testing.T) {
		_, err := provider.UndeployGNB(ctx, "non-existent")
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
		if status.Name == "" {
			t.Error("Name should not be empty")
		}
		if status.Host != "local" {
			t.Errorf("Expected Host 'local', got '%s'", status.Host)
		}
		if status.Type == "" {
			t.Error("Type should not be empty")
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
