package guild

import (
	"database/sql"
	"errors"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/FlameInTheDark/gochat/internal/dto"
	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/FlameInTheDark/gochat/internal/idgen"
	"github.com/FlameInTheDark/gochat/internal/permissions"
)

// generateInviteCodeFromID returns an 8-char, uppercase base36 code derived from the snowflake ID
// Ensures fixed 8 chars by left-padding with '0' and trimming higher digits if necessary
func generateInviteCodeFromID(inviteID int64) string {
	const alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ" // base36, uppercase
	// encode base36
	n := uint64(inviteID)
	if n == 0 {
		return "00000000"
	}
	buf := make([]byte, 0, 16)
	for n > 0 {
		r := n % 36
		buf = append(buf, alphabet[r])
		n /= 36
	}
	// reverse
	for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}
	// ensure length 8
	if len(buf) < 8 {
		pad := make([]byte, 8-len(buf))
		for i := range pad {
			pad[i] = '0'
		}
		buf = append(pad, buf...)
	} else if len(buf) > 8 {
		// keep least significant 8 digits (right-most)
		buf = buf[len(buf)-8:]
	}
	return string(buf)
}

// ReceiveInvite
//
//	@Summary	Get invite info by code
//	@Produce	json
//	@Tags		Guild Invites
//	@Param		invite_code	path		string				true	"Invite code"
//	@Success	200			{object}	dto.InvitePreview	"Invite preview"
//	@failure	404			{string}	string				"invite not found"
//	@Router		/guild/invites/receive/{invite_code} [get]
func (e *entity) ReceiveInvite(c *fiber.Ctx) error {
	code := c.Params("invite_code")
	if len(code) != 8 {
		return fiber.NewError(fiber.StatusBadRequest, ErrInviteCodeInvalid)
	}

	inv, err := e.inv.FetchInvite(c.UserContext(), code)
	if err != nil {
		if err == sql.ErrNoRows {
			return fiber.NewError(fiber.StatusNotFound, ErrInviteNotFound)
		}
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetInvites)
	}

	g, err := e.g.GetGuildById(c.UserContext(), inv.GuildId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetGuildByID)
	}

	// Count members in the guild
	membersCount, err := e.memb.CountGuildMembers(c.UserContext(), inv.GuildId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetGuildMember)
	}

	return c.JSON(dto.InvitePreview{
		Id:           inv.InviteId,
		Code:         inv.InviteCode,
		Guild:        buildGuildDTO(&g),
		AuthorId:     inv.AuthorId,
		CreatedAt:    inv.CreatedAt,
		ExpiresAt:    inv.ExpiresAt,
		MembersCount: int(membersCount),
	})
}

// AcceptInvite
//
//	@Summary	Accept invite and join guild
//	@Produce	json
//	@Tags		Guild Invites
//	@Param		invite_code	path		string		true	"Invite code"
//	@Success	200			{object}	dto.Guild	"Joined guild"
//	@failure	404			{string}	string		"invite not found"
//	@failure	401			{string}	string		"unauthorized"
//	@Router		/guild/invites/accept/{invite_code} [post]
func (e *entity) AcceptInvite(c *fiber.Ctx) error {
	code := c.Params("invite_code")
	if len(code) != 8 {
		return fiber.NewError(fiber.StatusBadRequest, ErrInviteCodeInvalid)
	}

	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}

	inv, err := e.inv.FetchInvite(c.UserContext(), code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, ErrInviteNotFound)
		}
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetInvites)
	}

	// If already a member, just return guild
	isMember, err := e.memb.IsGuildMember(c.UserContext(), inv.GuildId, user.Id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetGuildMember)
	}
	if !isMember {
		if err := e.memb.AddMember(c.UserContext(), user.Id, inv.GuildId); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetGuildMember)
		}
	}

	g, err := e.g.GetGuildById(c.UserContext(), inv.GuildId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetGuildByID)
	}
	return c.JSON(buildGuildDTO(&g))
}

