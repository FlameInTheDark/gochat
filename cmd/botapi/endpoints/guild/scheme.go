package guild

type Response struct {
	Id                 int64  `json:"id"`
	Name               string `json:"name"`
	Owner              int64  `json:"owner"`
	Public             bool   `json:"public"`
	GrantedPermissions int64  `json:"granted_permissions"`
}
