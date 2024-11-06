package permissions

func CheckPermissions(perm int64, permissions ...RolePermission) bool {
	for _, p := range permissions {
		if perm&int64(p) == 0 {
			return false
		}
	}
	return true
}

func CreatePermissions(permissions ...RolePermission) int64 {
	var perm int64
	for _, p := range permissions {
		perm |= int64(p)
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
