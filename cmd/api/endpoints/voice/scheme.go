package voice

// Region represents a voice region with id and human-friendly name.
type Region struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// VoiceRegionsResponse wraps list of regions.
type VoiceRegionsResponse struct {
	Regions []Region `json:"regions"`
}
