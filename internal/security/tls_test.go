package security

import (
	"crypto/tls"
	"os"
	"path/filepath"
	"testing"
)

func TestBuildTLSConfigEmpty(t *testing.T) {
	result, err := BuildTLSConfig(TLSOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatal("expected nil result for empty options")
	}
}

func TestBuildTLSConfigAutoTLS(t *testing.T) {
	dir := t.TempDir()
	result, err := BuildTLSConfig(TLSOptions{AutoTLS: true, DataDir: dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result for AutoTLS")
	}
	if !result.AutoCert {
		t.Error("AutoCert should be true")
	}
	if result.CertSource != "auto-generated" {
		t.Errorf("CertSource = %q, want %q", result.CertSource, "auto-generated")
	}
	if result.CertDir == "" {
		t.Error("CertDir should be set for auto-generated certs")
	}
	if result.MTLSEnabled {
		t.Error("MTLSEnabled should be false")
	}
	if result.Config.MinVersion != tls.VersionTLS12 {
		t.Errorf("MinVersion = %d, want TLS 1.2", result.Config.MinVersion)
	}
}

func TestBuildTLSConfigPartialCertKey(t *testing.T) {
	_, err := BuildTLSConfig(TLSOptions{CertFile: "cert.pem"})
	if err == nil {
		t.Fatal("expected error for cert without key")
	}

	_, err = BuildTLSConfig(TLSOptions{KeyFile: "key.pem"})
	if err == nil {
		t.Fatal("expected error for key without cert")
	}
}

func TestBuildTLSConfigUserCert(t *testing.T) {
	// Generate certs to disk for this test.
	srcDir := t.TempDir()
	if _, err := EnsureCert(srcDir); err != nil {
		t.Fatal(err)
	}
	paths := CertDir(srcDir)

	result, err := BuildTLSConfig(TLSOptions{CertFile: paths.ServerCert, KeyFile: paths.ServerKey})
	if err != nil {
		t.Fatalf("BuildTLSConfig: %v", err)
	}
	if result.AutoCert {
		t.Error("AutoCert should be false for user-provided cert")
	}
	if result.CertSource != "user-provided" {
		t.Errorf("CertSource = %q, want %q", result.CertSource, "user-provided")
	}
	if result.CertDir != "" {
		t.Errorf("CertDir should be empty for user-provided certs, got %q", result.CertDir)
	}
	if result.MTLSEnabled {
		t.Error("MTLSEnabled should be false without CA")
	}
}

func TestBuildTLSConfigAutoGenWithMTLS(t *testing.T) {
	// Use a self-signed cert as the CA file for testing.
	srcDir := t.TempDir()
	if _, err := EnsureCert(srcDir); err != nil {
		t.Fatal(err)
	}
	caPath := CertDir(srcDir).CACert

	dataDir := t.TempDir()
	result, err := BuildTLSConfig(TLSOptions{MTLSCAFile: caPath, DataDir: dataDir})
	if err != nil {
		t.Fatalf("BuildTLSConfig: %v", err)
	}
	if !result.AutoCert {
		t.Error("AutoCert should be true when no cert/key provided")
	}
	if result.CertSource != "auto-generated" {
		t.Errorf("CertSource = %q, want %q", result.CertSource, "auto-generated")
	}
	if !result.MTLSEnabled {
		t.Error("MTLSEnabled should be true")
	}
	if result.Config.ClientAuth != tls.RequireAndVerifyClientCert {
		t.Errorf("ClientAuth = %v, want RequireAndVerifyClientCert", result.Config.ClientAuth)
	}
}

func TestBuildTLSConfigMinVersion(t *testing.T) {
	dir := t.TempDir()
	result, err := BuildTLSConfig(TLSOptions{AutoTLS: true, DataDir: dir})
	if err != nil {
		t.Fatal(err)
	}
	if result.Config.MinVersion != tls.VersionTLS12 {
		t.Errorf("MinVersion = %d, want %d (TLS 1.2)", result.Config.MinVersion, tls.VersionTLS12)
	}
}

func TestBuildTLSConfigInvalidCertFile(t *testing.T) {
	_, err := BuildTLSConfig(TLSOptions{CertFile: "/nonexistent", KeyFile: "/nonexistent"})
	if err == nil {
		t.Fatal("expected error for invalid cert files")
	}
}

func TestBuildTLSConfigInvalidCAFile(t *testing.T) {
	dir := t.TempDir()
	_, err := BuildTLSConfig(TLSOptions{MTLSCAFile: "/nonexistent", DataDir: dir})
	if err == nil {
		t.Fatal("expected error for invalid CA file")
	}
}

func TestBuildTLSConfigInvalidCAPEM(t *testing.T) {
	dir := t.TempDir()
	caPath := filepath.Join(dir, "bad-ca.pem")
	if err := os.WriteFile(caPath, []byte("not a certificate"), 0o600); err != nil {
		t.Fatal(err)
	}

	dataDir := t.TempDir()
	_, err := BuildTLSConfig(TLSOptions{MTLSCAFile: caPath, DataDir: dataDir})
	if err == nil {
		t.Fatal("expected error for invalid PEM content")
	}
}

