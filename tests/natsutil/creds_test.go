package natsutil_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	natsutil "DataConsumer/internal/natsutil"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nkeys"
)

func TestJWTAuth_SetsUserJWTAndSignatureCallbacks(t *testing.T) {
	kp, err := nkeys.CreateUser()
	if err != nil {
		t.Fatalf("CreateUser() unexpected error: %v", err)
	}

	seed, err := kp.Seed()
	if err != nil {
		t.Fatalf("Seed() unexpected error: %v", err)
	}

	const token = "user.jwt.token"
	opt := natsutil.JWTAuth(token, string(seed))

	var options nats.Options
	if err := opt(&options); err != nil {
		t.Fatalf("option apply unexpected error: %v", err)
	}

	if options.UserJWT == nil {
		t.Fatal("UserJWT callback should be set")
	}
	if options.SignatureCB == nil {
		t.Fatal("SignatureCB callback should be set")
	}

	gotToken, err := options.UserJWT()
	if err != nil {
		t.Fatalf("UserJWT() unexpected error: %v", err)
	}
	if gotToken != token {
		t.Fatalf("UserJWT() got %q, want %q", gotToken, token)
	}

	nonce := []byte("nonce-value")
	sig, err := options.SignatureCB(nonce)
	if err != nil {
		t.Fatalf("SignatureCB() unexpected error: %v", err)
	}

	pub, err := kp.PublicKey()
	if err != nil {
		t.Fatalf("PublicKey() unexpected error: %v", err)
	}
	if !nkeys.IsValidPublicUserKey(pub) {
		t.Fatalf("generated key %q is not a valid public user key", pub)
	}

	pubKP, err := nkeys.FromPublicKey(pub)
	if err != nil {
		t.Fatalf("FromPublicKey() unexpected error: %v", err)
	}

	if err := pubKP.Verify(nonce, sig); err != nil {
		t.Fatalf("signature verification failed: %v", err)
	}
}

func TestJWTAuth_SignatureCallbackReturnsErrorOnInvalidSeed(t *testing.T) {
	opt := natsutil.JWTAuth("token", "invalid-seed")

	var options nats.Options
	if err := opt(&options); err != nil {
		t.Fatalf("option apply unexpected error: %v", err)
	}

	if options.SignatureCB == nil {
		t.Fatal("SignatureCB callback should be set")
	}

	if _, err := options.SignatureCB([]byte("nonce")); err == nil {
		t.Fatal("SignatureCB() expected error for invalid seed, got nil")
	}
}

func TestCredsFileAuth_SetsUserJWTAndSignatureCallbacks(t *testing.T) {
	kp, err := nkeys.CreateUser()
	if err != nil {
		t.Fatalf("CreateUser() unexpected error: %v", err)
	}

	seed, err := kp.Seed()
	if err != nil {
		t.Fatalf("Seed() unexpected error: %v", err)
	}

	const token = "user.jwt.token"
	credsPath := filepath.Join(t.TempDir(), "user.creds")
	creds := fmt.Sprintf(`-----BEGIN NATS USER JWT-----
%s
------END NATS USER JWT------

************************* IMPORTANT *************************
NKEY Seed printed below can be used to sign and prove identity.
NKEYs are sensitive and should be treated as secrets.

-----BEGIN USER NKEY SEED-----
%s
------END USER NKEY SEED------

*************************************************************
`, token, seed)

	if err := os.WriteFile(credsPath, []byte(creds), 0o600); err != nil {
		t.Fatalf("WriteFile() unexpected error: %v", err)
	}

	opt := natsutil.CredsFileAuth(credsPath)

	var options nats.Options
	if err := opt(&options); err != nil {
		t.Fatalf("option apply unexpected error: %v", err)
	}

	if options.UserJWT == nil {
		t.Fatal("UserJWT callback should be set")
	}
	if options.SignatureCB == nil {
		t.Fatal("SignatureCB callback should be set")
	}

	gotToken, err := options.UserJWT()
	if err != nil {
		t.Fatalf("UserJWT() unexpected error: %v", err)
	}
	if gotToken != token {
		t.Fatalf("UserJWT() got %q, want %q", gotToken, token)
	}

	nonce := []byte("nonce-value")
	sig, err := options.SignatureCB(nonce)
	if err != nil {
		t.Fatalf("SignatureCB() unexpected error: %v", err)
	}

	pub, err := kp.PublicKey()
	if err != nil {
		t.Fatalf("PublicKey() unexpected error: %v", err)
	}
	pubKP, err := nkeys.FromPublicKey(pub)
	if err != nil {
		t.Fatalf("FromPublicKey() unexpected error: %v", err)
	}
	if err := pubKP.Verify(nonce, sig); err != nil {
		t.Fatalf("signature verification failed: %v", err)
	}
}

func TestCredsFileAuth_UserJWTReturnsErrorOnMissingFile(t *testing.T) {
	opt := natsutil.CredsFileAuth(filepath.Join(t.TempDir(), "missing.creds"))

	var options nats.Options
	if err := opt(&options); err == nil {
		t.Fatal("option apply expected error for missing creds file, got nil")
	}
}

func TestCAPemAuth_EmptyPathIsNoop(t *testing.T) {
	opt := natsutil.CAPemAuth("")

	var options nats.Options
	if err := opt(&options); err != nil {
		t.Fatalf("option apply unexpected error: %v", err)
	}

	if options.Secure {
		t.Fatal("Secure should remain false for empty CA path")
	}
}

func TestCAPemAuth_ReturnsErrorOnMissingFile(t *testing.T) {
	opt := natsutil.CAPemAuth(filepath.Join(t.TempDir(), "missing-ca.pem"))

	var options nats.Options
	if err := opt(&options); err == nil {
		t.Fatal("option apply expected error for missing CA file, got nil")
	}
}
