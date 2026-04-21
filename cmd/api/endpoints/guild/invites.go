package guild

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"log/slog"
	"math/big"
	"strconv"
	"time"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/FlameInTheDark/gochat/internal/dto"
	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/FlameInTheDark/gochat/internal/idgen"
	"github.com/FlameInTheDark/gochat/internal/messageposition"
	"github.com/FlameInTheDark/gochat/internal/mq"
	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
	"github.com/FlameInTheDark/gochat/internal/observability"
	"github.com/FlameInTheDark/gochat/internal/permissions"
	"github.com/gofiber/fiber/v2"
)

const (
	inviteCodeLength        = 8
	createInviteTTLCap      = time.Hour
	defaultInviteTTL        = 7 * 24 * time.Hour
	unlimitedInviteLifetime = 100
)

func generateInviteCode() (string, error) {
	const alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ" // base36, uppercase

	buf := make([]byte, inviteCodeLength)
	max := big.NewInt(int64(len(alphabet)))
	for i := range buf {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		buf[i] = alphabet[n.Int64()]
	}
	return string(buf), nil
}

func (e *entity) canManageInvites(ctx context.Context, guildId, userId int64) (bool, error) {
	_, ok, err := e.perm.GuildPerm(ctx, guildId, userId, permissions.PermAdministrator)
	return ok, err
}

func (e *entity) canCreateInvite(ctx context.Context, guildId, userId int64) (bool, error) {
	_, ok, err := e.perm.GuildPerm(ctx, guildId, userId, permissions.PermMembershipCreateInvite)
	return ok, err
}

func inviteExpiryFromRequest(req CreateInviteRequest, allowExtendedTTL bool) time.Time {
	now := time.Now()
	if req.ExpiresInSec == nil {
		if allowExtendedTTL {
			return now.Add(defaultInviteTTL)
		}
		return now.Add(createInviteTTLCap)
	}

	if *req.ExpiresInSec == 0 {
		if allowExtendedTTL {
			return now.AddDate(unlimitedInviteLifetime, 0, 0)
		}
		return now.Add(createInviteTTLCap)
	}

	ttl := time.Duration(*req.ExpiresInSec) * time.Second
	if !allowExtendedTTL && ttl > createInviteTTLCap {
		ttl = createInviteTTLCap
	}
	return now.Add(ttl)
}

