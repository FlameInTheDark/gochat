package mls

import (
	"encoding/base64"
	"encoding/binary"
	"strings"
	"testing"
)

var keyPackageFixtures = map[int64]string{
	501: "AAEAAkBBBBoSujDDF1qzYBAnD2jJ2hecTzbOleqEGvvlnPPjJMlb/FMXz9LmD7g4xEeot6IC8N74fhfpN2yGdAPK45esd2tAQQSYZLJhQdGPOfVUtEycQjVRAoq25NBxzaWtqngwkhEthAsFKpIGDX6DAvVS2VUj0O7GtEvuDoMiHcM4YmIRvZfgQEEEQWD9v0QUlBKHzjrNhqX5JduWo7BvOS74KMGD5lCdfH7BYhDQxZum4F2qPO2CTUUAl2b/EOCcdaVZ/lM16RsVpgABCAAAAAAAAAH1AgABAgACAAACAAEBAAAAAAAAAAD//////////wBARzBFAiEA6BtADgVSJ3kr/LwXt4fFVZcBqabPSfqmnA4JqWQz+E0CICnwze3Cserp1dEYe8awDnfTJjqzQeaJlFNJhPJe2qWfAEBHMEUCIQCL2i+ux9W9mrV2KHYvQrf6iFu1sEGi2cf2iqKSf634tQIgNeLXtd1a+szYhtWJXTRTcd+hBxWLAOqn03r9OY4WA4A=",
	502: "AAEAAkBBBDMqRkt2euRV4MrR7y0sWgQNqDfk3NLcK9O/ukaDOkykA1xeWp0r23XQutI9Usy4etMYk/uWWWuRX67nrh6nZPRAQQR3Uln1qePIdu6H4/BhHY1YZAiEdTyKWr0FjM92kjDinkuxemPStmZL5j/qTvd0U+KHlFNWF6rymTCuDt+4n9HrQEEEk7fWAbDxho33M5YHDfzMtxvVInGFW+K6KrD8AJpNOmaA6i9i8TjP6aqNs2v8XQzkVKpAds4Rsbk1rgOR2HXvSAABCAAAAAAAAAH2AgABAgACAAACAAEBAAAAAAAAAAD//////////wBARzBFAiAjcpoWuZiQoVWGNmUVc9thxuNIUzsd5l7QkY9QbRf+KQIhAKAY6Cgv5ACUZ4Mu2ofgsnYVR8KKmkdCgybIXuzTBILYAEBGMEQCIE/2omCFfGN3m2xAaFLkA7bK/UYpH+63jQNyVOU/2remAiAy8V1jYa88NFdE2HoTFCfRDcDhcP7TzKTptfDT/PDsnw==",
}

func mustFixtureKeyPackage(t *testing.T, userID int64) []byte {
	t.Helper()

	raw, ok := keyPackageFixtures[userID]
	if !ok {
		t.Fatalf("missing fixture for user %d", userID)
	}
	decoded, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		t.Fatalf("decode key package fixture: %v", err)
	}
	return decoded
}

func TestParseAndValidateKeyPackage(t *testing.T) {
	keyPackage, err := ParseAndValidateKeyPackage(mustFixtureKeyPackage(t, 501), 501)
	if err != nil {
		t.Fatalf("ParseAndValidateKeyPackage: %v", err)
	}
	if keyPackage.UserID != 501 {
		t.Fatalf("UserID = %d, want 501", keyPackage.UserID)
	}
	if got := int64(binary.BigEndian.Uint64(keyPackage.Identity)); got != 501 {
		t.Fatalf("identity user id = %d, want 501", got)
	}
	if len(keyPackage.SignatureKey) == 0 {
		t.Fatal("expected signature key bytes")
	}
	if len(keyPackage.Inner) == 0 || len(keyPackage.Raw) == 0 {
		t.Fatal("expected raw and inner key package bytes")
	}
}

func TestParseAndValidateKeyPackageRejectsWrongUser(t *testing.T) {
	_, err := ParseAndValidateKeyPackage(mustFixtureKeyPackage(t, 501), 999)
	if err == nil {
		t.Fatal("expected wrong-user key package validation error")
	}
	if !strings.Contains(err.Error(), "does not match authenticated user") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBuildExternalAddProposal(t *testing.T) {
	sender, err := NewExternalSender([]byte("gochat-sfu"))
	if err != nil {
		t.Fatalf("NewExternalSender: %v", err)
	}
	keyPackage, err := ParseAndValidateKeyPackage(mustFixtureKeyPackage(t, 502), 502)
	if err != nil {
		t.Fatalf("ParseAndValidateKeyPackage: %v", err)
	}

	proposal, ref, err := BuildExternalAddProposal(100, 0, 7, sender, keyPackage)
	if err != nil {
		t.Fatalf("BuildExternalAddProposal: %v", err)
	}
	if len(ref) == 0 {
		t.Fatal("expected non-empty proposal ref")
	}
	if got := binary.BigEndian.Uint16(proposal[:2]); got != protocolVersionMLS10 {
		t.Fatalf("proposal version = %d, want %d", got, protocolVersionMLS10)
	}
	if got := binary.BigEndian.Uint16(proposal[2:4]); got != wireFormatPublicMessage {
		t.Fatalf("proposal wire format = %d, want %d", got, wireFormatPublicMessage)
	}

	offset := 4
	groupID, err := readOpaque(proposal, &offset)
	if err != nil {
		t.Fatalf("read group id: %v", err)
	}
	if got := int64(binary.BigEndian.Uint64(groupID)); got != 100 {
		t.Fatalf("group id = %d, want 100", got)
	}
	if got := readUint64(proposal, &offset); got != 0 {
		t.Fatalf("epoch = %d, want 0", got)
	}
	if got := proposal[offset]; got != senderTypeExternal {
		t.Fatalf("sender type = %d, want %d", got, senderTypeExternal)
	}
	offset++
	if got := binary.BigEndian.Uint32(proposal[offset : offset+4]); got != 7 {
		t.Fatalf("sender index = %d, want 7", got)
	}
	offset += 4
	if _, err := readOpaque(proposal, &offset); err != nil {
		t.Fatalf("read authenticated data: %v", err)
	}
	if got := proposal[offset]; got != contentTypeProposal {
		t.Fatalf("content type = %d, want %d", got, contentTypeProposal)
	}
	offset++
	if got := readUint16(proposal, &offset); got != proposalTypeAdd {
		t.Fatalf("proposal type = %d, want %d", got, proposalTypeAdd)
	}
	if remaining := proposal[offset:]; len(remaining) < len(keyPackage.Inner) || string(remaining[:len(keyPackage.Inner)]) != string(keyPackage.Inner) {
		t.Fatal("proposal does not contain the expected key package payload")
	}
	offset += len(keyPackage.Inner)
	if _, err := readOpaque(proposal, &offset); err != nil {
		t.Fatalf("read auth data signature: %v", err)
	}
	if offset != len(proposal) {
		t.Fatalf("proposal contains trailing bytes: %d", len(proposal)-offset)
	}
}
