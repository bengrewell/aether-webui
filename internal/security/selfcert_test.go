package security

import (
	"crypto/tls"
	"crypto/x509"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestEnsureCertGeneratesFiles(t *testing.T) {
	dir := t.TempDir()
	cert, err := EnsureCert(dir)
	if err != nil {
		t.Fatalf("EnsureCert() error: %v", err)
	}
	if len(cert.Certificate) == 0 {
		t.Fatal("no certificate in chain")
	}

	// All four files should exist.
	paths := CertDir(dir)
	for _, path := range []string{paths.CACert, paths.CAKey, paths.ServerCert, paths.ServerKey} {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected file %s to exist: %v", filepath.Base(path), err)
		}
	}

	// Key files should be owner-only.
	for _, path := range []string{paths.CAKey, paths.ServerKey} {
		fi, err := os.Stat(path)
		if err != nil {
			continue
		}
		if fi.Mode().Perm()&0o077 != 0 {
			t.Errorf("%s permissions %o allow group/other access", filepath.Base(path), fi.Mode().Perm())
		}
	}
}

func TestEnsureCertServerCertProperties(t *testing.T) {
	dir := t.TempDir()
	cert, err := EnsureCert(dir)
	if err != nil {
		t.Fatalf("EnsureCert() error: %v", err)
	}

	parsed, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		t.Fatalf("parse certificate: %v", err)
	}

	// Check subject CN.
	if parsed.Subject.CommonName != "aether-webd" {
		t.Errorf("CN = %q, want %q", parsed.Subject.CommonName, "aether-webd")
	}

	// Should NOT be a CA.
	if parsed.IsCA {
		t.Error("server cert should not be a CA")
	}

	// Check validity window.
	now := time.Now()
	if parsed.NotBefore.After(now) {
		t.Errorf("NotBefore %v is in the future", parsed.NotBefore)
	}
	expectedExpiry := now.Add(365 * 24 * time.Hour)
	if parsed.NotAfter.Before(expectedExpiry.Add(-time.Minute)) {
		t.Errorf("NotAfter %v too early (expected ~%v)", parsed.NotAfter, expectedExpiry)
	}

	// Check default SANs.
	wantDNS := []string{"localhost"}
	if len(parsed.DNSNames) != len(wantDNS) {
		t.Errorf("DNSNames = %v, want %v", parsed.DNSNames, wantDNS)
	} else {
		for i, want := range wantDNS {
			if parsed.DNSNames[i] != want {
				t.Errorf("DNSNames[%d] = %q, want %q", i, parsed.DNSNames[i], want)
			}
		}
	}

	hasIPv4 := false
	hasIPv6 := false
	for _, ip := range parsed.IPAddresses {
		if ip.Equal(net.IPv4(127, 0, 0, 1)) {
			hasIPv4 = true
		}
		if ip.Equal(net.IPv6loopback) {
			hasIPv6 = true
		}
	}
	if !hasIPv4 {
		t.Error("missing 127.0.0.1 IP SAN")
	}
	if !hasIPv6 {
		t.Error("missing ::1 IP SAN")
	}
}

func TestEnsureCertCACertProperties(t *testing.T) {
	dir := t.TempDir()
	if _, err := EnsureCert(dir); err != nil {
		t.Fatalf("EnsureCert() error: %v", err)
	}

	paths := CertDir(dir)
	caPEM, err := os.ReadFile(paths.CACert)
	if err != nil {
		t.Fatal(err)
	}

	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caPEM) {
		t.Fatal("ca.pem contains no valid certificates")
	}
}

func TestEnsureCertExtraSANs(t *testing.T) {
	dir := t.TempDir()
	cert, err := EnsureCert(dir, "example.com", "192.168.1.1")
	if err != nil {
		t.Fatalf("EnsureCert() error: %v", err)
	}

	parsed, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		t.Fatalf("parse certificate: %v", err)
	}

	foundExample := false
	for _, name := range parsed.DNSNames {
		if name == "example.com" {
			foundExample = true
		}
	}
	if !foundExample {
		t.Errorf("DNSNames %v missing example.com", parsed.DNSNames)
	}

	foundExtra := false
	for _, ip := range parsed.IPAddresses {
		if ip.Equal(net.ParseIP("192.168.1.1")) {
			foundExtra = true
		}
	}
	if !foundExtra {
		t.Errorf("IPAddresses %v missing 192.168.1.1", parsed.IPAddresses)
	}
}

func TestEnsureCertReusesExisting(t *testing.T) {
	dir := t.TempDir()

	// First call generates.
	cert1, err := EnsureCert(dir)
	if err != nil {
		t.Fatalf("first EnsureCert: %v", err)
	}

	// Second call loads from disk.
	cert2, err := EnsureCert(dir)
	if err != nil {
		t.Fatalf("second EnsureCert: %v", err)
	}

	// Both should have the same certificate bytes.
	if len(cert1.Certificate) == 0 || len(cert2.Certificate) == 0 {
		t.Fatal("empty certificate chain")
	}
	if string(cert1.Certificate[0]) != string(cert2.Certificate[0]) {
		t.Error("second call returned different certificate; expected reuse")
	}
}

func TestEnsureCertTLSHandshake(t *testing.T) {
	dir := t.TempDir()
	cert, err := EnsureCert(dir)
	if err != nil {
		t.Fatalf("EnsureCert() error: %v", err)
	}

	// Load the CA cert to use as a trust anchor.
	paths := CertDir(dir)
	caPEM, err := os.ReadFile(paths.CACert)
	if err != nil {
		t.Fatal(err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caPEM) {
		t.Fatal("ca.pem invalid")
	}

	serverCfg := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}
	clientCfg := &tls.Config{
		RootCAs:    pool,
		ServerName: "localhost",
		MinVersion: tls.VersionTLS12,
	}

	ln, err := tls.Listen("tcp", "127.0.0.1:0", serverCfg)
	if err != nil {
		t.Fatalf("tls.Listen: %v", err)
	}
	defer ln.Close()

	errCh := make(chan error, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			errCh <- err
			return
		}
		defer conn.Close()
		errCh <- conn.(*tls.Conn).Handshake()
	}()

	conn, err := tls.Dial("tcp", ln.Addr().String(), clientCfg)
	if err != nil {
		t.Fatalf("tls.Dial: %v", err)
	}
	conn.Close()

	if err := <-errCh; err != nil {
		t.Fatalf("server handshake error: %v", err)
	}
}