// ReceiveInvite
//
//	@Summary	Get invite info by code
//	@Produce	json
//	@Tags		Guild Invites
//	@Param		invite_code	path		string				true	"Invite code"	example(PWBJ124G)
//	@Success	200			{object}	dto.InvitePreview	"Invite preview"
//	@failure	404			{string}	string				"invite not found"
//	@Router		/guild/invites/receive/{invite_code} [get]
func (e *entity) ReceiveInvite(c *fiber.Ctx) error {
	code := c.Params("invite_code")
	if len(code) != inviteCodeLength {
		return fiber.NewError(fiber.StatusBadRequest, ErrInviteCodeInvalid)
	}

	inv, err := e.inv.FetchInvite(c.UserContext(), code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, ErrInviteNotFound)
		}
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetInvites)
	}

	g, err := e.g.GetGuildById(c.UserContext(), inv.GuildId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetGuildByID)
	}

	membersCount, err := e.memb.CountGuildMembers(c.UserContext(), inv.GuildId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetGuildMember)
	}

	return c.JSON(dto.InvitePreview{
		Id:           inv.InviteId,
		Code:         inv.InviteCode,
		Guild:        e.dtoGuildWithIcon(c, &g),
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
//	@Param		invite_code	path		string		true	"Invite code"	example(PWBJ124G)
//	@Success	200			{object}	dto.Guild	"Joined guild"
//	@failure	404			{string}	string		"invite not found"
//	@failure	401			{string}	string		"unauthorized"
//	@Router		/guild/invites/accept/{invite_code} [post]
func (e *entity) AcceptInvite(c *fiber.Ctx) error {
	code := c.Params("invite_code")
	if len(code) != inviteCodeLength {
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

	if banned, err := e.isGuildUserBanned(c.UserContext(), inv.GuildId, user.Id); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCheckGuildBan)
	} else if banned {
		return fiber.NewError(fiber.StatusForbidden, ErrUserIsBanned)
	}

	u, err := e.user.GetUserById(c.UserContext(), user.Id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetUser)
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

	disc, err := e.disc.GetDiscriminatorByUserId(c.UserContext(), u.Id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetDiscriminator)
	}

	g, err := e.g.GetGuildById(c.UserContext(), inv.GuildId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetGuildByID)
	}

	asyncCtx := observability.BackgroundFromContext(c.UserContext())
	asyncLog := observability.LoggerWithContext(asyncCtx, e.log)
	go func() {
		err := mq.SendGuildUpdate(asyncCtx, e.mqt, inv.GuildId, &mqmsg.AddGuildMember{
			GuildId: inv.GuildId,
			UserId:  u.Id,
			Member: dto.Member{
				User:     userToDTO(u, disc.Discriminator),
				Username: nil,
				Avatar:   nil,
				JoinAt:   time.Now(),
				Roles:    nil,
			},
		})
		if err != nil {
			asyncLog.Error("unable to send add guild member event", slog.String("error", err.Error()))
		}
		if g.SystemMessages != nil {
			if _, err := e.ch.GetChannel(asyncCtx, *g.SystemMessages); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return
				}
				asyncLog.Error("unable to get system channel", slog.String("error", err.Error()))
				return
			}
			msgid := idgen.Next()
			position, err := messageposition.Next(asyncCtx, e.cache, e.ch, *g.SystemMessages)
			if err != nil {
				asyncLog.Error("unable to allocate join message position", slog.String("error", err.Error()))
				return
			}
			err = e.msg.CreateSystemMessage(asyncCtx, msgid, *g.SystemMessages, user.Id, "", model.MessageTypeJoin, position)
			if err != nil {
				asyncLog.Error("unable to send system user join message", slog.String("error", err.Error()))
				return
			}
			err = e.ch.SetLastMessage(asyncCtx, *g.SystemMessages, msgid)
			if err != nil {
				asyncLog.Error("unable to set last message id", slog.String("error", err.Error()))
			}
			if err := mq.SendChannelMessage(asyncCtx, e.mqt, *g.SystemMessages, &mqmsg.CreateMessage{
				GuildId: &g.Id,
				Message: dto.Message{
					Id:        msgid,
					ChannelId: *g.SystemMessages,
					Author:    userToDTO(u, disc.Discriminator),
					Position:  guildOptionalInt64(position),
					Type:      int(model.MessageTypeJoin),
				},
			}); err != nil {
				asyncLog.Error("unable to send join message event", slog.String("error", err.Error()))
			}
			if err := e.imq.IndexMessageContext(asyncCtx, dto.IndexMessage{
				MessageId: msgid,
				UserId:    u.Id,
				ChannelId: *g.SystemMessages,
				GuildId:   &g.Id,
				Type:      int(model.MessageTypeJoin),
			}); err != nil {
				asyncLog.Error("failed to send index message event",
					"message_id", msgid,
					"error", err.Error())
			}
		}
	}()

	return c.JSON(e.dtoGuildWithIcon(c, &g))
}

// ListInvites
//
//	@Summary	List active invites for guild
//	@Produce	json
//	@Tags		Guild Invites
//	@Param		guild_id	path		int64			true	"Guild id"	example(2230469276416868352)
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

	if ok, perr := e.canManageInvites(c.UserContext(), guildId, user.Id); perr != nil {
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
//	@Param		guild_id	path		int64	true	"Guild id"	example(2230469276416868352)
//	@Param		invite_id	path		int64	true	"Invite id"	example(2230469276416868352)
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
	if ok, perr := e.canManageInvites(c.UserContext(), guildId, user.Id); perr != nil {
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
//	@Param		guild_id	path		int64				true	"Guild id"	example(2230469276416868352)
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
	allowExtendedTTL, perr := e.canManageInvites(c.UserContext(), guildId, user.Id)
	if perr != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetGuildByID)
	}
	if !allowExtendedTTL {
		if ok, checkErr := e.canCreateInvite(c.UserContext(), guildId, user.Id); checkErr != nil {
			return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetGuildByID)
		} else if !ok {
			return fiber.NewError(fiber.StatusUnauthorized, ErrPermissionsRequired)
		}
	}

	var req CreateInviteRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToParseBody)
	}
	if err := req.Validate(); err != nil {
		return badRequestValidationError(err)
	}

	expiresAt := inviteExpiryFromRequest(req, allowExtendedTTL)

	var inv model.GuildInvite
	var ierr error
	for attempt := 0; attempt < 5; attempt++ {
		invId := idgen.Next()
		code, codeErr := generateInviteCode()
		if codeErr != nil {
			return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCreateInvite)
		}
		inv, ierr = e.inv.CreateInvite(c.UserContext(), code, invId, guildId, user.Id, expiresAt.Unix())
		if ierr == nil {
			break
		}
	}
	if ierr != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCreateInvite)
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
