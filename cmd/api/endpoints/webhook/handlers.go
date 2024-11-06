package webhook

import (
	"context"
	"log/slog"

	"github.com/gofiber/fiber/v2"
)

// StorageEvents
//
//	@Summary	Storage event
//	@Produce	json
//	@Tags		Webhook
//	@Success	200		{string}	string	"Ok"
//	@failure	400		{string}	string	"Incorrect ID"
//	@failure	404		{string}	string	"User not found"
//	@failure	500		{string}	string	"Something bad happened"
//	@Router		/webhook/storage/events [post]
func (e *entity) StorageEvents(c *fiber.Ctx) error {
	var event S3Event
	err := c.BodyParser(&event)
	if err != nil {
		e.log.Error("unable to parse s3 event message", slog.String("error", err.Error()))
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToParseRequestBody)
	}

	if len(event.Records) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, ErrNoEventsProvided)
	}

	switch event.EventName {
	case S3EventPut:
		switch event.Records[0].S3.Bucket.Name {
		case "media":
			err = e.putAttachment(c.UserContext(), &event)
			if err != nil {
				return err
			}
		}
	case S3EventDelete:
		switch event.Records[0].S3.Bucket.Name {
		case "media":
			err = e.deleteAttachment(c.UserContext(), &event)
			if err != nil {
				return err
			}
		}
	}
	return c.SendStatus(fiber.StatusOK)
}

func (e *entity) putAttachment(ctx context.Context, event *S3Event) *fiber.Error {
	objectId, channelId, err := extractAttachmentID(&event.Records[0].S3)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToExtractID)
	}
	at, err := e.at.GetAttachment(ctx, objectId, channelId)
	if err != nil {
		err := e.storage.RemoveAttachment(context.Background(), event.Key, event.Records[0].S3.Bucket.Name)
		if err != nil {
			e.log.Error("unable to get and remove attachment", slog.Int64("objectId", objectId), slog.Int64("channelId", channelId))
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		e.log.Error("unable to get attachment", slog.Int64("objectId", objectId), slog.Int64("channelId", channelId))
		return fiber.NewError(fiber.StatusNotFound, ErrAttachmentNotFount)
	}
	if at.FileSize != event.Records[0].S3.Object.Size {
		err := e.storage.RemoveAttachment(context.Background(), event.Key, event.Records[0].S3.Bucket.Name)
		if err != nil {
			e.log.Error("incorrect filesize, unable to remove object", slog.String("error", err.Error()), slog.Int64("objectId", objectId), slog.Int64("channelId", channelId), slog.Int64("expected", at.FileSize), slog.Int64("actual", event.Records[0].S3.Object.Size))
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		e.log.Error("incorrect filesize", slog.Int64("objectId", objectId), slog.Int64("channelId", channelId), slog.Int64("expected", at.FileSize), slog.Int64("actual", event.Records[0].S3.Object.Size))
		return fiber.NewError(fiber.StatusInternalServerError, ErrIncorrectFileSize)
	}
	err = e.at.DoneAttachment(ctx, objectId, channelId, event.Records[0].S3.Object.ContentType)
	if err != nil {
		err := e.storage.RemoveAttachment(context.Background(), event.Key, event.Records[0].S3.Bucket.Name)
		if err != nil {
			e.log.Error("incorrect filesize, unable to remove object", slog.String("error", err.Error()), slog.Int64("objectId", objectId), slog.Int64("channelId", channelId), slog.Int64("expected", at.FileSize), slog.Int64("actual", event.Records[0].S3.Object.Size))
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		e.log.Error("unable to complete attachment", slog.String("error", err.Error()))
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToDoneAttachment)
	}
	return nil
}

func (e *entity) deleteAttachment(ctx context.Context, event *S3Event) *fiber.Error {
	objectId, channelId, err := extractAttachmentID(&event.Records[0].S3)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToExtractID)
	}
	err = e.at.RemoveAttachment(ctx, objectId, channelId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToRemoveAttachemnt)
	}
	return nil
}
