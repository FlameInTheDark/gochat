package permissions

func CheckPermissions(perm int64, permissions ...RolePermission) bool {
	for _, p := range permissions {
		if perm&int64(p) == 0 {
			return false
		}
	}
	return true
}

func CheckPermissionsAny(perm int64, permissions ...RolePermission) bool {
	for _, p := range permissions {
		if perm&int64(p) != 0 {
			return true
		}
	}
	return false
}

func CreatePermissions(permissions ...RolePermission) int64 {
	var perm int64
	for _, p := range permissions {
		perm |= int64(p)
	}
	return perm
}

func Allowing(perm int64, perms ...int64) int64 {
	for _, p := range perms {
		perm &= p
	}
	return perm
}

func AddPermissions(perm int64, permissions ...RolePermission) int64 {
	for _, p := range permissions {
		perm |= int64(p)
	}
	return perm
}

func RemovePermissions(perm int64, permissions ...RolePermission) int64 {
	for _, p := range permissions {
		perm &= ^int64(p)
	}
	return perm
}

func SubtractRoles(perm int64, remove int64) int64 {
	return perm &^ remove
}

func AddRoles(perm int64, add ...int64) int64 {
	for _, p := range add {
		perm |= p
	}
	return perm
}

func HasOverlap(first, second int64) bool {
	return (first & second) != 0
}
