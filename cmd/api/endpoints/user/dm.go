package user

import (
    "github.com/gofiber/fiber/v2"

    "github.com/FlameInTheDark/gochat/internal/dto"
    "github.com/FlameInTheDark/gochat/internal/helper"
)

// GetMyDMChannels
//
//	@Summary	List all DM and Group DM channels for current user
//	@Produce	json
//	@Tags		User
//	@Success	200	{array}		dto.Channel	"Channels"
//	@failure	400	{string}	string		"Bad request"
//	@failure	500	{string}	string		"Internal server error"
//	@Router		/user/me/channels [get]
func (e *entity) GetMyDMChannels(c *fiber.Ctx) error {
	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}

	// Gather DM channels
	dms, err := e.dm.GetUserDmChannels(c.UserContext(), user.Id)
	if err != nil {
		return helper.HttpDbError(err, ErrUnableToGetChannel)
	}
	// Gather Group DM channels
	gdms, err := e.gdm.GetUserGroupDmChannels(c.UserContext(), user.Id)
	if err != nil {
		return helper.HttpDbError(err, ErrUnableToGetChannel)
	}

	ids := make([]int64, 0, len(dms)+len(gdms))
	for _, d := range dms {
		ids = append(ids, d.ChannelId)
	}
	for _, g := range gdms {
		ids = append(ids, g.ChannelId)
	}

	if len(ids) == 0 {
		return c.JSON([]dto.Channel{})
	}

    // Build participant map for 1:1 DMs
    participants := make(map[int64]int64, len(dms)) // channelId -> participantId
    for _, d := range dms {
        participants[d.ChannelId] = d.ParticipantId
    }

    // Channels data
    chs, err := e.ch.GetChannelsBulk(c.UserContext(), ids)
    if err != nil {
        return helper.HttpDbError(err, ErrUnableToGetChannel)
    }
	// Last messages from CQL (DM store)
	last, err := e.dmlm.GetChannelsMessages(c.UserContext(), ids)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "unable to get dm last messages")
	}

    result := make([]dto.Channel, len(chs))
    for i := range chs {
        var pid *int64
        if v, ok := participants[chs[i].Id]; ok {
            pid = &v
        }
        result[i] = dmChannelModelToDTO(&chs[i], last, pid)
    }
    return c.JSON(result)
}