// ListInvites
//
//	@Summary	List active invites for guild
//	@Produce	json
//	@Tags		Guild Invites
//	@Param		guild_id	path		int64			true	"Guild id"
//	@Success	200			{array}		dto.GuildInvite	"List of invites"
//	@failure	401			{string}	string			"Unauthorized"
//	@Router		/guild/invites/{guild_id} [get]
func (e *entity) ListInvites(c *fiber.Ctx) error {
	guildId, err := e.parseGuildID(c)
	if err != nil {
		return err
	}

	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}

	// Require permission to create invites as proxy for managing invites
	if _, ok, perr := e.perm.GuildPerm(c.UserContext(), guildId, user.Id, permissions.PermMembershipCreateInvite); perr != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetGuildByID)
	} else if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, ErrPermissionsRequired)
	}

	invs, err := e.inv.GetGuildInvites(c.UserContext(), guildId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetInvites)
	}

	// Map to DTOs
	out := make([]dto.GuildInvite, 0, len(invs))
	for _, it := range invs {
		out = append(out, dto.GuildInvite{
			Id:        it.InviteId,
			Code:      it.InviteCode,
			GuildId:   it.GuildId,
			AuthorId:  it.AuthorId,
			CreatedAt: it.CreatedAt,
			ExpiresAt: it.ExpiresAt,
		})
	}
	return c.JSON(out)
}

// DeleteInvite
//
//	@Summary	Delete an invite by id
//	@Produce	json
//	@Tags		Guild Invites
//	@Param		guild_id	path		int64	true	"Guild id"
//	@Param		invite_id	path		int64	true	"Invite id"
//	@Success	204			{string}	string	"Deleted"
//	@failure	404			{string}	string	"invite not found"
//	@Router		/guild/invites/{guild_id}/{invite_id} [delete]
func (e *entity) DeleteInvite(c *fiber.Ctx) error {
	guildId, err := e.parseGuildID(c)
	if err != nil {
		return err
	}
	inviteIdStr := c.Params("invite_id")
	inviteId, err := strconv.ParseInt(inviteIdStr, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrIncorrectInviteID)
	}

	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}
	if _, ok, perr := e.perm.GuildPerm(c.UserContext(), guildId, user.Id, permissions.PermMembershipCreateInvite); perr != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetGuildByID)
	} else if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, ErrPermissionsRequired)
	}

	if err := e.inv.DeleteInviteByID(c.UserContext(), guildId, inviteId); err != nil {
		if err == sql.ErrNoRows {
			return fiber.NewError(fiber.StatusNotFound, ErrInviteNotFound)
		}
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToDeleteInvite)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// CreateInvite
//
//	@Summary	Create a new invite
//	@Produce	json
//	@Tags		Guild Invites
//	@Param		guild_id	path		int64				true	"Guild id"
//	@Param		request		body		CreateInviteRequest	true	"Invite options"
//	@Success	201			{object}	dto.GuildInvite		"Invite"
//	@failure	401			{string}	string				"Unauthorized"
//	@Router		/guild/invites/{guild_id} [post]
func (e *entity) CreateInvite(c *fiber.Ctx) error {
	guildId, err := e.parseGuildID(c)
	if err != nil {
		return err
	}

	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}
	if _, ok, perr := e.perm.GuildPerm(c.UserContext(), guildId, user.Id, permissions.PermMembershipCreateInvite); perr != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetGuildByID)
	} else if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, ErrPermissionsRequired)
	}

	var req CreateInviteRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToParseBody)
	}
	if err := req.Validate(); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	// Expiration handling:
	// - nil: default 7 days
	// - 0: unlimited (set far-future expiry)
	// - >0: that many seconds
	var expiresAt time.Time
	if req.ExpiresInSec == nil {
		expiresAt = time.Now().Add(7 * 24 * time.Hour)
	} else if *req.ExpiresInSec == 0 {
		expiresAt = time.Now().AddDate(100, 0, 0) // effectively unlimited
	} else {
		expiresAt = time.Now().Add(time.Duration(*req.ExpiresInSec) * time.Second)
	}

	// Derive code from snowflake-based invite ID
	invId := idgen.Next()
	code := generateInviteCodeFromID(invId)

	inv, ierr := e.inv.CreateInvite(c.UserContext(), code, invId, guildId, user.Id, expiresAt.Unix())
	if ierr != nil {
		// In the unlikely event of collision, regenerate with a new ID and retry once
		invId = idgen.Next()
		code = generateInviteCodeFromID(invId)
		inv, ierr = e.inv.CreateInvite(c.UserContext(), code, invId, guildId, user.Id, expiresAt.Unix())
		if ierr != nil {
			return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCreateInvite)
		}
	}

	return c.Status(fiber.StatusCreated).JSON(dto.GuildInvite{
		Id:        inv.InviteId,
		Code:      inv.InviteCode,
		GuildId:   inv.GuildId,
		AuthorId:  inv.AuthorId,
		CreatedAt: inv.CreatedAt,
		ExpiresAt: inv.ExpiresAt,
	})
}
