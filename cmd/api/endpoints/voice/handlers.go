package voice

import "github.com/gofiber/fiber/v2"

// GetRegions lists configured SFU regions from application config.
//
//	@Summary	List available voice regions
//	@Produce	json
//	@Tags		Voice
//	@Success	200	{object}	VoiceRegionsResponse
//	@Router		/voice/regions [get]
func (e *entity) GetRegions(c *fiber.Ctx) error {
	return c.JSON(VoiceRegionsResponse{Regions: e.regions})
}
