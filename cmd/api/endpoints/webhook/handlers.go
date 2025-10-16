package webhook

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// StorageEvents
//
//	@Summary	Storage event
//	@Produce	json
//	@Tags		Webhook
//	@Param		request	body		S3Event	true	"S3 event"
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
		return errors.New(ErrUnableToParseRequestBody)
	}

	if len(event.Records) == 0 {
		return errors.New(ErrNoEventsProvided)
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
	return nil
}

func (e *entity) putAttachment(ctx context.Context, event *S3Event) error {
	objectId, channelId, err := extractAttachmentID(&event.Records[0].S3)
	if err != nil {
		return errors.New(ErrUnableToExtractID)
	}
	at, err := e.at.GetAttachment(ctx, objectId, channelId)
	if err != nil {
		cleanupErr := e.storage.RemoveAttachment(context.Background(), event.Key)
		if cleanupErr != nil {
			e.log.Error("failed to get attachment and also failed to cleanup S3 object", slog.Int64("objectId", objectId), slog.Int64("channelId", channelId), slog.String("get_error", err.Error()), slog.String("cleanup_error", cleanupErr.Error()))
		} else {
			e.log.Error("unable to get attachment, S3 object cleaned up", slog.Int64("objectId", objectId), slog.Int64("channelId", channelId), slog.String("error", err.Error()))
		}
		return errors.New(ErrAttachmentNotFount)
	}

	if at.FileSize != event.Records[0].S3.Object.Size {
		cleanupErr := e.storage.RemoveAttachment(context.Background(), event.Key)
		if cleanupErr != nil {
			e.log.Error("incorrect filesize and also failed to cleanup S3 object", slog.Int64("objectId", objectId), slog.Int64("channelId", channelId), slog.Int64("expected", at.FileSize), slog.Int64("actual", event.Records[0].S3.Object.Size), slog.String("cleanup_error", cleanupErr.Error()))
		} else {
			e.log.Error("incorrect filesize, S3 object cleaned up", slog.Int64("objectId", objectId), slog.Int64("channelId", channelId), slog.Int64("expected", at.FileSize), slog.Int64("actual", event.Records[0].S3.Object.Size))
		}
		return errors.New(ErrIncorrectFileSize)
	}

	if strings.HasSuffix(event.Key, "/preview.webp") {
		err = e.at.DoneAttachment(ctx, objectId, channelId, at.ContentType, at.URL, &event.Key, at.Height, at.Width)
	} else {
		err = e.at.DoneAttachment(ctx, objectId, channelId, event.Records[0].S3.Object.ContentType, &event.Key, nil, nil, nil)
	}
	if err != nil {
		cleanupErr := e.storage.RemoveAttachment(context.Background(), event.Key)
		if cleanupErr != nil {
			e.log.Error("failed to mark attachment done and also failed to cleanup S3 object", slog.Int64("objectId", objectId), slog.Int64("channelId", channelId), slog.String("done_error", err.Error()), slog.String("cleanup_error", cleanupErr.Error()))
		} else {
			e.log.Error("unable to mark attachment done, S3 object cleaned up", slog.Int64("objectId", objectId), slog.Int64("channelId", channelId), slog.String("error", err.Error()))
		}
		return errors.New(ErrUnableToDoneAttachment)
	}
	return nil
}

func (e *entity) deleteAttachment(ctx context.Context, event *S3Event) error {
	objectId, channelId, err := extractAttachmentID(&event.Records[0].S3)
	if err != nil {
		return errors.New(ErrUnableToExtractID)
	}
	err = e.at.RemoveAttachment(ctx, objectId, channelId)
	if err != nil {
		e.log.Error("unable to remove attachment from database", slog.Int64("objectId", objectId), slog.Int64("channelId", channelId), slog.String("error", err.Error()))
		return errors.New(ErrUnableToRemoveAttachment)
	}
	return nil
}
