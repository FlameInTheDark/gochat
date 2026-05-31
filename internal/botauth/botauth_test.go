package botauth

import "testing"

func TestParseAuthorizationRequiresBotScheme(t *testing.T) {
	token, err := ParseAuthorization("Bot gcb_test")
	if err != nil {
		t.Fatalf("ParseAuthorization returned error: %v", err)
	}
	if token != "gcb_test" {
		t.Fatalf("expected token gcb_test, got %q", token)
	}

	if _, err := ParseAuthorization("Bearer gcb_test"); err == nil {
		t.Fatal("expected Bearer scheme to be rejected")
	}
}

func TestGenerateTokenHashesOpaqueToken(t *testing.T) {
	token, prefix, hash, err := GenerateToken()
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}
	if token == "" || prefix == "" || hash == "" {
		t.Fatalf("expected token, prefix, and hash to be populated")
	}
	if hash == token {
		t.Fatal("hash must not expose the raw token")
	}
	if HashToken(token) != hash {
		t.Fatal("HashToken did not reproduce generated hash")
	}
}
