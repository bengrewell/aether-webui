package security

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"path/filepath"
)

// TLSOptions holds the file paths that control TLS configuration.
type TLSOptions struct {
	AutoTLS    bool   // enable TLS with auto-generated cert when no cert/key provided
	DataDir    string // base directory for persistent auto-generated certs
	CertFile   string // path to PEM-encoded server certificate
	KeyFile    string // path to PEM-encoded private key
	MTLSCAFile string // path to CA cert for client verification (enables mTLS)
}

// TLSResult holds a ready-to-use tls.Config along with metadata about how it
// was constructed (for logging/introspection).
type TLSResult struct {
	Config      *tls.Config
	AutoCert    bool   // true when certs were auto-generated
	CertSource  string // "auto-generated", "user-provided", or ""
	CertDir     string // directory containing auto-generated certs (empty for user-provided)
	MTLSEnabled bool
}

// BuildTLSConfig constructs a *TLSResult from the given options.
//
// Returns (nil, nil) when no TLS is requested (all fields empty).
// Returns an error if only one of CertFile/KeyFile is provided.
//
// When auto-generating, certificates are persisted under DataDir/certs/ and
// reused on subsequent starts. The CA certificate (ca.pem) can be added to
// browser or system trust stores.
func BuildTLSConfig(opts TLSOptions) (*TLSResult, error) {
	hasCert := opts.CertFile != ""
	hasKey := opts.KeyFile != ""
	hasMTLS := opts.MTLSCAFile != ""

	// Nothing requested.
	if !opts.AutoTLS && !hasCert && !hasKey && !hasMTLS {
		return nil, nil
	}

	// Partial cert/key is an error.
	if hasCert != hasKey {
		return nil, fmt.Errorf("both --tls-cert and --tls-key are required (got cert=%q, key=%q)", opts.CertFile, opts.KeyFile)
	}

	result := &TLSResult{}
	cfg := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	// Load or generate server certificate.
	if hasCert {
		cert, err := tls.LoadX509KeyPair(opts.CertFile, opts.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("loading TLS certificate: %w", err)
		}
		cfg.Certificates = []tls.Certificate{cert}
		result.CertSource = "user-provided"
	} else {
		certDir := filepath.Join(opts.DataDir, "certs")
		cert, err := EnsureCert(certDir)
		if err != nil {
			return nil, fmt.Errorf("auto-generating TLS certificates: %w", err)
		}
		cfg.Certificates = []tls.Certificate{cert}
		result.AutoCert = true
		result.CertSource = "auto-generated"
		result.CertDir = certDir
	}

	// mTLS: require and verify client certificates against provided CA.
	if hasMTLS {
		caPEM, err := os.ReadFile(opts.MTLSCAFile)
		if err != nil {
			return nil, fmt.Errorf("reading mTLS CA certificate: %w", err)
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(caPEM) {
			return nil, fmt.Errorf("mTLS CA file contains no valid certificates: %s", opts.MTLSCAFile)
		}
		cfg.ClientCAs = pool
		cfg.ClientAuth = tls.RequireAndVerifyClientCert
		result.MTLSEnabled = true
	}

	result.Config = cfg
	return result, nil
}
