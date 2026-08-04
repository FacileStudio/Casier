package crypto

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"testing"
)

func testKey(t *testing.T) []byte {
	t.Helper()
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("generate key: %v", err)
	}
	return key
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	key := testKey(t)

	cases := []struct {
		name      string
		plaintext []byte
	}{
		{name: "empty", plaintext: []byte{}},
		{name: "short", plaintext: []byte("hello")},
		{name: "binary", plaintext: []byte{0x00, 0x01, 0xFF, 0xFE, 0x00}},
		{name: "unicode", plaintext: []byte("clé secrète 🔑")},
		{name: "large", plaintext: bytes.Repeat([]byte("A"), 64*1024)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ciphertext, nonce, err := Encrypt(tc.plaintext, key)
			if err != nil {
				t.Fatalf("Encrypt: %v", err)
			}
			if bytes.Contains(ciphertext, tc.plaintext) && len(tc.plaintext) > 0 {
				t.Fatal("ciphertext contains plaintext")
			}
			plaintext, err := Decrypt(ciphertext, nonce, key)
			if err != nil {
				t.Fatalf("Decrypt: %v", err)
			}
			if !bytes.Equal(plaintext, tc.plaintext) {
				t.Fatalf("round trip mismatch: got %q, want %q", plaintext, tc.plaintext)
			}
		})
	}
}

func TestDecryptRejectsTampering(t *testing.T) {
	key := testKey(t)
	otherKey := testKey(t)

	ciphertext, nonce, err := Encrypt([]byte("sensitive value"), key)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	cases := []struct {
		name       string
		ciphertext func() []byte
		nonce      func() []byte
		key        []byte
	}{
		{
			name:       "flipped ciphertext byte",
			ciphertext: func() []byte { c := bytes.Clone(ciphertext); c[0] ^= 0x01; return c },
			nonce:      func() []byte { return nonce },
			key:        key,
		},
		{
			name:       "flipped auth tag byte",
			ciphertext: func() []byte { c := bytes.Clone(ciphertext); c[len(c)-1] ^= 0x01; return c },
			nonce:      func() []byte { return nonce },
			key:        key,
		},
		{
			name:       "truncated ciphertext",
			ciphertext: func() []byte { return ciphertext[:len(ciphertext)-1] },
			nonce:      func() []byte { return nonce },
			key:        key,
		},
		{
			name:       "flipped nonce byte",
			ciphertext: func() []byte { return ciphertext },
			nonce:      func() []byte { n := bytes.Clone(nonce); n[0] ^= 0x01; return n },
			key:        key,
		},
		{
			name:       "wrong key",
			ciphertext: func() []byte { return ciphertext },
			nonce:      func() []byte { return nonce },
			key:        otherKey,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := Decrypt(tc.ciphertext(), tc.nonce(), tc.key); err == nil {
				t.Fatal("Decrypt succeeded, want error")
			}
		})
	}
}

func TestKeyFromBase64(t *testing.T) {
	valid := base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{0x42}, 32))

	cases := []struct {
		name    string
		encoded string
		wantErr bool
	}{
		{name: "valid 32-byte key", encoded: valid, wantErr: false},
		{name: "invalid base64", encoded: "not base64!!!", wantErr: true},
		{name: "too short", encoded: base64.StdEncoding.EncodeToString([]byte("short")), wantErr: true},
		{name: "too long", encoded: base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{0x42}, 33)), wantErr: true},
		{name: "empty", encoded: "", wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			key, err := KeyFromBase64(tc.encoded)
			if tc.wantErr {
				if err == nil {
					t.Fatal("KeyFromBase64 succeeded, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("KeyFromBase64: %v", err)
			}
			if len(key) != 32 {
				t.Fatalf("key length: got %d, want 32", len(key))
			}
		})
	}
}
