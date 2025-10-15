package user

import (
	"context"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/FlameInTheDark/gochat/internal/dto"
	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
)

// GetFriends
//
//	@Summary	Get my friends
//	@Produce	json
//	@Tags		User
//	@Success	200	{array}		dto.User	"Friends list"
//	@failure	400	{string}	string		"Bad request"
//	@failure	500	{string}	string		"Internal server error"
//	@Router		/user/me/friends [get]
func (e *entity) GetFriends(c *fiber.Ctx) error {
	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}

	frs, err := e.fr.GetFriends(c.UserContext(), user.Id)
	if err != nil {
		return helper.HttpDbError(err, ErrUnableToGetFriends)
	}
	if len(frs) == 0 {
		return c.JSON([]dto.User{})
	}

	ids := make([]int64, len(frs))
	for i, f := range frs {
		ids[i] = f.FriendID
	}

	users, err := e.user.GetUsersList(c.UserContext(), ids)
	if err != nil {
		return helper.HttpDbError(err, ErrUnableToGetUser)
	}
	discs, err := e.disc.GetDiscriminatorsByUserIDs(c.UserContext(), ids)
	if err != nil {
		return helper.HttpDbError(err, ErrUnableToGetDiscriminator)
	}

	return c.JSON(usersWithDiscriminators(users, discs))
}

// GetOrCreateFriendDM
//
//	@Summary	Get or create DM with a user
//	@Produce	json
//	@Tags		User
//	@Param		user_id	path		int64		true	"User id"	example(2230469276416868352)
//	@Success	200		{object}	dto.Channel	"DM channel"
//	@failure	400		{string}	string		"Bad request"
//	@failure	404		{string}	string		"User not found"
//	@failure	500		{string}	string		"Internal server error"
//	@Router		/user/me/friends/{user_id} [get]
func (e *entity) GetOrCreateFriendDM(c *fiber.Ctx) error {
	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}

	idStr := c.Params("user_id")
	recipientId, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToParseID)
	}

	if recipientId == user.Id {
		return fiber.NewError(fiber.StatusBadRequest, ErrBadRequest)
	}

	// Ensure recipient exists
	if _, err := e.validateRecipient(c, recipientId); err != nil {
		return err
	}

	// Check if DM exists
	if ch, err := e.findExistingDMChannel(c, user.Id, recipientId); err != nil {
		return err
	} else if ch != nil {
		return c.JSON(*ch)
	}

	// Create DM
	channel, err := e.createNewDMChannel(c, user.Id, recipientId)
	if err != nil {
		return err
	}
	return c.JSON(channel)
}

// CreateFriendRequest
//
//	@Summary	Send a friend request by discriminator
//	@Accept		json
//	@Produce	json
//	@Tags		User
//	@Param		request	body		CreateFriendRequestRequest	true	"Friend request"
//	@Success	200		{string}	string						"ok"
//	@failure	400		{string}	string						"Bad request"
//	@failure	500		{string}	string						"Internal server error"
//	@Router		/user/me/friends [post]
func (e *entity) CreateFriendRequest(c *fiber.Ctx) error {
	var req CreateFriendRequestRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToParseRequestBody)
	}
	if err := req.Validate(); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	me, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}

	disc, err := e.disc.GetUserIdByDiscriminator(c.UserContext(), req.Discriminator)
	if err != nil {
		return helper.HttpDbError(err, ErrUnableToGetDiscriminator)
	}

	if disc.UserId == me.Id {
		return fiber.NewError(fiber.StatusBadRequest, ErrBadRequest)
	}

	// Send friend request (recipient is disc.UserId)
	if err := e.fr.CreateFriendRequest(c.UserContext(), me.Id, disc.UserId); err != nil {
		return helper.HttpDbError(err, ErrUnableToCreateFriendRequest)
	}
	// Emit WS event to recipient about incoming request (best-effort)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()
		if u, uerr := e.user.GetUserById(ctx, me.Id); uerr == nil {
			if d, derr := e.disc.GetDiscriminatorByUserId(ctx, me.Id); derr == nil {
				_ = e.mqt.SendUserUpdate(disc.UserId, &mqmsg.IncomingFriendRequest{
					From: mqmsg.UserBrief{Id: u.Id, Name: u.Name, Discriminator: d.Discriminator, Avatar: u.Avatar},
				})
			}
		}
	}()

	return c.SendStatus(fiber.StatusOK)
}

