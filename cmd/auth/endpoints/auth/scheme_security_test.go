package auth

import "testing"

func TestPasswordChangeRequestValidationAllowsNoSecondFactorPair(t *testing.T) {
	req := PasswordChangeRequest{
		CurrentPassword: "curr3nt-password",
		NewPassword:     "n3w-password",
	}

	if err := req.Validate(); err != nil {
		t.Fatalf("expected request to be valid without code pair, got %v", err)
	}
}

func TestPasswordChangeRequestValidationRequiresCompleteSecondFactorPair(t *testing.T) {
	req := PasswordChangeRequest{
		CurrentPassword: "curr3nt-password",
		NewPassword:     "n3w-password",
		CodeType:        codeTypeTOTP,
	}

	if err := req.Validate(); err == nil {
		t.Fatal("expected validation error when code is missing")
	}
}

func TestTOTPConfirmRequestValidationRejectsNonNumericOTP(t *testing.T) {
	req := TOTPConfirmRequest{
		SetupID: "setup",
		Code:    "12ab56",
	}

	if err := req.Validate(); err == nil {
		t.Fatal("expected validation error for non-numeric OTP")
	}
}
