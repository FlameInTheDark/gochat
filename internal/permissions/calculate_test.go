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
		{
			name: "add single permission to empty",
			args: args{perm: 0, permissions: []RolePermission{PermServerViewChannels}},
			want: 1,
		},
		{
			name: "add multiple permissions to empty",
			args: args{perm: 0, permissions: []RolePermission{PermServerViewChannels, PermServerManageChannels}},
			want: 3,
		},
		{
			name: "add permission to existing",
			args: args{perm: 1, permissions: []RolePermission{PermServerManageChannels}},
			want: 3,
		},
		{
			name: "add already existing permission",
			args: args{perm: 1, permissions: []RolePermission{PermServerViewChannels}},
			want: 1,
		},
		{
			name: "add no permissions",
			args: args{perm: 5, permissions: []RolePermission{}},
			want: 5,
		},
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
		{
			name: "add single role to empty",
			args: args{perm: 0, add: []int64{5}},
			want: 5,
		},
		{
			name: "add multiple roles",
			args: args{perm: 1, add: []int64{2, 4}},
			want: 7,
		},
		{
			name: "add overlapping roles",
			args: args{perm: 3, add: []int64{5, 7}},
			want: 7,
		},
		{
			name: "add no roles",
			args: args{perm: 10, add: []int64{}},
			want: 10,
		},
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
		{
			name: "allow all permissions",
			args: args{perm: 15, perms: []int64{15}},
			want: 15,
		},
		{
			name: "allow subset of permissions",
			args: args{perm: 15, perms: []int64{7}},
			want: 7,
		},
		{
			name: "allow intersection of multiple",
			args: args{perm: 15, perms: []int64{7, 3}},
			want: 3,
		},
		{
			name: "allow with no overlap",
			args: args{perm: 8, perms: []int64{4}},
			want: 0,
		},
		{
			name: "allow with empty perms",
			args: args{perm: 15, perms: []int64{}},
			want: 15,
		},
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
		{
			name: "has all requested permissions",
			args: args{perm: 7, permissions: []RolePermission{PermServerViewChannels, PermServerManageChannels}},
			want: true,
		},
		{
			name: "missing one permission",
			args: args{perm: 1, permissions: []RolePermission{PermServerViewChannels, PermServerManageChannels}},
			want: false,
		},
		{
			name: "administrator overrides all",
			args: args{perm: int64(PermAdministrator), permissions: []RolePermission{PermServerViewChannels, PermServerManageChannels}},
			want: true,
		},
		{
			name: "check with no permissions requested",
			args: args{perm: 5, permissions: []RolePermission{}},
			want: true,
		},
		{
			name: "empty permissions",
			args: args{perm: 0, permissions: []RolePermission{PermServerViewChannels}},
			want: false,
		},
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
		{
			name: "has one of requested permissions",
			args: args{perm: 1, permissions: []RolePermission{PermServerViewChannels, PermServerManageChannels}},
			want: true,
		},
		{
			name: "has all requested permissions",
			args: args{perm: 3, permissions: []RolePermission{PermServerViewChannels, PermServerManageChannels}},
			want: true,
		},
		{
			name: "has none of requested permissions",
			args: args{perm: 8, permissions: []RolePermission{PermServerViewChannels, PermServerManageChannels}},
			want: false,
		},
		{
			name: "empty permissions",
			args: args{perm: 5, permissions: []RolePermission{}},
			want: false,
		},
		{
			name: "zero perm with requests",
			args: args{perm: 0, permissions: []RolePermission{PermServerViewChannels}},
			want: false,
		},
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
		{
			name: "create from single permission",
			args: args{permissions: []RolePermission{PermServerViewChannels}},
			want: 1,
		},
		{
			name: "create from multiple permissions",
			args: args{permissions: []RolePermission{PermServerViewChannels, PermServerManageChannels, PermServerManageRoles}},
			want: 7,
		},
		{
			name: "create from no permissions",
			args: args{permissions: []RolePermission{}},
			want: 0,
		},
		{
			name: "create with duplicate permissions",
			args: args{permissions: []RolePermission{PermServerViewChannels, PermServerViewChannels}},
			want: 1,
		},
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
		{
			name: "has overlap",
			args: args{first: 7, second: 3},
			want: true,
		},
		{
			name: "identical permissions",
			args: args{first: 5, second: 5},
			want: true,
		},
		{
			name: "no overlap",
			args: args{first: 8, second: 4},
			want: false,
		},
		{
			name: "one empty",
			args: args{first: 0, second: 5},
			want: false,
		},
		{
			name: "both empty",
			args: args{first: 0, second: 0},
			want: false,
		},
		{
			name: "partial overlap",
			args: args{first: 15, second: 12},
			want: true,
		},
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
		{
			name: "remove single permission",
			args: args{perm: 3, permissions: []RolePermission{PermServerViewChannels}},
			want: 2,
		},
		{
			name: "remove multiple permissions",
			args: args{perm: 7, permissions: []RolePermission{PermServerViewChannels, PermServerManageChannels}},
			want: 4,
		},
		{
			name: "remove non-existent permission",
			args: args{perm: 1, permissions: []RolePermission{PermServerManageChannels}},
			want: 1,
		},
		{
			name: "remove all permissions",
			args: args{perm: 3, permissions: []RolePermission{PermServerViewChannels, PermServerManageChannels}},
			want: 0,
		},
		{
			name: "remove no permissions",
			args: args{perm: 5, permissions: []RolePermission{}},
			want: 5,
		},
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
		{
			name: "subtract single bit",
			args: args{perm: 7, remove: 1},
			want: 6,
		},
		{
			name: "subtract multiple bits",
			args: args{perm: 15, remove: 3},
			want: 12,
		},
		{
			name: "subtract non-existent bits",
			args: args{perm: 5, remove: 2},
			want: 5,
		},
		{
			name: "subtract all bits",
			args: args{perm: 7, remove: 7},
			want: 0,
		},
		{
			name: "subtract from empty",
			args: args{perm: 0, remove: 5},
			want: 0,
		},
		{
			name: "subtract zero",
			args: args{perm: 10, remove: 0},
			want: 10,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SubtractRoles(tt.args.perm, tt.args.remove); got != tt.want {
				t.Errorf("SubtractRoles() = %v, want %v", got, tt.want)
			}
		})
	}
}
