package store

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"
)

func NormalizeLabels(labels map[string]string) (norm string, hashHex string) {
	if len(labels) == 0 {
		sum := sha256.Sum256(nil)
		return "", hex.EncodeToString(sum[:])
	}

	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Stable, queryable representation
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+"="+labels[k])
	}
	norm = strings.Join(parts, "\n")

	sum := sha256.Sum256([]byte(norm))
	hashHex = hex.EncodeToString(sum[:])
	return norm, hashHex
}
