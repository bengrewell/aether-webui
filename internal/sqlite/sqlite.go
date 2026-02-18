package sqlite

import (
	"context"
	"database/sql"
	"time"

	_ "modernc.org/sqlite"

	"github.com/bengrewell/aether-webui/internal/store"
)

type Crypter interface {
	Encrypt(plaintext []byte) (ciphertext []byte, err error)
	Decrypt(ciphertext []byte) (plaintext []byte, err error)
}

type NoopCrypter struct{}

func (NoopCrypter) Encrypt(p []byte) ([]byte, error) { return p, nil }
func (NoopCrypter) Decrypt(c []byte) ([]byte, error) { return c, nil }

type Config struct {
	Path          string
	BusyTimeout   time.Duration
	Crypter       Crypter
	Now           func() time.Time
	MetricsMaxAge time.Duration // optional retention cleanup in CompactMetrics
}

type Store struct {
	db      *sql.DB
	crypter Crypter
	now     func() time.Time

	metricsMaxAge time.Duration
}

func Open(ctx context.Context, cfg Config) (*Store, error) {
	if cfg.Path == "" {
		return nil, store.ErrInvalidArgument
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

	db, err := sql.Open("sqlite", cfg.Path)
	if err != nil {
		return nil, err
	}

	s := &Store{
		db:            db,
		crypter:       cfg.Crypter,
		now:           cfg.Now,
		metricsMaxAge: cfg.MetricsMaxAge,
	}

	if err := applyPragmas(ctx, db, cfg.BusyTimeout); err != nil {
		_ = db.Close()
		return nil, err
	}

	return s, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) Health(ctx context.Context) error {
	var one int
	return s.db.QueryRowContext(ctx, "SELECT 1").Scan(&one)
}

func (s *Store) Migrate(ctx context.Context) error {
	return migrate(ctx, s.db)
}

func (s *Store) GetSchemaVersion() (int, error) {
	var version int
	err := s.db.QueryRow(`SELECT COALESCE(MAX(version), 0) FROM schema_migrations`).Scan(&version)
	return version, err
}

func (s *Store) CompactMetrics(ctx context.Context) error {
	if s.metricsMaxAge <= 0 {
		return nil
	}
	cutoff := s.now().Add(-s.metricsMaxAge).Unix()
	_, err := s.db.ExecContext(ctx, `DELETE FROM metrics_samples WHERE ts < ?`, cutoff)
	return err
}
