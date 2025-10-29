package discovery

import (
	"context"
)

// Instance represents a discoverable SFU server.
type Instance struct {
	ID        string `json:"id"`
	Region    string `json:"region"`
	URL       string `json:"url"`  // public wss://.../signal
	Load      int64  `json:"load"` // simple load metric (active peers)
	UpdatedAt int64  `json:"updated_at"`
}

// Manager is a minimal discovery interface.
type Manager interface {
	// Register or refresh this instance in discovery for the given region.
	Register(ctx context.Context, region string, inst Instance) error
	// List currently available instances for a region.
	List(ctx context.Context, region string) ([]Instance, error)
	// Regions returns known regions. Not all managers implement discovery of regions; may return empty.
	Regions(ctx context.Context) ([]string, error)
}
