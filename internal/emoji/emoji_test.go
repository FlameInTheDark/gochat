package emoji

import "testing"

func TestSelectClosestVariant(t *testing.T) {
	tests := []struct {
		name string
		size int
		want string
	}{
		{name: "default master", size: 0, want: "master"},
		{name: "exact 44", size: 44, want: "44"},
		{name: "exact 96", size: 96, want: "96"},
		{name: "closest low", size: 50, want: "44"},
		{name: "tie prefers larger", size: 70, want: "96"},
		{name: "closest master", size: 120, want: "master"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SelectClosestVariant(tt.size); got != tt.want {
				t.Fatalf("SelectClosestVariant(%d) = %q, want %q", tt.size, got, tt.want)
			}
		})
	}
}
