package sfu

import (
	"fmt"
)

type HeartbeatRequest struct {
	ID     string `json:"id"`
	Region string `json:"region"`
	URL    string `json:"url"`
	Load   int64  `json:"load"`
}

func (r HeartbeatRequest) Validate() error {
	if r.ID == "" {
		return fmt.Errorf("id is required")
	}
	if r.Region == "" {
		return fmt.Errorf("region is required")
	}
	if r.URL == "" {
		return fmt.Errorf("url is required")
	}
	return nil
}
