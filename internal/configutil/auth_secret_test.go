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
