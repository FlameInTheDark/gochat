package permissions

import "testing"

func TestAddPermissions(t *testing.T) {
	type args struct {
		perm        int64
		permissions []RolePermission
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AddPermissions(tt.args.perm, tt.args.permissions...); got != tt.want {
				t.Errorf("AddPermissions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddRoles(t *testing.T) {
	type args struct {
		perm int64
		add  []int64
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AddRoles(tt.args.perm, tt.args.add...); got != tt.want {
				t.Errorf("AddRoles() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAllowing(t *testing.T) {
	type args struct {
		perm  int64
		perms []int64
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Allowing(tt.args.perm, tt.args.perms...); got != tt.want {
				t.Errorf("Allowing() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckPermissions(t *testing.T) {
	type args struct {
		perm        int64
		permissions []RolePermission
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "allow read and see", args: args{perm: 7927905, permissions: []RolePermission{PermTextReadMessageHistory, PermServerViewChannels}}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckPermissions(tt.args.perm, tt.args.permissions...); got != tt.want {
				t.Errorf("CheckPermissions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckPermissionsAny(t *testing.T) {
	type args struct {
		perm        int64
		permissions []RolePermission
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckPermissionsAny(tt.args.perm, tt.args.permissions...); got != tt.want {
				t.Errorf("CheckPermissionsAny() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreatePermissions(t *testing.T) {
	type args struct {
		permissions []RolePermission
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CreatePermissions(tt.args.permissions...); got != tt.want {
				t.Errorf("CreatePermissions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasOverlap(t *testing.T) {
	type args struct {
		first  int64
		second int64
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasOverlap(tt.args.first, tt.args.second); got != tt.want {
				t.Errorf("HasOverlap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRemovePermissions(t *testing.T) {
	type args struct {
		perm        int64
		permissions []RolePermission
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RemovePermissions(tt.args.perm, tt.args.permissions...); got != tt.want {
				t.Errorf("RemovePermissions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSubtractRoles(t *testing.T) {
	type args struct {
		perm   int64
		remove int64
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SubtractRoles(tt.args.perm, tt.args.remove); got != tt.want {
				t.Errorf("SubtractRoles() = %v, want %v", got, tt.want)
			}
		})
	}
}
