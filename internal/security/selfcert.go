package security

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

// CertPaths holds the file paths for auto-generated CA and server certificates.
type CertPaths struct {
	CACert    string // ca.pem      — add this to browser/system trust store
	CAKey     string // ca-key.pem
	ServerCert string // server.pem
	ServerKey  string // server-key.pem
}

// CertDir returns a CertPaths rooted in the given directory.
func CertDir(dir string) CertPaths {
	return CertPaths{
		CACert:     filepath.Join(dir, "ca.pem"),
		CAKey:      filepath.Join(dir, "ca-key.pem"),
		ServerCert: filepath.Join(dir, "server.pem"),
		ServerKey:  filepath.Join(dir, "server-key.pem"),
	}
}

// EnsureCert loads existing CA-signed server certificates from dir, or
// generates them on first run. The CA certificate (ca.pem) can be added to a
// browser or system trust store so the server cert is trusted without
// per-connection overrides.
//
// Generated files:
//   - ca.pem, ca-key.pem       — local CA (1-year validity)
//   - server.pem, server-key.pem — server cert signed by the CA
//
// The server certificate includes SANs for localhost, 127.0.0.1, and ::1,
// plus any extras provided.
func EnsureCert(dir string, extraSANs ...string) (tls.Certificate, error) {
	paths := CertDir(dir)

	// If server cert+key already exist, just load them.
	if fileExists(paths.ServerCert) && fileExists(paths.ServerKey) {
		return tls.LoadX509KeyPair(paths.ServerCert, paths.ServerKey)
	}

	// Create the directory if needed. 0755 so non-root users can read
	// ca.pem for trust-store import (key files are still 0600).
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return tls.Certificate{}, fmt.Errorf("creating cert directory: %w", err)
	}

	// Generate CA.
	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, err
	}

	caSerial, err := randomSerial()
	if err != nil {
		return tls.Certificate{}, err
	}

	now := time.Now()
	caTmpl := &x509.Certificate{
		SerialNumber:          caSerial,
		Subject:               pkix.Name{CommonName: "Aether WebUI CA"},
		NotBefore:             now,
		NotAfter:              now.Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            0,
		MaxPathLenZero:        true,
	}

	caCertDER, err := x509.CreateCertificate(rand.Reader, caTmpl, caTmpl, &caKey.PublicKey, caKey)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("creating CA certificate: %w", err)
	}

	caCert, err := x509.ParseCertificate(caCertDER)
	if err != nil {
		return tls.Certificate{}, err
	}

	// Generate server cert signed by CA.
	srvKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, err
	}

	srvSerial, err := randomSerial()
	if err != nil {
		return tls.Certificate{}, err
	}

	srvTmpl := &x509.Certificate{
		SerialNumber: srvSerial,
		Subject:      pkix.Name{CommonName: "aether-webd"},
		NotBefore:    now,
		NotAfter:     now.Add(365 * 24 * time.Hour),

		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},

		DNSNames:    []string{"localhost"},
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
	}

	for _, san := range extraSANs {
		if ip := net.ParseIP(san); ip != nil {
			srvTmpl.IPAddresses = append(srvTmpl.IPAddresses, ip)
		} else {
			srvTmpl.DNSNames = append(srvTmpl.DNSNames, san)
		}
	}

	srvCertDER, err := x509.CreateCertificate(rand.Reader, srvTmpl, caCert, &srvKey.PublicKey, caKey)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("creating server certificate: %w", err)
	}

	// Write all PEM files to disk.
	if err := writePEM(paths.CACert, "CERTIFICATE", caCertDER, 0o644); err != nil {
		return tls.Certificate{}, err
	}
	if err := writeKeyPEM(paths.CAKey, caKey); err != nil {
		return tls.Certificate{}, err
	}
	if err := writePEM(paths.ServerCert, "CERTIFICATE", srvCertDER, 0o644); err != nil {
		return tls.Certificate{}, err
	}
	if err := writeKeyPEM(paths.ServerKey, srvKey); err != nil {
		return tls.Certificate{}, err
	}

	return tls.LoadX509KeyPair(paths.ServerCert, paths.ServerKey)
}

func randomSerial() (*big.Int, error) {
	return rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
}

func writePEM(path, blockType string, data []byte, perm os.FileMode) error {
	out := pem.EncodeToMemory(&pem.Block{Type: blockType, Bytes: data})
	if err := os.WriteFile(path, out, perm); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	return nil
}

func writeKeyPEM(path string, key *ecdsa.PrivateKey) error {
	der, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return err
	}
	return writePEM(path, "EC PRIVATE KEY", der, 0o600)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
