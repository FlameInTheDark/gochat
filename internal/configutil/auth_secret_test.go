package configutil

import "testing"

func TestValidateAuthSecret(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		secret  string
		wantErr bool
	}{
		{
			name:    "empty",
			secret:  "",
			wantErr: true,
		},
		{
			name:    "whitespace",
			secret:  "   \t\n",
			wantErr: true,
		},
		{
			name:    "dev placeholder allowed",
			secret:  "change_me_before_use_it_in_production",
			wantErr: false,
		},
		{
			name:    "random value allowed",
			secret:  "super-secret-for-tests",
			wantErr: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateAuthSecret(tc.secret)
			if (err != nil) != tc.wantErr {
				t.Fatalf("ValidateAuthSecret() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestValidateAuthSecretWithMode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		secret      string
		enforcement string
		wantErr     bool
	}{
		{
			name:        "warn mode keeps compatibility",
			secret:      "change_me_before_use_it_in_production",
			enforcement: AuthSecretEnforcementWarn,
			wantErr:     false,
		},
		{
			name:        "strict rejects short secrets",
			secret:      "short-secret",
			enforcement: AuthSecretEnforcementStrict,
			wantErr:     true,
		},
		{
			name:        "strict rejects placeholder secrets",
			secret:      "change_me_before_use_it_in_production",
			enforcement: AuthSecretEnforcementStrict,
			wantErr:     true,
		},
		{
			name:        "strict accepts long random secrets",
			secret:      "0123456789abcdef0123456789abcdef",
			enforcement: AuthSecretEnforcementStrict,
			wantErr:     false,
		},
		{
			name:        "invalid mode rejected",
			secret:      "0123456789abcdef0123456789abcdef",
			enforcement: "invalid",
			wantErr:     true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateAuthSecretWithMode(tc.secret, tc.enforcement)
			if (err != nil) != tc.wantErr {
				t.Fatalf("ValidateAuthSecretWithMode() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}
