package store

import (
	"context"
	"database/sql"
	"time"

	_ "modernc.org/sqlite"
)

// Crypter encrypts and decrypts credential secrets at rest.
type Crypter interface {
	Encrypt(plaintext []byte) (ciphertext []byte, err error)
	Decrypt(ciphertext []byte) (plaintext []byte, err error)
}

// NoopCrypter stores secrets in plaintext (no encryption).
type NoopCrypter struct{}

func (NoopCrypter) Encrypt(p []byte) ([]byte, error) { return p, nil }
func (NoopCrypter) Decrypt(c []byte) ([]byte, error) { return c, nil }

type dbConfig struct {
	Path          string
	BusyTimeout   time.Duration
	Crypter       Crypter
	Now           func() time.Time
	MetricsMaxAge time.Duration
}

type db struct {
	conn    *sql.DB
	crypter Crypter
	now     func() time.Time

	metricsMaxAge time.Duration
}

func openDB(ctx context.Context, cfg dbConfig) (*db, error) {
	if cfg.Path == "" {
		return nil, ErrInvalidArgument
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	if cfg.Crypter == nil {
		cfg.Crypter = NoopCrypter{}
	}
	if cfg.BusyTimeout <= 0 {
		cfg.BusyTimeout = 5 * time.Second
	}

	conn, err := sql.Open("sqlite", cfg.Path)
	if err != nil {
		return nil, err
	}

	d := &db{
		conn:          conn,
		crypter:       cfg.Crypter,
		now:           cfg.Now,
		metricsMaxAge: cfg.MetricsMaxAge,
	}

	if err := applyPragmas(ctx, conn, cfg.BusyTimeout); err != nil {
		_ = conn.Close()
		return nil, err
	}

	return d, nil
}

func (d *db) Close() error {
	return d.conn.Close()
}

func (d *db) Health(ctx context.Context) error {
	var one int
	return d.conn.QueryRowContext(ctx, "SELECT 1").Scan(&one)
}

func (d *db) Migrate(ctx context.Context) error {
	return migrate(ctx, d.conn)
}

func (d *db) GetSchemaVersion() (int, error) {
	var version int
	err := d.conn.QueryRow(`SELECT COALESCE(MAX(version), 0) FROM schema_migrations`).Scan(&version)
	return version, err
}

func (d *db) CompactMetrics(ctx context.Context) error {
	if d.metricsMaxAge <= 0 {
		return nil
	}
	cutoff := d.now().Add(-d.metricsMaxAge).Unix()
	_, err := d.conn.ExecContext(ctx, `DELETE FROM metrics_samples WHERE ts < ?`, cutoff)
	return err
}