// Unfriend
//
//	@Summary	Remove user from friends
//	@Accept		json
//	@Produce	json
//	@Tags		User
//	@Param		request	body		UnfriendRequest	true	"Unfriend"
//	@Success	200		{string}	string			"ok"
//	@failure	400		{string}	string			"Bad request"
//	@failure	500		{string}	string			"Internal server error"
//	@Router		/user/me/friends [delete]
func (e *entity) Unfriend(c *fiber.Ctx) error {
	var req UnfriendRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToParseRequestBody)
	}
	if err := req.Validate(); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	me, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}

	if req.UserId == me.Id {
		return fiber.NewError(fiber.StatusBadRequest, ErrBadRequest)
	}

	if err := e.fr.RemoveFriend(c.UserContext(), me.Id, req.UserId); err != nil {
		return helper.HttpDbError(err, ErrUnableToRemoveFriend)
	}

	// Emit WS events to both users about friend removal (best-effort)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()
		// Notify me about removed friend
		if u, err := e.user.GetUserById(ctx, req.UserId); err == nil {
			if d, derr := e.disc.GetDiscriminatorByUserId(ctx, req.UserId); derr == nil {
				_ = e.mqt.SendUserUpdate(me.Id, &mqmsg.FriendRemoved{Friend: mqmsg.UserBrief{Id: u.Id, Name: u.Name, Discriminator: d.Discriminator, Avatar: u.Avatar}})
			}
		}
		// Notify other user about me
		if u, err := e.user.GetUserById(ctx, me.Id); err == nil {
			if d, derr := e.disc.GetDiscriminatorByUserId(ctx, me.Id); derr == nil {
				_ = e.mqt.SendUserUpdate(req.UserId, &mqmsg.FriendRemoved{Friend: mqmsg.UserBrief{Id: u.Id, Name: u.Name, Discriminator: d.Discriminator, Avatar: u.Avatar}})
			}
		}
	}()

	return c.SendStatus(fiber.StatusOK)
}

// GetFriendRequests
//
//	@Summary	Get incoming friend requests
//	@Produce	json
//	@Tags		User
//	@Success	200	{array}		dto.User	"Request senders"
//	@failure	400	{string}	string		"Bad request"
//	@failure	500	{string}	string		"Internal server error"
//	@Router		/user/me/friends/requests [get]
func (e *entity) GetFriendRequests(c *fiber.Ctx) error {
	me, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}
	reqs, err := e.fr.GetFriendRequests(c.UserContext(), me.Id)
	if err != nil {
		return helper.HttpDbError(err, ErrUnableToGetFriendRequests)
	}
	if len(reqs) == 0 {
		return c.JSON([]dto.User{})
	}
	ids := make([]int64, len(reqs))
	for i, r := range reqs {
		ids[i] = r.FriendId // sender id
	}

	users, err := e.user.GetUsersList(c.UserContext(), ids)
	if err != nil {
		return helper.HttpDbError(err, ErrUnableToGetUser)
	}
	discs, err := e.disc.GetDiscriminatorsByUserIDs(c.UserContext(), ids)
	if err != nil {
		return helper.HttpDbError(err, ErrUnableToGetDiscriminator)
	}
	return c.JSON(usersWithDiscriminators(users, discs))
}

// AcceptFriendRequest
//
//	@Summary	Accept a friend request
//	@Accept		json
//	@Produce	json
//	@Tags		User
//	@Param		request	body		FriendRequestAction	true	"Accept"
//	@Success	200		{string}	string				"ok"
//	@failure	400		{string}	string				"Bad request"
//	@failure	500		{string}	string				"Internal server error"
//	@Router		/user/me/friends/requests [post]
func (e *entity) AcceptFriendRequest(c *fiber.Ctx) error {
	var req FriendRequestAction
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToParseRequestBody)
	}
	if err := req.Validate(); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	me, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}
	if req.UserId == me.Id {
		return fiber.NewError(fiber.StatusBadRequest, ErrBadRequest)
	}

	if err := e.fr.AddFriend(c.UserContext(), me.Id, req.UserId); err != nil {
		return helper.HttpDbError(err, ErrUnableToAcceptFriendRequest)
	}
	// Remove friend request entry
	if err := e.fr.RemoveFriendRequest(c.UserContext(), me.Id, req.UserId); err != nil {
		return helper.HttpDbError(err, ErrUnableToAcceptFriendRequest)
	}
	// Emit WS events for both users (best-effort)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()
		// Notify me about new friend (the requester)
		if u, err := e.user.GetUserById(ctx, req.UserId); err == nil {
			if d, derr := e.disc.GetDiscriminatorByUserId(ctx, req.UserId); derr == nil {
				_ = e.mqt.SendUserUpdate(me.Id, &mqmsg.FriendAdded{Friend: mqmsg.UserBrief{Id: u.Id, Name: u.Name, Discriminator: d.Discriminator, Avatar: u.Avatar}})
			}
		}
		// Notify requester about me
		if u, err := e.user.GetUserById(ctx, me.Id); err == nil {
			if d, derr := e.disc.GetDiscriminatorByUserId(ctx, me.Id); derr == nil {
				_ = e.mqt.SendUserUpdate(req.UserId, &mqmsg.FriendAdded{Friend: mqmsg.UserBrief{Id: u.Id, Name: u.Name, Discriminator: d.Discriminator, Avatar: u.Avatar}})
			}
		}
	}()

	return c.SendStatus(fiber.StatusOK)
}

// DeclineFriendRequest
//
//	@Summary	Decline a friend request
//	@Accept		json
//	@Produce	json
//	@Tags		User
//	@Param		request	body		FriendRequestAction	true	"Decline"
//	@Success	200		{string}	string				"ok"
//	@failure	400		{string}	string				"Bad request"
//	@failure	500		{string}	string				"Internal server error"
//	@Router		/user/me/friends/requests [delete]
func (e *entity) DeclineFriendRequest(c *fiber.Ctx) error {
	var req FriendRequestAction
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToParseRequestBody)
	}
	if err := req.Validate(); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	me, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}

	if err := e.fr.RemoveFriendRequest(c.UserContext(), me.Id, req.UserId); err != nil {
		return helper.HttpDbError(err, ErrUnableToDeclineFriendRequest)
	}
	return c.SendStatus(fiber.StatusOK)
}
