package dto

type Guild struct {
	Id     int64  `json:"id"`
	Name   string `json:"name"`
	Icon   *int64 `json:"icon"`
	Owner  bool   `json:"owner"`
	Public bool   `json:"public"`
}
