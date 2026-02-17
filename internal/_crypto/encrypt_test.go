package crypto

import (
	"strings"
	"testing"
)

const testKey = "01234567890123456789012345678901" // exactly 32 bytes

func TestEncryptDecryptRoundTrip(t *testing.T) {
	plaintext := "super-secret-password"

	encrypted, err := Encrypt(plaintext, testKey)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}
	if encrypted == plaintext {
		t.Error("encrypted text should differ from plaintext")
	}

	decrypted, err := Decrypt(encrypted, testKey)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}
	if decrypted != plaintext {
		t.Errorf("expected %q, got %q", plaintext, decrypted)
	}
}

func TestEncryptDecryptEmpty(t *testing.T) {
	encrypted, err := Encrypt("", testKey)
	if err != nil {
		t.Fatalf("Encrypt empty failed: %v", err)
	}

	decrypted, err := Decrypt(encrypted, testKey)
	if err != nil {
		t.Fatalf("Decrypt empty failed: %v", err)
	}
	if decrypted != "" {
		t.Errorf("expected empty string, got %q", decrypted)
	}
}

func TestDecryptWrongKey(t *testing.T) {
	encrypted, err := Encrypt("secret", testKey)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	wrongKey := "98765432109876543210987654321098"
	_, err = Decrypt(encrypted, wrongKey)
	if err == nil {
		t.Error("expected error when decrypting with wrong key")
	}
}

func TestDecryptCorruptedData(t *testing.T) {
	encrypted, err := Encrypt("secret", testKey)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	// Corrupt the encrypted data
	corrupted := encrypted[:len(encrypted)-4] + "XXXX"
	_, err = Decrypt(corrupted, testKey)
	if err == nil {
		t.Error("expected error when decrypting corrupted data")
	}
}

func TestEncryptInvalidKeyLength(t *testing.T) {
	_, err := Encrypt("test", "short-key")
	if err == nil {
		t.Error("expected error for short key")
	}
}

func TestDecryptInvalidKeyLength(t *testing.T) {
	_, err := Decrypt("dGVzdA==", "short-key")
	if err == nil {
		t.Error("expected error for short key")
	}
}

func TestDecryptInvalidBase64(t *testing.T) {
	_, err := Decrypt("not-valid-base64!!!", testKey)
	if err == nil {
		t.Error("expected error for invalid base64")
	}
}

func TestDecryptTooShort(t *testing.T) {
	// Base64 of a very short byte slice (less than nonce size)
	_, err := Decrypt("AQID", testKey)
	if err == nil {
		t.Error("expected error for ciphertext too short")
	}
}

func TestEncryptProducesDifferentCiphertexts(t *testing.T) {
	plaintext := "same-input"
	enc1, err := Encrypt(plaintext, testKey)
	if err != nil {
		t.Fatalf("Encrypt 1 failed: %v", err)
	}
	enc2, err := Encrypt(plaintext, testKey)
	if err != nil {
		t.Fatalf("Encrypt 2 failed: %v", err)
	}
	if enc1 == enc2 {
		t.Error("two encryptions of the same plaintext should produce different ciphertexts (random nonce)")
	}
}

func TestGenerateKey(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}
	if len(key) != 32 {
		t.Errorf("expected 32-byte key, got %d bytes", len(key))
	}

	// Verify the generated key works with encrypt/decrypt
	encrypted, err := Encrypt("test", key)
	if err != nil {
		t.Fatalf("Encrypt with generated key failed: %v", err)
	}
	decrypted, err := Decrypt(encrypted, key)
	if err != nil {
		t.Fatalf("Decrypt with generated key failed: %v", err)
	}
	if decrypted != "test" {
		t.Errorf("expected 'test', got %q", decrypted)
	}
}

func TestEncryptDecryptLongText(t *testing.T) {
	plaintext := strings.Repeat("a", 10000)
	encrypted, err := Encrypt(plaintext, testKey)
	if err != nil {
		t.Fatalf("Encrypt long text failed: %v", err)
	}
	decrypted, err := Decrypt(encrypted, testKey)
	if err != nil {
		t.Fatalf("Decrypt long text failed: %v", err)
	}
	if decrypted != plaintext {
		t.Error("round-trip failed for long text")
	}
}
